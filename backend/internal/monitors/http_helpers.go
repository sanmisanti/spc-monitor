package monitors

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// checkContentPresence verifica que el HTML contenga todos los textos esperados
func checkContentPresence(body string, expectedTexts []string) (bool, string) {
	if len(expectedTexts) == 0 {
		return true, ""
	}

	for _, text := range expectedTexts {
		if !strings.Contains(body, text) {
			return false, "Contenido esperado no encontrado: " + text
		}
	}

	return true, "Contenido verificado correctamente"
}

// validateSSLCertificate verifica el certificado SSL de una URL
func validateSSLCertificate(urlStr string, warningDays int) (status string, message string, daysRemaining int) {
	// Hacer petición HTTPS para obtener certificado
	resp, err := http.Get(urlStr)
	if err != nil {
		return "error", "No se pudo verificar SSL: " + err.Error(), 0
	}
	defer resp.Body.Close()

	// Verificar que la conexión usó TLS
	if resp.TLS == nil {
		return "warning", "Conexión no usa TLS/SSL", 0
	}

	// Obtener el primer certificado (del servidor)
	if len(resp.TLS.PeerCertificates) == 0 {
		return "warning", "No se encontraron certificados", 0
	}

	cert := resp.TLS.PeerCertificates[0]

	// Calcular días hasta expiración
	now := time.Now()
	if now.After(cert.NotAfter) {
		return "error", "Certificado SSL vencido", 0
	}

	daysUntilExpiry := int(cert.NotAfter.Sub(now).Hours() / 24)

	// Verificar si está por vencer
	if daysUntilExpiry <= warningDays {
		return "warning",
			fmt.Sprintf("Certificado SSL expira pronto (en %d días)", daysUntilExpiry),
			daysUntilExpiry
	}

	return "ok", "Certificado SSL válido", daysUntilExpiry
}

// evaluateResponseTime evalúa el tiempo de respuesta según umbrales
func evaluateResponseTime(elapsedMs int64, warningThresholdMs int64, errorThresholdMs int64) string {
	if errorThresholdMs > 0 && elapsedMs >= errorThresholdMs {
		return "error"
	}
	if warningThresholdMs > 0 && elapsedMs >= warningThresholdMs {
		return "warning"
	}
	return "ok"
}

// readResponseBody lee el body de una respuesta HTTP de forma segura
func readResponseBody(resp *http.Response) (string, error) {
	// Limitar lectura a 1MB para evitar problemas de memoria
	limitedReader := io.LimitReader(resp.Body, 1024*1024)
	bodyBytes, err := io.ReadAll(limitedReader)
	if err != nil {
		return "", err
	}
	return string(bodyBytes), nil
}

// getHTTPClient retorna un cliente HTTP configurado con timeouts y opciones SSL
func getHTTPClient(timeoutSeconds int, skipSSLVerify bool) *http.Client {
	return &http.Client{
		Timeout: time.Duration(timeoutSeconds) * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: skipSSLVerify, // Configurable según necesidad
			},
		},
	}
}
