package monitors

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/microsoft/go-mssqldb"
	"github.com/saltacompra/monitor/internal/models"
)

// MailCheckConfig contiene la configuración para el check de correos
type MailCheckConfig struct {
	Host                         string
	Port                         int
	User                         string
	Password                     string
	Database                     string
	MaxMinutesWithoutSending     int // Umbral de minutos sin enviar antes de warning
	MaxFailedCount               int // Umbral de mails fallidos antes de error
	MaxUnsentCount               int // Umbral de mails unsent antes de warning (cola atascada)
}

// CheckMailService verifica el estado del servicio de mails en SQL Server
func CheckMailService(config MailCheckConfig, checkID string, checkName string) models.Check {
	check := models.Check{
		ID:        checkID,
		Type:      "database",
		Name:      checkName,
		LastCheck: time.Now(),
		Metadata:  make(map[string]interface{}),
	}

	// Construir connection string
	connString := fmt.Sprintf("server=%s;user id=%s;password=%s;port=%d;database=%s",
		config.Host, config.User, config.Password, config.Port, config.Database)

	start := time.Now()
	db, err := sql.Open("sqlserver", connString)
	if err != nil {
		check.Status = "error"
		check.Message = "Error al conectar a SQL Server: " + err.Error()
		return check
	}
	defer db.Close()

	// Test de conexión
	err = db.Ping()
	elapsed := time.Since(start).Milliseconds()
	check.ResponseTime = elapsed

	if err != nil {
		check.Status = "error"
		check.Message = "No se pudo conectar a la base de datos: " + err.Error()
		return check
	}

	// Query para obtener los últimos mails con su estado convertido a string
	query := `
		SELECT TOP 10
			mailitem_id,
			CASE sent_status
				WHEN 0 THEN 'unsent'
				WHEN 1 THEN 'sent'
				WHEN 3 THEN 'retrying'
				ELSE 'failed'
			END as sent_status,
			sent_date,
			last_mod_date
		FROM msdb.dbo.sysmail_mailitems
		ORDER BY mailitem_id DESC
	`

	rows, err := db.Query(query)
	if err != nil {
		check.Status = "error"
		check.Message = "Error al consultar sysmail_mailitems: " + err.Error()
		return check
	}
	defer rows.Close()

	// Procesar resultados para análisis
	var lastSentDate, lastModDate *time.Time
	var mailitemID int
	var sentStatus string
	statusCount := make(map[string]int) // Conteo dinámico de estados

	for rows.Next() {
		err := rows.Scan(&mailitemID, &sentStatus, &lastSentDate, &lastModDate)
		if err != nil {
			check.Status = "error"
			check.Message = "Error al leer datos: " + err.Error()
			return check
		}

		// Contar cada estado dinámicamente
		statusCount[sentStatus]++
	}

	// Calcular tiempo desde último mail
	var minutesSinceLastMail int64
	if lastSentDate != nil {
		minutesSinceLastMail = int64(time.Since(*lastSentDate).Minutes())
		check.Metadata["last_email_sent"] = lastSentDate.Format(time.RFC3339)
		check.Metadata["minutes_since_last_email"] = minutesSinceLastMail
	}

	// Guardar conteo dinámico completo en metadata
	check.Metadata["status_counts"] = statusCount

	// Mantener compatibilidad con campos específicos
	check.Metadata["failed_emails_last_10"] = statusCount["failed"]
	check.Metadata["success_emails_last_10"] = statusCount["sent"]
	check.Metadata["unsent_emails_last_10"] = statusCount["unsent"]

	// Determinar estado usando umbrales configurables
	// Prioridad: error > warning (unsent) > warning (tiempo)
	failedCount := statusCount["failed"]
	unsentCount := statusCount["unsent"]

	if failedCount > config.MaxFailedCount {
		check.Status = "error"
		check.Message = fmt.Sprintf("Muchos mails fallidos: %d de los últimos 10 (sent: %d, unsent: %d, failed: %d)",
			failedCount, statusCount["sent"], unsentCount, failedCount)
	} else if unsentCount > config.MaxUnsentCount {
		check.Status = "warning"
		check.Message = fmt.Sprintf("Cola de correos atascada: %d de los últimos 10 sin enviar (sent: %d, unsent: %d, failed: %d)",
			unsentCount, statusCount["sent"], unsentCount, failedCount)
	} else if minutesSinceLastMail > int64(config.MaxMinutesWithoutSending) {
		check.Status = "warning"
		check.Message = fmt.Sprintf("Último mail enviado hace %d minutos", minutesSinceLastMail)
	} else {
		check.Status = "ok"
		check.Message = fmt.Sprintf("Servicio funcionando correctamente. Último mail hace %d minutos (sent: %d, unsent: %d, failed: %d)",
			minutesSinceLastMail, statusCount["sent"], unsentCount, failedCount)
	}

	return check
}
