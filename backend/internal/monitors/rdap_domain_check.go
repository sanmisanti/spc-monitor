package monitors

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/saltacompra/monitor/internal/models"
)

// RDAPCheckConfig contiene la configuración para verificaciones RDAP
type RDAPCheckConfig struct {
	Domain      string
	CheckID     string
	CheckName   string
	WarningDays int // Días antes de expiración para warning
	ErrorDays   int // Días antes de expiración para error
}

// RDAPResponse estructura simplificada de la respuesta RDAP
type RDAPResponse struct {
	Events []RDAPEvent `json:"events"`
	Status []string    `json:"status"`
}

// RDAPEvent representa un evento en la respuesta RDAP
type RDAPEvent struct {
	EventAction string `json:"eventAction"`
	EventDate   string `json:"eventDate"`
}

// CheckRDAPDomain verifica la fecha de expiración de un dominio vía RDAP
func CheckRDAPDomain(config RDAPCheckConfig) models.Check {
	check := models.Check{
		ID:        config.CheckID,
		Type:      "rdap",
		Name:      config.CheckName,
		LastCheck: time.Now(),
		Metadata:  make(map[string]interface{}),
	}

	// URL del servicio RDAP de NIC Argentina
	url := fmt.Sprintf("https://rdap.nic.ar/domain/%s", config.Domain)

	// Realizar petición HTTP
	start := time.Now()
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	elapsed := time.Since(start).Milliseconds()
	check.ResponseTime = elapsed

	if err != nil {
		check.Status = "error"
		check.Message = "No se pudo consultar RDAP: " + err.Error()
		return check
	}
	defer resp.Body.Close()

	// Verificar código HTTP
	if resp.StatusCode != 200 {
		check.Status = "error"
		check.Message = fmt.Sprintf("Error en respuesta RDAP (HTTP %d)", resp.StatusCode)
		return check
	}

	// Leer y parsear respuesta JSON
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		check.Status = "error"
		check.Message = "Error al leer respuesta RDAP: " + err.Error()
		return check
	}

	var rdapData RDAPResponse
	if err := json.Unmarshal(body, &rdapData); err != nil {
		check.Status = "error"
		check.Message = "Error al parsear JSON de RDAP: " + err.Error()
		return check
	}

	// Buscar fecha de expiración en los eventos
	var expirationDate time.Time
	found := false
	for _, event := range rdapData.Events {
		if event.EventAction == "expiration" {
			expirationDate, err = time.Parse(time.RFC3339, event.EventDate)
			if err != nil {
				check.Status = "error"
				check.Message = "Error al parsear fecha de expiración: " + err.Error()
				return check
			}
			found = true
			break
		}
	}

	if !found {
		check.Status = "error"
		check.Message = "No se encontró fecha de expiración en respuesta RDAP"
		return check
	}

	// Calcular días hasta expiración
	now := time.Now()
	daysRemaining := int(expirationDate.Sub(now).Hours() / 24)

	// Guardar metadata
	check.Metadata["expiration_date"] = expirationDate.Format("2006-01-02")
	check.Metadata["days_remaining"] = daysRemaining
	check.Metadata["domain_status"] = rdapData.Status

	// Determinar estado según umbrales
	if daysRemaining < 0 {
		check.Status = "error"
		check.Message = fmt.Sprintf("¡DOMINIO VENCIDO! Expiró hace %d días", -daysRemaining)
	} else if daysRemaining <= config.ErrorDays {
		check.Status = "error"
		check.Message = fmt.Sprintf("¡Dominio expira pronto! Quedan solo %d días (expira: %s)",
			daysRemaining, expirationDate.Format("02/01/2006"))
	} else if daysRemaining <= config.WarningDays {
		check.Status = "warning"
		check.Message = fmt.Sprintf("Dominio debe renovarse pronto. Quedan %d días (expira: %s)",
			daysRemaining, expirationDate.Format("02/01/2006"))
	} else {
		check.Status = "ok"
		check.Message = fmt.Sprintf("Dominio válido. Expira en %d días (%s)",
			daysRemaining, expirationDate.Format("02/01/2006"))
	}

	return check
}
