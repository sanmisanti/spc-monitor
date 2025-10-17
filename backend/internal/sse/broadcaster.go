package sse

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"github.com/saltacompra/monitor/internal/models"
)

// Client representa un cliente SSE conectado
type Client struct {
	ID      string
	Channel chan string
}

// Broadcaster maneja múltiples clientes SSE
type Broadcaster struct {
	mu      sync.RWMutex
	clients map[string]*Client
}

// NewBroadcaster crea una nueva instancia del broadcaster
func NewBroadcaster() *Broadcaster {
	return &Broadcaster{
		clients: make(map[string]*Client),
	}
}

// Register registra un nuevo cliente SSE
func (b *Broadcaster) Register(clientID string) *Client {
	b.mu.Lock()
	defer b.mu.Unlock()

	client := &Client{
		ID:      clientID,
		Channel: make(chan string, 10), // Buffer de 10 mensajes
	}

	b.clients[clientID] = client
	log.Printf("[SSE] Cliente registrado: %s (total: %d)", clientID, len(b.clients))

	return client
}

// Unregister elimina un cliente SSE
func (b *Broadcaster) Unregister(clientID string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if client, exists := b.clients[clientID]; exists {
		close(client.Channel)
		delete(b.clients, clientID)
		log.Printf("[SSE] Cliente desregistrado: %s (total: %d)", clientID, len(b.clients))
	}
}

// Broadcast envía un mensaje a todos los clientes conectados
func (b *Broadcaster) Broadcast(eventType string, data interface{}) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if len(b.clients) == 0 {
		return
	}

	// Serializar datos a JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Printf("[SSE] Error al serializar datos: %v", err)
		return
	}

	// Formato SSE: event: tipo\ndata: json\n\n
	message := fmt.Sprintf("event: %s\ndata: %s\n\n", eventType, string(jsonData))

	// Enviar a todos los clientes
	for clientID, client := range b.clients {
		select {
		case client.Channel <- message:
			// Mensaje enviado correctamente
		default:
			log.Printf("[SSE] Buffer lleno para cliente %s, mensaje descartado", clientID)
		}
	}

	log.Printf("[SSE] Broadcast enviado: %s a %d clientes", eventType, len(b.clients))
}

// BroadcastSystem envía el estado actualizado de un sistema específico
func (b *Broadcaster) BroadcastSystem(system models.System) {
	b.Broadcast("system_update", system)
}

// BroadcastCheckComplete envía notificación de que todos los checks completaron
func (b *Broadcaster) BroadcastCheckComplete() {
	b.Broadcast("check_complete", map[string]interface{}{
		"message": "Todos los checks han completado",
	})
}

// ClientCount retorna el número de clientes conectados
func (b *Broadcaster) ClientCount() int {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return len(b.clients)
}
