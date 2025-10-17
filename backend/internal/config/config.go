package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config contiene toda la configuración de la aplicación
type Config struct {
	Server             ServerConfig
	DatabaseProd       DatabaseConfig
	DatabasePreProd    DatabaseConfig
	PostgreSQLAppSPC   PostgreSQLConfig
	GoogleSheets       GoogleSheetsConfig
	SaltaCompra        SaltaCompraConfig
	AppSaltaCompra     AppSaltaCompraConfig
	Infrastructure     InfrastructureConfig
	Monitors           MonitorsConfig
	VPNCheck           VPNCheckConfig
	Scheduler          SchedulerConfig
	Cache              CacheConfig
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

// PostgreSQLConfig configuración de conexión a PostgreSQL
type PostgreSQLConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
}

// AppSaltaCompraConfig configuración para monitoreo de App.SaltaCompra
type AppSaltaCompraConfig struct {
	URL                 string
	ExpectedContent     string
	SkipSSLVerification bool // true para certificados autofirmados/CA privada
}

// VPNCheckConfig configuración para verificación de VPN
type VPNCheckConfig struct {
	Host       string // Host para verificar conectividad VPN (normalmente el host de PostgreSQL)
	TimeoutMs  int    // Timeout en milisegundos para la verificación
}

// MonitorsConfig configuración de umbrales para monitores
type MonitorsConfig struct {
	MailMaxMinutesWithoutSent    int   // Minutos máximos sin correo 'sent' antes de warning
	MailDailyWarningFailedPercent int   // % de fallidos del día para warning
	MailDailyErrorFailedPercent   int   // % de fallidos del día para error
	HTTPTimeoutWarningMs         int64 // Umbral de ms para warning en checks HTTP
	HTTPTimeoutErrorMs           int64 // Umbral de ms para error en checks HTTP
	SSLWarningDays               int   // Días antes de expiración SSL para warning
	HTTPTimeoutSeconds           int   // Timeout general para peticiones HTTP
	DomainWarningDays            int   // Días antes de expiración de dominio para warning
	DomainErrorDays              int   // Días antes de expiración de dominio para error
}

// SchedulerConfig configuración para el background worker
type SchedulerConfig struct {
	IntervalMinutes     int // Intervalo en minutos para ejecutar checks automáticamente
	IdleTimeoutMinutes  int // Minutos sin actividad antes de pausar el worker
}

// CacheConfig configuración para el cache de sistemas
type CacheConfig struct {
	MaxAgeMinutes int // Edad máxima en minutos antes de considerar datos desactualizados
}

