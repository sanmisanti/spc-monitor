package api

import (
	"encoding/json"
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

	wg.Wait()

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

	// Check HTTP
	httpCheck := monitors.CheckHTTP(
		"https://saltacompra.gob.ar/",
		"http-check",
		"Sitio web accesible",
	)
	system.Checks = append(system.Checks, httpCheck)

	// Check servicio de mails
	mailConfig := monitors.MailCheckConfig{
		Host:                     h.config.Database.Host,
		Port:                     h.config.Database.Port,
		User:                     h.config.Database.User,
		Password:                 h.config.Database.Password,
		Database:                 h.config.Database.Database,
		MaxMinutesWithoutSending: h.config.Monitors.MailMaxMinutesWithoutSending,
		MaxFailedCount:           h.config.Monitors.MailMaxFailedCount,
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

	// Check HTTP
	httpCheck := monitors.CheckHTTP(
		"https://preproduccion.saltacompra.gob.ar/",
		"http-check",
		"Sitio web accesible",
	)
	system.Checks = append(system.Checks, httpCheck)

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
