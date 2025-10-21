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
  refreshing: boolean;
  refreshProgress: { updated: number; total: number };
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
  const [refreshing, setRefreshing] = useState(false);
  const [refreshProgress, setRefreshProgress] = useState({ updated: 0, total: 0 });
  const [autoRefreshDone, setAutoRefreshDone] = useState(false);

  const { connected: sseConnected, connect: connectSSE } = useSSE();

  // Actualizar un sistema específico (desde SSE)
  const updateSystem = useCallback((updatedSystem: System) => {
    setSystems((prev) =>
      prev.map((sys) =>
        sys.id === updatedSystem.id
          ? {
              ...updatedSystem,
              source: 'sse' as const,
              localUpdatedAt: new Date(),
            }
          : sys
      )
    );
    setLastUpdate(new Date());

    // Incrementar progreso si hay refresh en curso
    setRefreshProgress((prev) => {
      if (prev.total > 0 && prev.updated < prev.total) {
        const newUpdated = prev.updated + 1;
        // Si completamos todos, resetear refreshing
        if (newUpdated === prev.total) {
          setTimeout(() => {
            setRefreshing(false);
            setRefreshProgress({ updated: 0, total: 0 });
          }, 500); // Pequeño delay para que se vea el 100%
        }
        return { ...prev, updated: newUpdated };
      }
      return prev;
    });
  }, []);

  // Carga inicial: obtener cache
  useEffect(() => {
    const loadInitialData = async () => {
      try {
        setLoading(true);
        setError(null);
        const response = await getSystems();
        // Marcar todos los sistemas como 'cache' en carga inicial
        const systemsWithSource = response.systems.map((sys) => ({
          ...sys,
          source: 'cache' as const,
          localUpdatedAt: new Date(),
        }));
        setSystems(systemsWithSource);
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

  // Refresh de todos los sistemas (manual o automático)
  const refreshAll = useCallback(async () => {
    try {
      setError(null);
      setRefreshing(true);
      setRefreshProgress({ updated: 0, total: systems.length });
      await refreshAllSystems();
      // No actualizamos estado aquí, los updates llegarán por SSE
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Error al refrescar');
      setRefreshing(false);
      setRefreshProgress({ updated: 0, total: 0 });
    }
  }, [systems.length]);

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

  // Auto-refresh al conectar SSE (solo una vez)
  useEffect(() => {
    if (sseConnected && !autoRefreshDone && systems.length > 0) {
      console.log('[useSystems] SSE conectado, disparando auto-refresh...');
      setAutoRefreshDone(true);
      refreshAll();
    }
  }, [sseConnected, autoRefreshDone, systems.length, refreshAll]);

  return {
    systems,
    loading,
    error,
    cached,
    lastUpdate,
    sseConnected,
    refreshing,
    refreshProgress,
    refreshAll,
  };
}
