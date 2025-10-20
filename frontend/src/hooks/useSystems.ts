import { useState, useEffect, useCallback } from 'react';
import type { System } from '../types/system';
import { getSystems, refreshAllSystems } from '../services/api';
import { useSSE } from './useSSE';

interface UseSystemsReturn {
  systems: System[];
  loading: boolean;
  error: string | null;
  cached: boolean;
  lastUpdate: Date | null;
  sseConnected: boolean;
  refreshAll: () => Promise<void>;
}

/**
 * Hook principal para manejar estado de sistemas
 * Implementa estrategia híbrida: cache primero, SSE después
 */
export function useSystems(): UseSystemsReturn {
  const [systems, setSystems] = useState<System[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [cached, setCached] = useState(false);
  const [lastUpdate, setLastUpdate] = useState<Date | null>(null);

  const { connected: sseConnected, connect: connectSSE } = useSSE();

  // Actualizar un sistema específico (desde SSE)
  const updateSystem = useCallback((updatedSystem: System) => {
    setSystems((prev) =>
      prev.map((sys) => (sys.id === updatedSystem.id ? updatedSystem : sys))
    );
    setLastUpdate(new Date());
  }, []);

  // Carga inicial: obtener cache
  useEffect(() => {
    const loadInitialData = async () => {
      try {
        setLoading(true);
        setError(null);
        const response = await getSystems();
        setSystems(response.systems);
        setCached(response.cached);
        setLastUpdate(new Date());
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Error desconocido');
      } finally {
        setLoading(false);
      }
    };

    loadInitialData();
  }, []);

  // Conectar SSE después de carga inicial
  useEffect(() => {
    if (!loading && systems.length > 0) {
      connectSSE({
        onSystemUpdate: updateSystem,
        onCheckComplete: (data) => {
          console.log('[useSystems] Checks completos:', data.message);
        },
        onError: (error) => {
          console.error('[useSystems] Error SSE:', error);
        },
      });
    }
  }, [loading, systems.length, connectSSE, updateSystem]);

  // Refresh manual de todos los sistemas
  const refreshAll = useCallback(async () => {
    try {
      setError(null);
      await refreshAllSystems();
      // No actualizamos estado aquí, los updates llegarán por SSE
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Error al refrescar');
    }
  }, []);

  return {
    systems,
    loading,
    error,
    cached,
    lastUpdate,
    sseConnected,
    refreshAll,
  };
}
