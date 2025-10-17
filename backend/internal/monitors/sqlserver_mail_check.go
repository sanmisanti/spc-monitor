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
	MaxMinutesWithoutSent        int // Umbral de minutos sin correo 'sent' antes de warning
	DailyWarningFailedPercent    int // % de fallidos para warning
	DailyErrorFailedPercent      int // % de fallidos para error
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

	// Query para obtener todos los correos del día
	// Intentar con send_request_date primero, si falla usar last_mod_date como fallback
	query := `
		SELECT
			mailitem_id,
			CASE sent_status
				WHEN 0 THEN 'unsent'
				WHEN 1 THEN 'sent'
				WHEN 3 THEN 'retrying'
				ELSE 'failed'
			END as sent_status,
			send_request_date,
			sent_date,
			last_mod_date
		FROM msdb.dbo.sysmail_mailitems
		WHERE CAST(send_request_date AS DATE) = CAST(GETDATE() AS DATE)
		ORDER BY mailitem_id DESC
	`

	rows, err := db.Query(query)
	if err != nil {
		// Si send_request_date no existe, usar last_mod_date como fallback
		query = `
			SELECT
				mailitem_id,
				CASE sent_status
					WHEN 0 THEN 'unsent'
					WHEN 1 THEN 'sent'
					WHEN 3 THEN 'retrying'
					ELSE 'failed'
				END as sent_status,
				last_mod_date as send_request_date,
				sent_date,
				last_mod_date
			FROM msdb.dbo.sysmail_mailitems
			WHERE CAST(last_mod_date AS DATE) = CAST(GETDATE() AS DATE)
			ORDER BY mailitem_id DESC
		`
		rows, err = db.Query(query)
		if err != nil {
			check.Status = "error"
			check.Message = "Error al consultar sysmail_mailitems: " + err.Error()
			return check
		}
	}
	defer rows.Close()

	// Variables para análisis
	var mailitemID int
	var sentStatus string
	var sendRequestDate, sentDate, lastModDate *time.Time

	statusCount := make(map[string]int)
	var lastSentTime, lastCreatedTime *time.Time
	totalToday := 0

	// Procesar todos los registros del día
	for rows.Next() {
		err := rows.Scan(&mailitemID, &sentStatus, &sendRequestDate, &sentDate, &lastModDate)
		if err != nil {
			check.Status = "error"
			check.Message = "Error al leer datos: " + err.Error()
			return check
		}

		totalToday++
		statusCount[sentStatus]++

		// Rastrear último correo enviado (sent)
		if sentStatus == "sent" && sentDate != nil {
			if lastSentTime == nil || sentDate.After(*lastSentTime) {
				lastSentTime = sentDate
			}
		}

		// Rastrear último registro creado
		if sendRequestDate != nil {
			if lastCreatedTime == nil || sendRequestDate.After(*lastCreatedTime) {
				lastCreatedTime = sendRequestDate
			}
		}
	}

	// Si no hay correos del día, reportar estado especial
	if totalToday == 0 {
		check.Status = "warning"
		check.Message = "No hay correos registrados hoy"
		check.Metadata["today_total"] = 0
		return check
	}

	// Calcular métricas
	sentCount := statusCount["sent"]
	unsentCount := statusCount["unsent"]
	failedCount := statusCount["failed"]
	retryingCount := statusCount["retrying"]

	failedPercent := float64(0)
	if totalToday > 0 {
		failedPercent = (float64(failedCount) / float64(totalToday)) * 100
	}

	// Tiempos desde última acción
	var minutesSinceLastSent, minutesSinceLastCreated int64
	if lastSentTime != nil {
		minutesSinceLastSent = int64(time.Since(*lastSentTime).Minutes())
		check.Metadata["last_sent_time"] = lastSentTime.Format(time.RFC3339)
		check.Metadata["minutes_since_last_sent"] = minutesSinceLastSent
	}
	if lastCreatedTime != nil {
		minutesSinceLastCreated = int64(time.Since(*lastCreatedTime).Minutes())
		check.Metadata["last_created_time"] = lastCreatedTime.Format(time.RFC3339)
		check.Metadata["minutes_since_last_created"] = minutesSinceLastCreated
	}

	// Metadata completa
	check.Metadata["today_total"] = totalToday
	check.Metadata["today_sent"] = sentCount
	check.Metadata["today_unsent"] = unsentCount
	check.Metadata["today_failed"] = failedCount
	check.Metadata["today_retrying"] = retryingCount
	check.Metadata["today_failed_percentage"] = fmt.Sprintf("%.2f", failedPercent)
	check.Metadata["status_counts"] = statusCount

	// Determinar estado según umbrales
	// Prioridad: error (% crítico) > warning (% alto) > warning (sin envíos recientes) > ok
	if failedPercent >= float64(config.DailyErrorFailedPercent) {
		check.Status = "error"
		check.Message = fmt.Sprintf("%.1f%% de correos fallidos hoy (%d de %d). Revisar configuración SMTP",
			failedPercent, failedCount, totalToday)
	} else if failedPercent >= float64(config.DailyWarningFailedPercent) {
		check.Status = "warning"
		check.Message = fmt.Sprintf("%.1f%% de correos fallidos hoy (%d de %d). Último envío hace %d min",
			failedPercent, failedCount, totalToday, minutesSinceLastSent)
	} else if lastSentTime == nil || minutesSinceLastSent > int64(config.MaxMinutesWithoutSent) {
		check.Status = "warning"
		if lastSentTime == nil {
			check.Message = fmt.Sprintf("No hay correos enviados hoy (%d pendientes, %d fallidos)",
				unsentCount, failedCount)
		} else {
			check.Message = fmt.Sprintf("Sin correos enviados hace %d minutos (%d de %d enviados hoy)",
				minutesSinceLastSent, sentCount, totalToday)
		}
	} else {
		check.Status = "ok"
		check.Message = fmt.Sprintf("Servicio funcionando. %d de %d correos enviados hoy (%.1f%% fallidos, último envío hace %d min)",
			sentCount, totalToday, failedPercent, minutesSinceLastSent)
	}

	return check
}
