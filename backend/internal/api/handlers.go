package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/saltacompra/monitor/internal/cache"
	"github.com/saltacompra/monitor/internal/config"
	"github.com/saltacompra/monitor/internal/models"
	"github.com/saltacompra/monitor/internal/monitors"
	"github.com/saltacompra/monitor/internal/sse"
)

// Handler maneja las peticiones HTTP
type Handler struct {
	config      config.Config
	cache       *cache.SystemCache
	broadcaster *sse.Broadcaster
}

// NewHandler crea un nuevo handler
func NewHandler(cfg config.Config, cache *cache.SystemCache, broadcaster *sse.Broadcaster) *Handler {
	return &Handler{
		config:      cfg,
		cache:       cache,
		broadcaster: broadcaster,
	}
}

// GetSystems devuelve el estado de todos los sistemas desde el cache
func (h *Handler) GetSystems(w http.ResponseWriter, r *http.Request) {
	systems := h.cache.GetAll()

	response := map[string]interface{}{
		"systems": systems,
		"cached":  true,
		"count":   len(systems),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetEvents maneja la conexión SSE para enviar updates en tiempo real
func (h *Handler) GetEvents(w http.ResponseWriter, r *http.Request) {
	// Headers para SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Generar ID de cliente único
	clientID := fmt.Sprintf("client_%d", time.Now().UnixNano())

	// Registrar cliente
	client := h.broadcaster.Register(clientID)
	defer h.broadcaster.Unregister(clientID)

	// Enviar mensaje de bienvenida
	fmt.Fprintf(w, "event: connected\ndata: {\"client_id\": \"%s\"}\n\n", clientID)
	w.(http.Flusher).Flush()

	// Escuchar mensajes del broadcaster
	for {
		select {
		case message, ok := <-client.Channel:
			if !ok {
				return
			}
			fmt.Fprint(w, message)
			w.(http.Flusher).Flush()
		case <-r.Context().Done():
			return
		}
	}
}

// RefreshAllSystems dispara la ejecución de todos los checks (async)
func (h *Handler) RefreshAllSystems(w http.ResponseWriter, r *http.Request) {
	log.Println("[API] Refresh manual solicitado para todos los sistemas")

	// Ejecutar checks en background con broadcasts progresivos
	go func() {
		h.checkAllSystemsWithProgressiveBroadcast()
	}()

	// Responder inmediatamente
	w.WriteHeader(http.StatusAccepted)
	response := map[string]interface{}{
		"message": "Refresh iniciado",
		"status":  "processing",
	}
	json.NewEncoder(w).Encode(response)
}

// RefreshSystem dispara la ejecución de check de un sistema específico (async)
func (h *Handler) RefreshSystem(w http.ResponseWriter, r *http.Request) {
	// Extraer ID del sistema de la URL
	path := r.URL.Path
	parts := strings.Split(strings.TrimPrefix(path, "/api/systems/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, "ID de sistema no especificado", http.StatusBadRequest)
		return
	}
	systemID := parts[0]

	log.Printf("[API] Refresh manual solicitado para sistema: %s", systemID)

	// Ejecutar check en background
	go func() {
		system := h.checkSystemByID(systemID)
		if system.ID != "" {
			h.cache.Set(system.ID, system)
			h.broadcaster.BroadcastSystem(system)
			log.Printf("[API] Refresh manual completado: %s", systemID)
		} else {
			log.Printf("[API] Sistema no encontrado: %s", systemID)
		}
	}()

	// Responder inmediatamente
	w.WriteHeader(http.StatusAccepted)
	response := map[string]interface{}{
		"message":   "Refresh iniciado",
		"system_id": systemID,
		"status":    "processing",
	}
	json.NewEncoder(w).Encode(response)
}

// CheckAllSystems ejecuta todos los checks en paralelo y retorna los sistemas
// Este método se usa desde el worker (sin broadcasts)
func (h *Handler) CheckAllSystems() []models.System {
	var wg sync.WaitGroup
	systemsChan := make(chan models.System, 5)

	// Sistema 1: SaltaCompra Producción
	wg.Add(1)
	go func() {
		defer wg.Done()
		systemsChan <- h.checkSaltaCompraProd()
	}()

	// Sistema 2: SaltaCompra Preproducción
	wg.Add(1)
	go func() {
		defer wg.Done()
		systemsChan <- h.checkSaltaCompraPreProd()
	}()

	// Sistema 3: Infraestructura Compartida
	wg.Add(1)
	go func() {
		defer wg.Done()
		systemsChan <- h.checkInfrastructure()
	}()

	// Sistema 4: Google Sheets - Kairos
	wg.Add(1)
	go func() {
		defer wg.Done()
		systemsChan <- h.checkGoogleSheetsKairos()
	}()

	// Sistema 5: App.SaltaCompra
	wg.Add(1)
	go func() {
		defer wg.Done()
		systemsChan <- h.checkAppSaltaCompra()
	}()

	// Esperar a que terminen y cerrar canal
	go func() {
		wg.Wait()
		close(systemsChan)
	}()

	// Recolectar resultados
	systems := []models.System{}
	for system := range systemsChan {
		systems = append(systems, system)
	}

	return systems
}

// checkAllSystemsWithProgressiveBroadcast ejecuta checks en paralelo y envía SSE conforme completan
func (h *Handler) checkAllSystemsWithProgressiveBroadcast() {
	var wg sync.WaitGroup

	// Sistema 1: SaltaCompra Producción
	wg.Add(1)
	go func() {
		defer wg.Done()
		system := h.checkSaltaCompraProd()
		h.cache.Set(system.ID, system)
		h.broadcaster.BroadcastSystem(system)
		log.Printf("[API] Sistema actualizado: %s", system.Name)
	}()

	// Sistema 2: SaltaCompra Preproducción
	wg.Add(1)
	go func() {
		defer wg.Done()
		system := h.checkSaltaCompraPreProd()
		h.cache.Set(system.ID, system)
		h.broadcaster.BroadcastSystem(system)
		log.Printf("[API] Sistema actualizado: %s", system.Name)
	}()

	// Sistema 3: Infraestructura Compartida
	wg.Add(1)
	go func() {
		defer wg.Done()
		system := h.checkInfrastructure()
		h.cache.Set(system.ID, system)
		h.broadcaster.BroadcastSystem(system)
		log.Printf("[API] Sistema actualizado: %s", system.Name)
	}()

	// Sistema 4: Google Sheets - Kairos
	wg.Add(1)
	go func() {
		defer wg.Done()
		system := h.checkGoogleSheetsKairos()
		h.cache.Set(system.ID, system)
		h.broadcaster.BroadcastSystem(system)
		log.Printf("[API] Sistema actualizado: %s", system.Name)
	}()

	// Sistema 5: App.SaltaCompra
	wg.Add(1)
	go func() {
		defer wg.Done()
		system := h.checkAppSaltaCompra()
		h.cache.Set(system.ID, system)
		h.broadcaster.BroadcastSystem(system)
		log.Printf("[API] Sistema actualizado: %s", system.Name)
	}()

	// Esperar a que todos terminen
	wg.Wait()

	h.broadcaster.BroadcastCheckComplete()
	log.Println("[API] Todos los checks completados")
}

// checkSystemByID ejecuta el check de un sistema específico por ID
func (h *Handler) checkSystemByID(id string) models.System {
	switch id {
	case "saltacompra-prod":
		return h.checkSaltaCompraProd()
	case "saltacompra-preprod":
		return h.checkSaltaCompraPreProd()
	case "infrastructure":
		return h.checkInfrastructure()
	case "google-sheets-kairos":
		return h.checkGoogleSheetsKairos()
	case "app-saltacompra":
		return h.checkAppSaltaCompra()
	default:
		log.Printf("[API] Sistema no encontrado: %s", id)
		return models.System{}
	}
}

// checkSaltaCompraProd verifica el estado de SaltaCompra Producción
func (h *Handler) checkSaltaCompraProd() models.System {
	system := models.System{
		ID:          "saltacompra-prod",
		Name:        "SaltaCompra Producción",
		Type:        "web",
		Environment: "prod",
		Status:      "unknown",
		Checks:      []models.Check{},
	}

	// Check HTTP con validaciones completas
	httpCheck := monitors.CheckHTTP(monitors.HTTPCheckConfig{
		URL:              h.config.SaltaCompra.ProdURL,
		CheckID:          "http-check",
		CheckName:        "Sitio web accesible",
		ExpectedContent:  []string{h.config.SaltaCompra.ProdExpectedContent},
		ValidateSSL:      true,
		SSLWarningDays:   h.config.Monitors.SSLWarningDays,
		TimeoutWarningMs: h.config.Monitors.HTTPTimeoutWarningMs,
		TimeoutErrorMs:   h.config.Monitors.HTTPTimeoutErrorMs,
		TimeoutSeconds:   h.config.Monitors.HTTPTimeoutSeconds,
	})
	system.Checks = append(system.Checks, httpCheck)

	// Check servicio de mails
	mailConfig := monitors.MailCheckConfig{
		Host:                      h.config.DatabaseProd.Host,
		Port:                      h.config.DatabaseProd.Port,
		User:                      h.config.DatabaseProd.User,
		Password:                  h.config.DatabaseProd.Password,
		Database:                  h.config.DatabaseProd.Database,
		MaxMinutesWithoutSent:     h.config.Monitors.MailMaxMinutesWithoutSent,
		DailyWarningFailedPercent: h.config.Monitors.MailDailyWarningFailedPercent,
		DailyErrorFailedPercent:   h.config.Monitors.MailDailyErrorFailedPercent,
	}
	mailCheck := monitors.CheckMailService(mailConfig, "mail-service", "Servicio de correos")
	system.Checks = append(system.Checks, mailCheck)

	// Determinar estado general del sistema
	system.Status = determineSystemStatus(system.Checks)
	if len(system.Checks) > 0 {
		system.LastCheck = system.Checks[0].LastCheck
	}

	return system
}

// checkSaltaCompraPreProd verifica el estado de SaltaCompra Preproducción
func (h *Handler) checkSaltaCompraPreProd() models.System {
	system := models.System{
		ID:          "saltacompra-preprod",
		Name:        "SaltaCompra Preproducción",
		Type:        "web",
		Environment: "preprod",
		Status:      "unknown",
		Checks:      []models.Check{},
	}

	// Check HTTP con validaciones completas
	httpCheck := monitors.CheckHTTP(monitors.HTTPCheckConfig{
		URL:              h.config.SaltaCompra.PreProdURL,
		CheckID:          "http-check",
		CheckName:        "Sitio web accesible",
		ExpectedContent:  []string{h.config.SaltaCompra.PreProdExpectedContent},
		ValidateSSL:      true,
		SSLWarningDays:   h.config.Monitors.SSLWarningDays,
		TimeoutWarningMs: h.config.Monitors.HTTPTimeoutWarningMs,
		TimeoutErrorMs:   h.config.Monitors.HTTPTimeoutErrorMs,
		TimeoutSeconds:   h.config.Monitors.HTTPTimeoutSeconds,
	})
	system.Checks = append(system.Checks, httpCheck)

	// Check servicio de mails
	mailConfig := monitors.MailCheckConfig{
		Host:                      h.config.DatabasePreProd.Host,
		Port:                      h.config.DatabasePreProd.Port,
		User:                      h.config.DatabasePreProd.User,
		Password:                  h.config.DatabasePreProd.Password,
		Database:                  h.config.DatabasePreProd.Database,
		MaxMinutesWithoutSent:     h.config.Monitors.MailMaxMinutesWithoutSent,
		DailyWarningFailedPercent: h.config.Monitors.MailDailyWarningFailedPercent,
		DailyErrorFailedPercent:   h.config.Monitors.MailDailyErrorFailedPercent,
	}
	mailCheck := monitors.CheckMailService(mailConfig, "mail-service", "Servicio de correos")
	system.Checks = append(system.Checks, mailCheck)

	// Determinar estado general
	system.Status = determineSystemStatus(system.Checks)
	if len(system.Checks) > 0 {
		system.LastCheck = system.Checks[0].LastCheck
	}

	return system
}

// checkInfrastructure verifica el estado de la infraestructura compartida
func (h *Handler) checkInfrastructure() models.System {
	system := models.System{
		ID:          "infrastructure",
		Name:        "Infraestructura Compartida",
		Type:        "infrastructure",
		Environment: "shared",
		Status:      "unknown",
		Checks:      []models.Check{},
	}

	// Check RDAP para expiración de dominio
	rdapCheck := monitors.CheckRDAPDomain(monitors.RDAPCheckConfig{
		Domain:      h.config.Infrastructure.Domain,
		RDAPBaseURL: h.config.Infrastructure.RDAPBaseURL,
		CheckID:     "domain-expiry",
		CheckName:   "Expiración de dominio",
		WarningDays: h.config.Monitors.DomainWarningDays,
		ErrorDays:   h.config.Monitors.DomainErrorDays,
	})
	system.Checks = append(system.Checks, rdapCheck)

	// Determinar estado general
	system.Status = determineSystemStatus(system.Checks)
	if len(system.Checks) > 0 {
		system.LastCheck = system.Checks[0].LastCheck
	}

	return system
}

// checkGoogleSheetsKairos verifica el estado del sistema Google Sheets Kairos
func (h *Handler) checkGoogleSheetsKairos() models.System {
	system := models.System{
		ID:          "google-sheets-kairos",
		Name:        "Google Sheets - Kairos Actualizaciones",
		Type:        "google-script",
		Environment: "prod",
		Status:      "unknown",
		Checks:      []models.Check{},
	}

	// Check de actualización diaria
	kairosCheck := monitors.CheckGoogleSheetsKairos(monitors.GoogleSheetsCheckConfig{
		SpreadsheetID:   h.config.GoogleSheets.SpreadsheetID,
		SheetName:       h.config.GoogleSheets.SheetName,
		AuthMethod:      h.config.GoogleSheets.AuthMethod,
		CredentialsFile: h.config.GoogleSheets.CredentialsFile,
		APIKey:          h.config.GoogleSheets.APIKey,
		TimestampColumn: h.config.GoogleSheets.TimestampColumn,
		FilenameColumn:  h.config.GoogleSheets.FilenameColumn,
		WarningDays:     h.config.GoogleSheets.WarningDays,
		ErrorDays:       h.config.GoogleSheets.ErrorDays,
		CheckID:         "kairos-daily-update",
		CheckName:       "Actualización diaria Kairos",
	})
	system.Checks = append(system.Checks, kairosCheck)

	// Determinar estado general
	system.Status = determineSystemStatus(system.Checks)
	if len(system.Checks) > 0 {
		system.LastCheck = system.Checks[0].LastCheck
	}

	return system
}

// checkAppSaltaCompra verifica el estado de App.SaltaCompra
func (h *Handler) checkAppSaltaCompra() models.System {
	system := models.System{
		ID:          "app-saltacompra",
		Name:        "App.SaltaCompra",
		Type:        "web",
		Environment: "prod",
		Status:      "unknown",
		Checks:      []models.Check{},
	}

	// Check HTTP (público)
	httpCheck := monitors.CheckHTTP(monitors.HTTPCheckConfig{
		URL:                  h.config.AppSaltaCompra.URL,
		CheckID:              "http-check",
		CheckName:            "Sitio web accesible",
		ExpectedContent:      []string{h.config.AppSaltaCompra.ExpectedContent},
		ValidateSSL:          true,
		SkipSSLVerification:  h.config.AppSaltaCompra.SkipSSLVerification,
		SSLWarningDays:       h.config.Monitors.SSLWarningDays,
		TimeoutWarningMs:     h.config.Monitors.HTTPTimeoutWarningMs,
		TimeoutErrorMs:       h.config.Monitors.HTTPTimeoutErrorMs,
		TimeoutSeconds:       h.config.Monitors.HTTPTimeoutSeconds,
	})
	system.Checks = append(system.Checks, httpCheck)

	// Check PostgreSQL (requiere VPN)
	pgCheck := monitors.CheckPostgreSQL(monitors.PostgreSQLCheckConfig{
		Host:         h.config.PostgreSQLAppSPC.Host,
		Port:         h.config.PostgreSQLAppSPC.Port,
		User:         h.config.PostgreSQLAppSPC.User,
		Password:     h.config.PostgreSQLAppSPC.Password,
		Database:     h.config.PostgreSQLAppSPC.Database,
		CheckID:      "postgresql-check",
		CheckName:    "Base de datos PostgreSQL",
		VPNCheckHost: h.config.VPNCheck.Host,
		VPNTimeoutMs: h.config.VPNCheck.TimeoutMs,
	})
	system.Checks = append(system.Checks, pgCheck)

	// Determinar estado general
	system.Status = determineSystemStatus(system.Checks)
	if len(system.Checks) > 0 {
		system.LastCheck = system.Checks[0].LastCheck
	}

	return system
}

// determineSystemStatus determina el estado general basado en los checks
func determineSystemStatus(checks []models.Check) string {
	if len(checks) == 0 {
		return "unknown"
	}

	hasError := false
	hasWarning := false

	for _, check := range checks {
		if check.Status == "error" {
			hasError = true
		} else if check.Status == "warning" {
			hasWarning = true
		}
	}

	if hasError {
		return "error"
	} else if hasWarning {
		return "warning"
	}
	return "online"
}
