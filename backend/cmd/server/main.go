package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/saltacompra/monitor/internal/api"
	"github.com/saltacompra/monitor/internal/cache"
	"github.com/saltacompra/monitor/internal/config"
	"github.com/saltacompra/monitor/internal/scheduler"
	"github.com/saltacompra/monitor/internal/sse"
)

func main() {
	// Cargar archivo .env (ejecutar desde backend/)
	if err := godotenv.Load(".env"); err != nil {
		log.Fatal("ERROR CRÍTICO: No se pudo cargar el archivo .env - ", err)
	}

	// Cargar configuración desde variables de entorno
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("ERROR CRÍTICO: Configuración inválida - ", err)
	}

	// Inicializar componentes
	log.Println("[INIT] Inicializando componentes...")

	// 1. Cache de sistemas
	systemCache := cache.NewSystemCache()
	log.Println("[INIT] Cache inicializado")

	// 2. Broadcaster SSE
	broadcaster := sse.NewBroadcaster()
	log.Println("[INIT] Broadcaster SSE inicializado")

	// 3. Handler (con cache y broadcaster)
	handler := api.NewHandler(cfg, systemCache, broadcaster)
	log.Println("[INIT] Handler inicializado")

	// 4. Background Worker (con función de checks)
	worker := scheduler.NewSmartWorker(cfg, systemCache, broadcaster, handler.CheckAllSystems)
	worker.Start()
	log.Printf("[INIT] Background worker iniciado (intervalo: %d min, idle timeout: %d min)",
		cfg.Scheduler.IntervalMinutes, cfg.Scheduler.IdleTimeoutMinutes)

	// Ejecutar checks iniciales en background
	go func() {
		log.Println("[INIT] Ejecutando checks iniciales...")
		systems := handler.CheckAllSystems()
		for _, system := range systems {
			systemCache.Set(system.ID, system)
		}
		log.Printf("[INIT] Checks iniciales completados: %d sistemas listos", len(systems))
	}()

	// Middleware CORS
	corsMiddleware := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

			// Handle preflight requests
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next(w, r)
		}
	}

	// Configurar rutas con CORS
	http.HandleFunc("/api/systems", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		worker.MarkActivity()
		handler.GetSystems(w, r)
	}))

	http.HandleFunc("/api/events", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		worker.MarkActivity()
		handler.GetEvents(w, r)
	}))

	http.HandleFunc("/api/refresh", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		worker.MarkActivity()
		handler.RefreshAllSystems(w, r)
	}))

	http.HandleFunc("/api/systems/", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		worker.MarkActivity()
		handler.RefreshSystem(w, r)
	}))

	http.HandleFunc("/api/health", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	}))

	// Setup graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Servidor HTTP en goroutine
	addr := ":" + cfg.Server.Port
	log.Printf("[SERVER] Servidor iniciado en http://localhost%s", addr)
	log.Printf("[SERVER] API disponible:")
	log.Printf("[SERVER]   GET  /api/systems - Lista de sistemas (cache)")
	log.Printf("[SERVER]   GET  /api/events - Stream SSE de updates")
	log.Printf("[SERVER]   POST /api/refresh - Refresh de todos los sistemas")
	log.Printf("[SERVER]   POST /api/systems/:id/refresh - Refresh de sistema individual")

	go func() {
		if err := http.ListenAndServe(addr, nil); err != nil {
			log.Fatal("[ERROR] Error al iniciar servidor:", err)
		}
	}()

	// Esperar señal de shutdown
	<-stop
	log.Println("\n[SHUTDOWN] Señal recibida, cerrando servidor...")

	// Detener worker
	worker.Stop()
	log.Println("[SHUTDOWN] Background worker detenido")

	log.Println("[SHUTDOWN] Servidor cerrado correctamente")
}
