package scheduler

import (
	"log"
	"sync"
	"time"

	"github.com/saltacompra/monitor/internal/cache"
	"github.com/saltacompra/monitor/internal/config"
	"github.com/saltacompra/monitor/internal/models"
	"github.com/saltacompra/monitor/internal/sse"
)

// CheckFunc es una función que ejecuta checks y retorna sistemas
type CheckFunc func() []models.System

// SmartWorker es un worker inteligente que ejecuta checks periódicamente
// Se pausa automáticamente si no hay actividad
type SmartWorker struct {
	config          config.Config
	cache           *cache.SystemCache
	broadcaster     *sse.Broadcaster
	checkFunc       CheckFunc
	interval        time.Duration
	idleTimeout     time.Duration
	lastActivity    time.Time
	lastActivityMu  sync.RWMutex
	ticker          *time.Ticker
	stopChan        chan struct{}
	running         bool
	runningMu       sync.RWMutex
}

// NewSmartWorker crea una nueva instancia del worker
func NewSmartWorker(
	cfg config.Config,
	cache *cache.SystemCache,
	broadcaster *sse.Broadcaster,
	checkFunc CheckFunc,
) *SmartWorker {
	return &SmartWorker{
		config:       cfg,
		cache:        cache,
		broadcaster:  broadcaster,
		checkFunc:    checkFunc,
		interval:     time.Duration(cfg.Scheduler.IntervalMinutes) * time.Minute,
		idleTimeout:  time.Duration(cfg.Scheduler.IdleTimeoutMinutes) * time.Minute,
		lastActivity: time.Now(),
		stopChan:     make(chan struct{}),
		running:      false,
	}
}

// Start inicia el background worker
func (w *SmartWorker) Start() {
	w.runningMu.Lock()
	if w.running {
		w.runningMu.Unlock()
		return
	}
	w.running = true
	w.runningMu.Unlock()

	log.Printf("[Worker] Iniciando background worker (intervalo: %v, idle timeout: %v)",
		w.interval, w.idleTimeout)

	w.ticker = time.NewTicker(w.interval)

	go func() {
		for {
			select {
			case <-w.ticker.C:
				w.tick()
			case <-w.stopChan:
				log.Println("[Worker] Deteniendo background worker")
				return
			}
		}
	}()
}

// Stop detiene el background worker
func (w *SmartWorker) Stop() {
	w.runningMu.Lock()
	defer w.runningMu.Unlock()

	if !w.running {
		return
	}

	w.running = false
	if w.ticker != nil {
		w.ticker.Stop()
	}
	close(w.stopChan)
}

// MarkActivity registra actividad (una nueva request llegó)
func (w *SmartWorker) MarkActivity() {
	w.lastActivityMu.Lock()
	defer w.lastActivityMu.Unlock()

	w.lastActivity = time.Now()
}

// tick se ejecuta en cada intervalo del ticker
func (w *SmartWorker) tick() {
	w.lastActivityMu.RLock()
	timeSinceLastActivity := time.Since(w.lastActivity)
	w.lastActivityMu.RUnlock()

	// Si no hay actividad reciente, pausar
	if timeSinceLastActivity > w.idleTimeout {
		log.Printf("[Worker] Sin actividad por %v (> %v), pausando checks",
			timeSinceLastActivity.Round(time.Minute), w.idleTimeout)
		return
	}

	log.Printf("[Worker] Ejecutando checks periódicos (última actividad hace %v)",
		timeSinceLastActivity.Round(time.Minute))

	w.ExecuteChecks()
}

// ExecuteChecks ejecuta todos los checks y actualiza cache/SSE
func (w *SmartWorker) ExecuteChecks() {
	log.Println("[Worker] Iniciando ejecución de checks...")

	systems := w.checkFunc()

	// Actualizar cache y enviar eventos SSE
	for _, system := range systems {
		w.cache.Set(system.ID, system)
		w.broadcaster.BroadcastSystem(system)
	}

	w.broadcaster.BroadcastCheckComplete()
	log.Printf("[Worker] Checks completados: %d sistemas actualizados", len(systems))
}

// IsIdle retorna true si el worker está en modo idle (pausado)
func (w *SmartWorker) IsIdle() bool {
	w.lastActivityMu.RLock()
	defer w.lastActivityMu.RUnlock()

	return time.Since(w.lastActivity) > w.idleTimeout
}

// TimeSinceLastActivity retorna el tiempo desde la última actividad
func (w *SmartWorker) TimeSinceLastActivity() time.Duration {
	w.lastActivityMu.RLock()
	defer w.lastActivityMu.RUnlock()

	return time.Since(w.lastActivity)
}
