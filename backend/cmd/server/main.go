package main

import (
	"log"
	"net/http"

	"github.com/joho/godotenv"
	"github.com/saltacompra/monitor/internal/api"
	"github.com/saltacompra/monitor/internal/config"
)

func main() {
	// Cargar archivo .env
	if err := godotenv.Load("../../../.env"); err != nil {
		log.Println("No se pudo cargar .env, usando variables de entorno del sistema")
	}

	// Cargar configuración desde variables de entorno
	cfg := config.LoadConfig()

	// Crear handler
	handler := api.NewHandler(cfg)

	// Configurar rutas
	http.HandleFunc("/api/systems", handler.GetSystems)
	http.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	// Servidor HTTP
	addr := ":" + cfg.Server.Port
	log.Printf("Servidor iniciado en http://localhost%s", addr)
	log.Printf("API disponible en http://localhost%s/api/systems", addr)

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal("Error al iniciar servidor:", err)
	}
}
