import { Database, Radio, RefreshCw } from 'lucide-react';
import { cn } from '../lib/utils';

interface DataSourceBadgeProps {
  source?: 'cache' | 'sse';
  sseConnected: boolean;
  isRefreshing?: boolean;
}

/**
 * Badge que indica si los datos vienen del cache, SSE (en vivo), o están actualizándose
 */
export function DataSourceBadge({ source, sseConnected, isRefreshing }: DataSourceBadgeProps) {
  const isLive = source === 'sse' && sseConnected;

  // Prioridad 1: Si está actualizándose (solo si aún no es 'sse')
  if (isRefreshing && source !== 'sse') {
    return (
      <div
        className={cn(
          'inline-flex items-center gap-1.5 px-2 py-0.5 rounded-md text-xs font-medium border',
          'bg-warning-100 text-warning-800 border-warning-300'
        )}
      >
        <RefreshCw className="w-3 h-3 animate-spin" />
        <span>Actualizando...</span>
      </div>
    );
  }

  // Prioridad 2: En vivo (ya actualizado por SSE)
  if (isLive) {
    return (
      <div
        className={cn(
          'inline-flex items-center gap-1.5 px-2 py-0.5 rounded-md text-xs font-medium border',
          'bg-success-100 text-success-800 border-success-300 animate-pulse-subtle'
        )}
      >
        <Radio className="w-3 h-3 animate-pulse" />
        <span>En vivo</span>
      </div>
    );
  }

  // Prioridad 3: Cache
  if (source === 'cache') {
    return (
      <div
        className={cn(
          'inline-flex items-center gap-1.5 px-2 py-0.5 rounded-md text-xs font-medium border',
          'bg-gray-100 text-gray-700 border-gray-300'
        )}
      >
        <Database className="w-3 h-3" />
        <span>Cache</span>
      </div>
    );
  }

  // Fallback: si no hay source definido
  return null;
}
