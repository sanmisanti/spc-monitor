import { useState, useEffect, useCallback, useRef } from 'react';
import { SSEClient, type SSECallbacks } from '../services/sse';

interface UseSSEReturn {
  connected: boolean;
  clientId: string | null;
  connect: (callbacks: SSECallbacks) => void;
  disconnect: () => void;
}

/**
 * Hook para manejar conexión SSE con el backend
 * Auto-cleanup al desmontar componente
 */
export function useSSE(): UseSSEReturn {
  const [connected, setConnected] = useState(false);
  const [clientId, setClientId] = useState<string | null>(null);
  const clientRef = useRef<SSEClient | null>(null);
  const cleanupRef = useRef<(() => void) | null>(null);

  const connect = useCallback((callbacks: SSECallbacks) => {
    if (clientRef.current?.isConnected()) {
      console.warn('[useSSE] Ya existe una conexión activa');
      return;
    }

    // Crear cliente SSE si no existe
    if (!clientRef.current) {
      clientRef.current = new SSEClient();
    }

    // Envolver callbacks para actualizar estado local
    const wrappedCallbacks: SSECallbacks = {
      onConnected: (data) => {
        setConnected(true);
        setClientId(data.client_id);
        callbacks.onConnected?.(data);
      },
      onSystemUpdate: callbacks.onSystemUpdate,
      onCheckComplete: callbacks.onCheckComplete,
      onError: (error) => {
        setConnected(false);
        callbacks.onError?.(error);
      },
    };

    // Conectar y guardar función de cleanup
    cleanupRef.current = clientRef.current.connect(wrappedCallbacks);
  }, []);

  const disconnect = useCallback(() => {
    cleanupRef.current?.();
    clientRef.current?.disconnect();
    setConnected(false);
    setClientId(null);
  }, []);

  // Cleanup al desmontar
  useEffect(() => {
    return () => {
      disconnect();
    };
  }, [disconnect]);

  return { connected, clientId, connect, disconnect };
}
