package config

import (
	"os"
	"strconv"
)

// Config contiene toda la configuración de la aplicación
type Config struct {
	Server          ServerConfig
	DatabaseProd    DatabaseConfig
	DatabasePreProd DatabaseConfig
	GoogleSheets    GoogleSheetsConfig
	SaltaCompra     SaltaCompraConfig
	Infrastructure  InfrastructureConfig
	Monitors        MonitorsConfig
}

// ServerConfig configuración del servidor HTTP
type ServerConfig struct {
	Port string
}

// DatabaseConfig configuración de conexión a SQL Server
type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
}

// GoogleSheetsConfig configuración para Google Sheets API
type GoogleSheetsConfig struct {
	SpreadsheetID   string
	SheetName       string
	AuthMethod      string // "service_account" o "api_key"
	CredentialsFile string // Ruta al JSON de service account
	APIKey          string // API key (alternativa)
	TimestampColumn int    // Índice de columna TimeStamp (0)
	FilenameColumn  int    // Índice de columna Nombre Archivo (2)
	WarningDays     int    // Días de antigüedad para warning
	ErrorDays       int    // Días de antigüedad para error
}

// SaltaCompraConfig configuración para monitoreo de SaltaCompra
type SaltaCompraConfig struct {
	ProdURL                string
	ProdExpectedContent    string
	PreProdURL             string
	PreProdExpectedContent string
}

// InfrastructureConfig configuración para infraestructura compartida
type InfrastructureConfig struct {
	Domain      string
	RDAPBaseURL string
}

// MonitorsConfig configuración de umbrales para monitores
type MonitorsConfig struct {
	MailMaxMinutesWithoutSending int   // Minutos máximos sin enviar mails antes de warning
	MailMaxFailedCount           int   // Cantidad máxima de mails fallidos antes de error
	MailMaxUnsentCount           int   // Cantidad máxima de mails unsent antes de warning (cola atascada)
	HTTPTimeoutWarningMs         int64 // Umbral de ms para warning en checks HTTP
	HTTPTimeoutErrorMs           int64 // Umbral de ms para error en checks HTTP
	SSLWarningDays               int   // Días antes de expiración SSL para warning
	HTTPTimeoutSeconds           int   // Timeout general para peticiones HTTP
	DomainWarningDays            int   // Días antes de expiración de dominio para warning
	DomainErrorDays              int   // Días antes de expiración de dominio para error
}

// LoadConfig carga la configuración desde variables de entorno
func LoadConfig() Config {
	return Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
		},
		DatabaseProd: DatabaseConfig{
			Host:     getEnv("DB_PROD_HOST", "localhost"),
			Port:     getEnvAsInt("DB_PROD_PORT", 1433),
			User:     getEnv("DB_PROD_USER", "sa"),
			Password: getEnv("DB_PROD_PASSWORD", ""),
			Database: getEnv("DB_PROD_NAME", "master"),
		},
		DatabasePreProd: DatabaseConfig{
			Host:     getEnv("DB_PREPROD_HOST", "localhost"),
			Port:     getEnvAsInt("DB_PREPROD_PORT", 1433),
			User:     getEnv("DB_PREPROD_USER", "sa"),
			Password: getEnv("DB_PREPROD_PASSWORD", ""),
			Database: getEnv("DB_PREPROD_NAME", "master"),
		},
		GoogleSheets: GoogleSheetsConfig{
			SpreadsheetID:   getEnv("GSHEETS_SPREADSHEET_ID", ""),
			SheetName:       getEnv("GSHEETS_SHEET_NAME", "Log Actualizaciones Kairos"),
			AuthMethod:      getEnv("GSHEETS_AUTH_METHOD", "service_account"),
			CredentialsFile: getEnv("GSHEETS_CREDENTIALS_FILE", "./credentials/service-account.json"),
			APIKey:          getEnv("GSHEETS_API_KEY", ""),
			TimestampColumn: getEnvAsInt("GSHEETS_TIMESTAMP_COLUMN", 0),
			FilenameColumn:  getEnvAsInt("GSHEETS_FILENAME_COLUMN", 2),
			WarningDays:     getEnvAsInt("GSHEETS_WARNING_DAYS", 1),
			ErrorDays:       getEnvAsInt("GSHEETS_ERROR_DAYS", 2),
		},
		SaltaCompra: SaltaCompraConfig{
			ProdURL:                getEnv("SALTACOMPRA_PROD_URL", "https://saltacompra.gob.ar/"),
			ProdExpectedContent:    getEnv("SALTACOMPRA_PROD_EXPECTED_CONTENT", "SALTA COMPRA - Portal de Compras Públicas de la Provincia de Salta"),
			PreProdURL:             getEnv("SALTACOMPRA_PREPROD_URL", "https://preproduccion.saltacompra.gob.ar/"),
			PreProdExpectedContent: getEnv("SALTACOMPRA_PREPROD_EXPECTED_CONTENT", "SALTA COMPRA - Portal de Compras Públicas de la Provincia de Salta"),
		},
		Infrastructure: InfrastructureConfig{
			Domain:      getEnv("INFRASTRUCTURE_DOMAIN", "saltacompra.gob.ar"),
			RDAPBaseURL: getEnv("RDAP_BASE_URL", "https://rdap.nic.ar/domain/"),
		},
		Monitors: MonitorsConfig{
			MailMaxMinutesWithoutSending: getEnvAsInt("MAIL_MAX_MINUTES_WITHOUT_SENDING", 180),
			MailMaxFailedCount:           getEnvAsInt("MAIL_MAX_FAILED_COUNT", 5),
			MailMaxUnsentCount:           getEnvAsInt("MAIL_MAX_UNSENT_COUNT", 7),
			HTTPTimeoutWarningMs:         int64(getEnvAsInt("HTTP_TIMEOUT_WARNING_MS", 3000)),
			HTTPTimeoutErrorMs:           int64(getEnvAsInt("HTTP_TIMEOUT_ERROR_MS", 10000)),
			SSLWarningDays:               getEnvAsInt("SSL_WARNING_DAYS", 30),
			HTTPTimeoutSeconds:           getEnvAsInt("HTTP_TIMEOUT_SECONDS", 30),
			DomainWarningDays:            getEnvAsInt("DOMAIN_WARNING_DAYS", 60),
			DomainErrorDays:              getEnvAsInt("DOMAIN_ERROR_DAYS", 30),
		},
	}
}

// getEnv obtiene una variable de entorno o devuelve un valor por defecto
func getEnv(key string, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// getEnvAsInt obtiene una variable de entorno como int o devuelve un valor por defecto
func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}
