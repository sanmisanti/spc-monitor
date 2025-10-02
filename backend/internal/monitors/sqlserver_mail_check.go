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

	// Query para explorar la tabla sysmail_allitems
	// Primero vamos a ver qué columnas tiene y los últimos registros
	query := `
		SELECT TOP 10
			mailitem_id,
			sent_status,
			sent_date,
			last_mod_date
		FROM msdb.dbo.sysmail_allitems
		ORDER BY mailitem_id DESC
	`

	rows, err := db.Query(query)
	if err != nil {
		check.Status = "error"
		check.Message = "Error al consultar sysmail_allitems: " + err.Error()
		return check
	}
	defer rows.Close()

	// Procesar resultados para análisis
	var lastSentDate, lastModDate *time.Time
	var mailitemID int
	var sentStatus string
	failedCount := 0
	successCount := 0

	for rows.Next() {
		err := rows.Scan(&mailitemID, &sentStatus, &lastSentDate, &lastModDate)
		if err != nil {
			check.Status = "error"
			check.Message = "Error al leer datos: " + err.Error()
			return check
		}

		// Contar mails exitosos vs fallidos
		if sentStatus == "sent" {
			successCount++
		} else if sentStatus == "failed" {
			failedCount++
		}
	}

	// Calcular tiempo desde último mail
	var minutesSinceLastMail int64
	if lastSentDate != nil {
		minutesSinceLastMail = int64(time.Since(*lastSentDate).Minutes())
		check.Metadata["last_email_sent"] = lastSentDate.Format(time.RFC3339)
		check.Metadata["minutes_since_last_email"] = minutesSinceLastMail
	}

	check.Metadata["failed_emails_last_10"] = failedCount
	check.Metadata["success_emails_last_10"] = successCount

	// Determinar estado usando umbrales configurables
	if failedCount > config.MaxFailedCount {
		check.Status = "error"
		check.Message = fmt.Sprintf("Muchos mails fallidos: %d de los últimos 10", failedCount)
	} else if minutesSinceLastMail > int64(config.MaxMinutesWithoutSending) {
		check.Status = "warning"
		check.Message = fmt.Sprintf("Último mail enviado hace %d minutos", minutesSinceLastMail)
	} else {
		check.Status = "ok"
		check.Message = fmt.Sprintf("Servicio funcionando correctamente. Último mail hace %d minutos", minutesSinceLastMail)
	}

	return check
}
