package monitors

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"

	"github.com/saltacompra/monitor/internal/models"
)

// GoogleSheetsCheckConfig configuración para verificación de Google Sheets
type GoogleSheetsCheckConfig struct {
	SpreadsheetID   string
	SheetName       string
	AuthMethod      string // "service_account" o "api_key"
	CredentialsFile string
	APIKey          string
	TimestampColumn int
	FilenameColumn  int
	WarningDays     int
	ErrorDays       int
	CheckID         string
	CheckName       string
}

// CheckGoogleSheetsKairos verifica la actualización diaria en Google Sheets
func CheckGoogleSheetsKairos(config GoogleSheetsCheckConfig) models.Check {
	check := models.Check{
		ID:        config.CheckID,
		Type:      "google-sheets",
		Name:      config.CheckName,
		LastCheck: time.Now(),
		Metadata:  make(map[string]interface{}),
	}

	start := time.Now()

	// Crear servicio de Google Sheets según método de autenticación
	srv, err := createSheetsService(config.AuthMethod, config.CredentialsFile, config.APIKey)
	if err != nil {
		check.Status = "error"
		check.Message = "Error al conectar con Google Sheets API: " + err.Error()
		return check
	}

	// Obtener la última fila de datos
	readRange := fmt.Sprintf("%s!A:J", config.SheetName) // Columnas A-J
	resp, err := srv.Spreadsheets.Values.Get(config.SpreadsheetID, readRange).Do()
	elapsed := time.Since(start).Milliseconds()
	check.ResponseTime = elapsed

	if err != nil {
		check.Status = "error"
		check.Message = "Error al leer datos de la hoja: " + err.Error()
		return check
	}

	if len(resp.Values) == 0 {
		check.Status = "error"
		check.Message = "La hoja no contiene datos"
		return check
	}

	// Obtener última fila (omitir headers)
	lastRow := resp.Values[len(resp.Values)-1]

	if len(lastRow) <= config.FilenameColumn {
		check.Status = "error"
		check.Message = "La última fila no tiene todas las columnas esperadas"
		return check
	}

	// Extraer valores
	timestampStr := getStringValue(lastRow, config.TimestampColumn)
	filenameStr := getStringValue(lastRow, config.FilenameColumn)

	check.Metadata["last_timestamp"] = timestampStr
	check.Metadata["last_filename"] = filenameStr
	check.Metadata["total_rows"] = len(resp.Values) - 1 // Sin headers

	// Validar fecha en TimeStamp
	timestampDate, err := time.Parse("2006-01-02", timestampStr)
	if err != nil {
		check.Status = "error"
		check.Message = fmt.Sprintf("Formato de fecha inválido en TimeStamp: %s", timestampStr)
		return check
	}

	// Validar formato de nombre de archivo (bd_YYYYMMDD.zip)
	expectedFilename := fmt.Sprintf("bd_%s.zip", timestampDate.Format("20060102"))

	// También extraer fecha del nombre de archivo para validación adicional
	filenameDate, err := extractDateFromFilename(filenameStr)
	if err != nil {
		check.Status = "warning"
		check.Message = fmt.Sprintf("Formato de nombre de archivo no reconocido: %s (esperado: %s)", filenameStr, expectedFilename)
		check.Metadata["expected_filename"] = expectedFilename
		return check
	}

	check.Metadata["filename_date"] = filenameDate.Format("2006-01-02")
	check.Metadata["expected_filename"] = expectedFilename

	// Calcular días de antigüedad
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	timestampDay := time.Date(timestampDate.Year(), timestampDate.Month(), timestampDate.Day(), 0, 0, 0, 0, time.UTC)

	daysOld := int(today.Sub(timestampDay).Hours() / 24)
	check.Metadata["days_old"] = daysOld

	// Validar coherencia: fecha del TimeStamp debe coincidir con fecha del archivo
	if !timestampDate.Equal(filenameDate) {
		check.Status = "error"
		check.Message = fmt.Sprintf("Inconsistencia: TimeStamp (%s) no coincide con fecha del archivo (%s)",
			timestampDate.Format("2006-01-02"), filenameDate.Format("2006-01-02"))
		return check
	}

	// Determinar estado según antigüedad
	if daysOld == 0 {
		// Datos de hoy - TODO OK
		check.Status = "ok"
		check.Message = fmt.Sprintf("Actualización del día completada con %s", filenameStr)
	} else if daysOld <= config.WarningDays {
		// Datos de ayer - WARNING
		check.Status = "warning"
		check.Message = fmt.Sprintf("Última actualización es de hace %d día(s): %s con %s",
			daysOld, timestampStr, filenameStr)
	} else {
		// Datos muy antiguos - ERROR
		check.Status = "error"
		check.Message = fmt.Sprintf("Actualización desactualizada: hace %d días (%s con %s)",
			daysOld, timestampStr, filenameStr)
	}

	return check
}

// createSheetsService crea un servicio de Google Sheets según el método de auth
func createSheetsService(authMethod, credentialsFile, apiKey string) (*sheets.Service, error) {
	ctx := context.Background()

	if authMethod == "api_key" && apiKey != "" {
		return sheets.NewService(ctx, option.WithAPIKey(apiKey))
	}

	// Service Account (default)
	if credentialsFile == "" {
		return nil, fmt.Errorf("no se especificó archivo de credenciales")
	}

	return sheets.NewService(ctx, option.WithCredentialsFile(credentialsFile))
}

// getStringValue obtiene un valor string de una fila de forma segura
func getStringValue(row []interface{}, index int) string {
	if index >= len(row) {
		return ""
	}
	if row[index] == nil {
		return ""
	}
	return fmt.Sprintf("%v", row[index])
}

// extractDateFromFilename extrae la fecha de un nombre de archivo tipo "bd_YYYYMMDD.zip"
func extractDateFromFilename(filename string) (time.Time, error) {
	// Regex para extraer fecha: bd_YYYYMMDD.zip
	re := regexp.MustCompile(`bd_(\d{8})\.zip`)
	matches := re.FindStringSubmatch(filename)

	if len(matches) < 2 {
		return time.Time{}, fmt.Errorf("formato de archivo inválido")
	}

	dateStr := matches[1] // YYYYMMDD
	return time.Parse("20060102", dateStr)
}
