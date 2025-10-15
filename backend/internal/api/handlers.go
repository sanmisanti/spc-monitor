package api

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/saltacompra/monitor/internal/config"
	"github.com/saltacompra/monitor/internal/models"
	"github.com/saltacompra/monitor/internal/monitors"
)

// Handler maneja las peticiones HTTP
type Handler struct {
	config config.Config
}

// NewHandler crea un nuevo handler
func NewHandler(cfg config.Config) *Handler {
	return &Handler{config: cfg}
}

// GetSystems devuelve el estado de todos los sistemas monitoreados
func (h *Handler) GetSystems(w http.ResponseWriter, r *http.Request) {
	log.Printf("[API] Ejecutando checks de sistemas...")

	// Ejecutar todos los checks en paralelo
	var wg sync.WaitGroup
	systems := []models.System{}

	// Sistema 1: SaltaCompra Producción
	wg.Add(1)
	go func() {
		defer wg.Done()
		system := h.checkSaltaCompraProd()
		systems = append(systems, system)
	}()

	// Sistema 2: SaltaCompra Preproducción
	wg.Add(1)
	go func() {
		defer wg.Done()
		system := h.checkSaltaCompraPreProd()
		systems = append(systems, system)
	}()

	// Sistema 3: Infraestructura Compartida
	wg.Add(1)
	go func() {
		defer wg.Done()
		system := h.checkInfrastructure()
		systems = append(systems, system)
	}()

	// Sistema 4: Google Sheets - Kairos
	wg.Add(1)
	go func() {
		defer wg.Done()
		system := h.checkGoogleSheetsKairos()
		systems = append(systems, system)
	}()

	// Sistema 5: App.SaltaCompra
	wg.Add(1)
	go func() {
		defer wg.Done()
		system := h.checkAppSaltaCompra()
		systems = append(systems, system)
	}()

	wg.Wait()

	log.Printf("[API] Checks completados. Sistemas verificados: %d", len(systems))

	// Responder con JSON
	response := map[string]interface{}{
		"systems": systems,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
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
		Host:                     h.config.DatabaseProd.Host,
		Port:                     h.config.DatabaseProd.Port,
		User:                     h.config.DatabaseProd.User,
		Password:                 h.config.DatabaseProd.Password,
		Database:                 h.config.DatabaseProd.Database,
		MaxMinutesWithoutSending: h.config.Monitors.MailMaxMinutesWithoutSending,
		MaxFailedCount:           h.config.Monitors.MailMaxFailedCount,
		MaxUnsentCount:           h.config.Monitors.MailMaxUnsentCount,
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
		Host:                     h.config.DatabasePreProd.Host,
		Port:                     h.config.DatabasePreProd.Port,
		User:                     h.config.DatabasePreProd.User,
		Password:                 h.config.DatabasePreProd.Password,
		Database:                 h.config.DatabasePreProd.Database,
		MaxMinutesWithoutSending: h.config.Monitors.MailMaxMinutesWithoutSending,
		MaxFailedCount:           h.config.Monitors.MailMaxFailedCount,
		MaxUnsentCount:           h.config.Monitors.MailMaxUnsentCount,
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
		URL:              h.config.AppSaltaCompra.URL,
		CheckID:          "http-check",
		CheckName:        "Sitio web accesible",
		ExpectedContent:  []string{h.config.AppSaltaCompra.ExpectedContent},
		ValidateSSL:      true,
		SSLWarningDays:   h.config.Monitors.SSLWarningDays,
		TimeoutWarningMs: h.config.Monitors.HTTPTimeoutWarningMs,
		TimeoutErrorMs:   h.config.Monitors.HTTPTimeoutErrorMs,
		TimeoutSeconds:   h.config.Monitors.HTTPTimeoutSeconds,
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