// LoadConfig carga la configuración desde variables de entorno
// Retorna error si faltan variables requeridas o tienen valores inválidos
func LoadConfig() (Config, error) {
	var errors []string

	// Validar variables requeridas
	requiredVars := []string{
		"SERVER_PORT",
		"DB_PROD_HOST", "DB_PROD_PORT", "DB_PROD_USER", "DB_PROD_PASSWORD", "DB_PROD_NAME",
		"DB_PREPROD_HOST", "DB_PREPROD_PORT", "DB_PREPROD_USER", "DB_PREPROD_PASSWORD", "DB_PREPROD_NAME",
		"DB_APPSALTACOMPRA_HOST", "DB_APPSALTACOMPRA_PORT", "DB_APPSALTACOMPRA_USER", "DB_APPSALTACOMPRA_PASSWORD", "DB_APPSALTACOMPRA_NAME",
		"SALTACOMPRA_PROD_URL", "SALTACOMPRA_PROD_EXPECTED_CONTENT",
		"SALTACOMPRA_PREPROD_URL", "SALTACOMPRA_PREPROD_EXPECTED_CONTENT",
		"APPSALTACOMPRA_URL", "APPSALTACOMPRA_EXPECTED_CONTENT", "APPSALTACOMPRA_SKIP_SSL_VERIFICATION",
		"INFRASTRUCTURE_DOMAIN", "RDAP_BASE_URL",
		"VPN_CHECK_HOST", "VPN_CHECK_TIMEOUT_MS",
		"GSHEETS_SPREADSHEET_ID", "GSHEETS_SHEET_NAME", "GSHEETS_AUTH_METHOD",
		"GSHEETS_CREDENTIALS_FILE", "GSHEETS_TIMESTAMP_COLUMN", "GSHEETS_FILENAME_COLUMN",
		"GSHEETS_WARNING_DAYS", "GSHEETS_ERROR_DAYS",
		"MAIL_MAX_MINUTES_WITHOUT_SENT", "MAIL_DAILY_WARNING_FAILED_PERCENT", "MAIL_DAILY_ERROR_FAILED_PERCENT",
		"HTTP_TIMEOUT_WARNING_MS", "HTTP_TIMEOUT_ERROR_MS", "HTTP_TIMEOUT_SECONDS",
		"SSL_WARNING_DAYS", "DOMAIN_WARNING_DAYS", "DOMAIN_ERROR_DAYS",
		"BACKGROUND_CHECK_INTERVAL_MINUTES", "WORKER_IDLE_TIMEOUT_MINUTES", "CACHE_MAX_AGE_MINUTES",
	}

	for _, v := range requiredVars {
		if os.Getenv(v) == "" {
			errors = append(errors, fmt.Sprintf("Variable requerida no encontrada: %s", v))
		}
	}

	if len(errors) > 0 {
		return Config{}, fmt.Errorf("Errores de configuración:\n- %s", strings.Join(errors, "\n- "))
	}

	// Cargar configuración
	config := Config{
		Server: ServerConfig{
			Port: mustGetEnv("SERVER_PORT"),
		},
		DatabaseProd: DatabaseConfig{
			Host:     mustGetEnv("DB_PROD_HOST"),
			Port:     mustGetEnvAsInt("DB_PROD_PORT"),
			User:     mustGetEnv("DB_PROD_USER"),
			Password: mustGetEnv("DB_PROD_PASSWORD"),
			Database: mustGetEnv("DB_PROD_NAME"),
		},
		DatabasePreProd: DatabaseConfig{
			Host:     mustGetEnv("DB_PREPROD_HOST"),
			Port:     mustGetEnvAsInt("DB_PREPROD_PORT"),
			User:     mustGetEnv("DB_PREPROD_USER"),
			Password: mustGetEnv("DB_PREPROD_PASSWORD"),
			Database: mustGetEnv("DB_PREPROD_NAME"),
		},
		GoogleSheets: GoogleSheetsConfig{
			SpreadsheetID:   mustGetEnv("GSHEETS_SPREADSHEET_ID"),
			SheetName:       mustGetEnv("GSHEETS_SHEET_NAME"),
			AuthMethod:      mustGetEnv("GSHEETS_AUTH_METHOD"),
			CredentialsFile: mustGetEnv("GSHEETS_CREDENTIALS_FILE"),
			APIKey:          os.Getenv("GSHEETS_API_KEY"), // Opcional
			TimestampColumn: mustGetEnvAsInt("GSHEETS_TIMESTAMP_COLUMN"),
			FilenameColumn:  mustGetEnvAsInt("GSHEETS_FILENAME_COLUMN"),
			WarningDays:     mustGetEnvAsInt("GSHEETS_WARNING_DAYS"),
			ErrorDays:       mustGetEnvAsInt("GSHEETS_ERROR_DAYS"),
		},
		SaltaCompra: SaltaCompraConfig{
			ProdURL:                mustGetEnv("SALTACOMPRA_PROD_URL"),
			ProdExpectedContent:    mustGetEnv("SALTACOMPRA_PROD_EXPECTED_CONTENT"),
			PreProdURL:             mustGetEnv("SALTACOMPRA_PREPROD_URL"),
			PreProdExpectedContent: mustGetEnv("SALTACOMPRA_PREPROD_EXPECTED_CONTENT"),
		},
		Infrastructure: InfrastructureConfig{
			Domain:      mustGetEnv("INFRASTRUCTURE_DOMAIN"),
			RDAPBaseURL: mustGetEnv("RDAP_BASE_URL"),
		},
		PostgreSQLAppSPC: PostgreSQLConfig{
			Host:     mustGetEnv("DB_APPSALTACOMPRA_HOST"),
			Port:     mustGetEnvAsInt("DB_APPSALTACOMPRA_PORT"),
			User:     mustGetEnv("DB_APPSALTACOMPRA_USER"),
			Password: mustGetEnv("DB_APPSALTACOMPRA_PASSWORD"),
			Database: mustGetEnv("DB_APPSALTACOMPRA_NAME"),
		},
		AppSaltaCompra: AppSaltaCompraConfig{
			URL:                 mustGetEnv("APPSALTACOMPRA_URL"),
			ExpectedContent:     mustGetEnv("APPSALTACOMPRA_EXPECTED_CONTENT"),
			SkipSSLVerification: mustGetEnvAsBool("APPSALTACOMPRA_SKIP_SSL_VERIFICATION"),
		},
		VPNCheck: VPNCheckConfig{
			Host:      mustGetEnv("VPN_CHECK_HOST"),
			TimeoutMs: mustGetEnvAsInt("VPN_CHECK_TIMEOUT_MS"),
		},
		Monitors: MonitorsConfig{
			MailMaxMinutesWithoutSent:    mustGetEnvAsInt("MAIL_MAX_MINUTES_WITHOUT_SENT"),
			MailDailyWarningFailedPercent: mustGetEnvAsInt("MAIL_DAILY_WARNING_FAILED_PERCENT"),
			MailDailyErrorFailedPercent:   mustGetEnvAsInt("MAIL_DAILY_ERROR_FAILED_PERCENT"),
			HTTPTimeoutWarningMs:         int64(mustGetEnvAsInt("HTTP_TIMEOUT_WARNING_MS")),
			HTTPTimeoutErrorMs:           int64(mustGetEnvAsInt("HTTP_TIMEOUT_ERROR_MS")),
			SSLWarningDays:               mustGetEnvAsInt("SSL_WARNING_DAYS"),
			HTTPTimeoutSeconds:           mustGetEnvAsInt("HTTP_TIMEOUT_SECONDS"),
			DomainWarningDays:            mustGetEnvAsInt("DOMAIN_WARNING_DAYS"),
			DomainErrorDays:              mustGetEnvAsInt("DOMAIN_ERROR_DAYS"),
		},
		Scheduler: SchedulerConfig{
			IntervalMinutes:    mustGetEnvAsInt("BACKGROUND_CHECK_INTERVAL_MINUTES"),
			IdleTimeoutMinutes: mustGetEnvAsInt("WORKER_IDLE_TIMEOUT_MINUTES"),
		},
		Cache: CacheConfig{
			MaxAgeMinutes: mustGetEnvAsInt("CACHE_MAX_AGE_MINUTES"),
		},
	}

	return config, nil
}

// mustGetEnv obtiene una variable de entorno
// Asume que la variable ya fue validada en LoadConfig
func mustGetEnv(key string) string {
	return os.Getenv(key)
}

// mustGetEnvAsInt obtiene una variable de entorno como int
// Asume que la variable ya fue validada en LoadConfig
// Panic si el valor no es un entero válido (esto indica un bug de configuración)
func mustGetEnvAsInt(key string) int {
	valueStr := os.Getenv(key)
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		// Esto no debería pasar si las validaciones se hicieron correctamente
		panic(fmt.Sprintf("Variable %s contiene valor inválido: %s (debe ser un número entero)", key, valueStr))
	}
	return value
}

// mustGetEnvAsBool obtiene una variable de entorno como bool
// Asume que la variable ya fue validada en LoadConfig
// Panic si el valor no es un booleano válido (esto indica un bug de configuración)
func mustGetEnvAsBool(key string) bool {
	valueStr := os.Getenv(key)
	value, err := strconv.ParseBool(valueStr)
	if err != nil {
		panic(fmt.Sprintf("Variable %s contiene valor inválido: %s (debe ser true o false)", key, valueStr))
	}
	return value
}
