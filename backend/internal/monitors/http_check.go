package monitors

import (
	"net/http"
	"time"

	"github.com/saltacompra/monitor/internal/models"
)

// CheckHTTP verifica si una URL responde correctamente
func CheckHTTP(url string, checkID string, checkName string) models.Check {
	check := models.Check{
		ID:        checkID,
		Type:      "http",
		Name:      checkName,
		LastCheck: time.Now(),
	}

	start := time.Now()
	resp, err := http.Get(url)
	elapsed := time.Since(start).Milliseconds()

	check.ResponseTime = elapsed

	if err != nil {
		check.Status = "error"
		check.Message = "No se pudo conectar: " + err.Error()
		return check
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		check.Status = "ok"
		check.Message = "Sitio accesible (HTTP " + resp.Status + ")"
	} else if resp.StatusCode >= 500 {
		check.Status = "error"
		check.Message = "Error del servidor (HTTP " + resp.Status + ")"
	} else {
		check.Status = "warning"
		check.Message = "Respuesta inesperada (HTTP " + resp.Status + ")"
	}

	return check
}
