package monitors

import (
	"fmt"
	"time"

	"github.com/saltacompra/monitor/internal/models"
)

// HTTPCheckConfig contiene la configuración para verificaciones HTTP
type HTTPCheckConfig struct {
	URL                  string
	CheckID              string
	CheckName            string
	ExpectedContent      []string // Textos que deben estar presentes en el HTML
	ValidateSSL          bool     // Si debe validar certificado SSL
	SkipSSLVerification  bool     // Saltar verificación SSL (para certificados autofirmados)
	SSLWarningDays       int      // Días antes de expiración para warning
	TimeoutWarningMs     int64    // Umbral de ms para warning
	TimeoutErrorMs       int64    // Umbral de ms para error
	TimeoutSeconds       int      // Timeout de la petición HTTP
}

// CheckHTTP verifica si una URL responde correctamente con validaciones completas
func CheckHTTP(config HTTPCheckConfig) models.Check {
	check := models.Check{
		ID:        config.CheckID,
		Type:      "http",
		Name:      config.CheckName,
		LastCheck: time.Now(),
		Metadata:  make(map[string]interface{}),
	}

	// Cliente HTTP con timeout
	timeout := config.TimeoutSeconds
	if timeout == 0 {
		timeout = 30 // Default 30 segundos
	}
	client := getHTTPClient(timeout, config.SkipSSLVerification)

	// Realizar petición HTTP
	start := time.Now()
	resp, err := client.Get(config.URL)
	elapsed := time.Since(start).Milliseconds()
	check.ResponseTime = elapsed

	if err != nil {
		check.Status = "error"
		check.Message = "No se pudo conectar: " + err.Error()
		return check
	}
	defer resp.Body.Close()

	// Lista de problemas encontrados
	var issues []string
	worstStatus := "ok"

	// 1. Verificar código HTTP
	if resp.StatusCode >= 500 {
		issues = append(issues, fmt.Sprintf("Error del servidor (HTTP %d)", resp.StatusCode))
		worstStatus = "error"
	} else if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		issues = append(issues, fmt.Sprintf("Código HTTP inesperado: %d", resp.StatusCode))
		if worstStatus == "ok" {
			worstStatus = "warning"
		}
	}

	// 2. Verificar tiempo de respuesta
	timeStatus := evaluateResponseTime(elapsed, config.TimeoutWarningMs, config.TimeoutErrorMs)
	check.Metadata["response_time_status"] = timeStatus

	if timeStatus == "error" {
		issues = append(issues, fmt.Sprintf("Tiempo de respuesta muy alto: %dms", elapsed))
		worstStatus = "error"
	} else if timeStatus == "warning" {
		issues = append(issues, fmt.Sprintf("Tiempo de respuesta elevado: %dms", elapsed))
		if worstStatus == "ok" {
			worstStatus = "warning"
		}
	}

	// 3. Verificar contenido esperado
	if len(config.ExpectedContent) > 0 && resp.StatusCode >= 200 && resp.StatusCode < 300 {
		body, err := readResponseBody(resp)
		if err != nil {
			issues = append(issues, "Error al leer contenido: "+err.Error())
			worstStatus = "error"
		} else {
			contentOk, contentMsg := checkContentPresence(body, config.ExpectedContent)
			check.Metadata["content_validated"] = contentOk

			if !contentOk {
				issues = append(issues, contentMsg)
				worstStatus = "error"
			}
		}
	}

	// 4. Verificar SSL (solo si no se saltea la verificación)
	if config.ValidateSSL && !config.SkipSSLVerification {
		sslStatus, sslMsg, daysRemaining := validateSSLCertificate(config.URL, config.SSLWarningDays)
		check.Metadata["ssl_status"] = sslStatus
		check.Metadata["ssl_days_remaining"] = daysRemaining

		if sslStatus == "error" {
			issues = append(issues, sslMsg)
			worstStatus = "error"
		} else if sslStatus == "warning" {
			issues = append(issues, sslMsg)
			if worstStatus == "ok" {
				worstStatus = "warning"
			}
		}
	} else if config.SkipSSLVerification {
		// Si se saltea verificación SSL, marcar explícitamente
		check.Metadata["ssl_status"] = "skipped"
		check.Metadata["ssl_verification_skipped"] = true
	}

	// Construir mensaje final
	check.Status = worstStatus
	if len(issues) > 0 {
		check.Message = fmt.Sprintf("%s - Problemas: %v", resp.Status, issues)
	} else {
		check.Message = fmt.Sprintf("Sitio accesible y funcionando correctamente (HTTP %s, %dms)", resp.Status, elapsed)
	}

	return check
}
