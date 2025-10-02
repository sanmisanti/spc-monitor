package models

import "time"

// System representa un sistema monitoreado
type System struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Type        string    `json:"type"`        // "web", "api", "google-script"
	Environment string    `json:"environment"` // "prod", "preprod"
	Status      string    `json:"status"`      // "online", "offline", "degraded", "unknown"
	LastCheck   time.Time `json:"last_check"`
	Checks      []Check   `json:"checks"`
}

// Check representa una verificación individual
type Check struct {
	ID           string                 `json:"id"`
	Type         string                 `json:"type"`    // "http", "database", "login", "custom"
	Name         string                 `json:"name"`    // Nombre descriptivo del check
	Status       string                 `json:"status"`  // "ok", "warning", "error"
	Message      string                 `json:"message"` // Descripción del estado
	LastCheck    time.Time              `json:"last_check"`
	ResponseTime int64                  `json:"response_time_ms"` // Tiempo de respuesta en ms
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}
