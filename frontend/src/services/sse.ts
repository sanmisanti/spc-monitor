import type { System, SSEConnectedEvent, SSECheckCompleteEvent } from '../types/system';

const API_BASE = 'http://localhost:8080/api';

export interface SSECallbacks {
  onSystemUpdate?: (system: System) => void;
  onCheckComplete?: (data: SSECheckCompleteEvent) => void;
  onConnected?: (data: SSEConnectedEvent) => void;
  onError?: (error: Event) => void;
}

/**
 * Cliente SSE simple que envuelve EventSource nativo
 */
export class SSEClient {
  private eventSource: EventSource | null = null;
  private callbacks: SSECallbacks = {};

  /**
   * Conecta al stream SSE del backend
   */
  connect(callbacks: SSECallbacks): () => void {
    if (this.eventSource) {
      console.warn('[SSE] Ya existe una conexión activa');
      return () => {};
    }

    this.callbacks = callbacks;
    this.eventSource = new EventSource(`${API_BASE}/events`);

    // Evento: connected
    this.eventSource.addEventListener('connected', (event: MessageEvent) => {
      try {
        const data: SSEConnectedEvent = JSON.parse(event.data);
        console.log('[SSE] Conectado:', data.client_id);
        this.callbacks.onConnected?.(data);
      } catch (error) {
        console.error('[SSE] Error parsing connected event:', error);
      }
    });

    // Evento: system_update
    this.eventSource.addEventListener('system_update', (event: MessageEvent) => {
      try {
        const system: System = JSON.parse(event.data);
        console.log('[SSE] System update:', system.name);
        this.callbacks.onSystemUpdate?.(system);
      } catch (error) {
        console.error('[SSE] Error parsing system_update event:', error);
      }
    });

    // Evento: check_complete
    this.eventSource.addEventListener('check_complete', (event: MessageEvent) => {
      try {
        const data: SSECheckCompleteEvent = JSON.parse(event.data);
        console.log('[SSE] Check complete:', data.message);
        this.callbacks.onCheckComplete?.(data);
      } catch (error) {
        console.error('[SSE] Error parsing check_complete event:', error);
      }
    });

    // Manejo de errores
    this.eventSource.onerror = (error: Event) => {
      console.error('[SSE] Connection error:', error);
      this.callbacks.onError?.(error);

      // EventSource auto-reconecta, pero podemos agregar lógica adicional aquí
    };

    // Retorna función de cleanup
    return () => this.disconnect();
  }

  /**
   * Desconecta del stream SSE
   */
  disconnect(): void {
    if (this.eventSource) {
      console.log('[SSE] Desconectando...');
      this.eventSource.close();
      this.eventSource = null;
      this.callbacks = {};
    }
  }

  /**
   * Verifica si está conectado
   */
  isConnected(): boolean {
    return this.eventSource !== null && this.eventSource.readyState === EventSource.OPEN;
  }
}
