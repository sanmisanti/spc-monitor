package cache

import (
	"sync"
	"time"

	"github.com/saltacompra/monitor/internal/models"
)

// CachedSystem representa un sistema con su timestamp de actualización
type CachedSystem struct {
	Data      models.System
	UpdatedAt time.Time
}

// SystemCache es un cache thread-safe para almacenar el estado de los sistemas
type SystemCache struct {
	mu      sync.RWMutex
	systems map[string]CachedSystem
}

// NewSystemCache crea una nueva instancia del cache
func NewSystemCache() *SystemCache {
	return &SystemCache{
		systems: make(map[string]CachedSystem),
	}
}

// Get obtiene un sistema del cache por su ID
// Retorna el sistema y true si existe, o un sistema vacío y false si no existe
func (c *SystemCache) Get(id string) (models.System, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	cached, exists := c.systems[id]
	if !exists {
		return models.System{}, false
	}

	return cached.Data, true
}

// Set guarda o actualiza un sistema en el cache
func (c *SystemCache) Set(id string, system models.System) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.systems[id] = CachedSystem{
		Data:      system,
		UpdatedAt: time.Now(),
	}
}

// GetAll obtiene todos los sistemas del cache
// Retorna un slice vacío si el cache está vacío
func (c *SystemCache) GetAll() []models.System {
	c.mu.RLock()
	defer c.mu.RUnlock()

	systems := make([]models.System, 0, len(c.systems))
	for _, cached := range c.systems {
		systems = append(systems, cached.Data)
	}

	return systems
}

// IsStale verifica si un sistema está desactualizado según el maxAge
// Retorna true si no existe o si está desactualizado
func (c *SystemCache) IsStale(id string, maxAge time.Duration) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	cached, exists := c.systems[id]
	if !exists {
		return true
	}

	return time.Since(cached.UpdatedAt) > maxAge
}

// Clear limpia todo el cache
func (c *SystemCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.systems = make(map[string]CachedSystem)
}

// GetTimestamp obtiene el timestamp de última actualización de un sistema
// Retorna el timestamp y true si existe, o zero time y false si no existe
func (c *SystemCache) GetTimestamp(id string) (time.Time, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	cached, exists := c.systems[id]
	if !exists {
		return time.Time{}, false
	}

	return cached.UpdatedAt, true
}

// Count retorna la cantidad de sistemas en el cache
func (c *SystemCache) Count() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.systems)
}
