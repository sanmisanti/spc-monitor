package config

import (
	"os"
	"strconv"
)

// Config contiene toda la configuración de la aplicación
type Config struct {
	Server       ServerConfig
	DatabaseProd DatabaseConfig
	DatabasePreProd DatabaseConfig
	Monitors     MonitorsConfig
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

// MonitorsConfig configuración de umbrales para monitores
type MonitorsConfig struct {
	MailMaxMinutesWithoutSending int   // Minutos máximos sin enviar mails antes de warning
	MailMaxFailedCount           int   // Cantidad máxima de mails fallidos antes de error
	HTTPTimeoutWarningMs         int64 // Umbral de ms para warning en checks HTTP
	HTTPTimeoutErrorMs           int64 // Umbral de ms para error en checks HTTP
	SSLWarningDays               int   // Días antes de expiración SSL para warning
	HTTPTimeoutSeconds           int   // Timeout general para peticiones HTTP
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
		Monitors: MonitorsConfig{
			MailMaxMinutesWithoutSending: getEnvAsInt("MAIL_MAX_MINUTES_WITHOUT_SENDING", 180),
			MailMaxFailedCount:           getEnvAsInt("MAIL_MAX_FAILED_COUNT", 5),
			HTTPTimeoutWarningMs:         int64(getEnvAsInt("HTTP_TIMEOUT_WARNING_MS", 3000)),
			HTTPTimeoutErrorMs:           int64(getEnvAsInt("HTTP_TIMEOUT_ERROR_MS", 10000)),
			SSLWarningDays:               getEnvAsInt("SSL_WARNING_DAYS", 30),
			HTTPTimeoutSeconds:           getEnvAsInt("HTTP_TIMEOUT_SECONDS", 30),
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
