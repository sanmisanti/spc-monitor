import { RefreshCw } from 'lucide-react';
import { cn } from '../lib/utils';

interface ProgressBarProps {
  current: number;
  total: number;
  visible: boolean;
}

/**
 * Barra de progreso para mostrar actualización de sistemas
 */
export function ProgressBar({ current, total, visible }: ProgressBarProps) {
  const percentage = total > 0 ? Math.round((current / total) * 100) : 0;
  const isComplete = current === total && total > 0;

  if (!visible || total === 0) {
    return null;
  }

  return (
    <div
      className={cn(
        'bg-white border-2 border-blue-200 rounded-lg p-4 shadow-sm transition-all',
        isComplete ? 'animate-fade-highlight' : 'animate-fadeIn'
      )}
    >
      <div className="flex items-center justify-between mb-2">
        <div className="flex items-center gap-2">
          <RefreshCw
            className={cn(
              'w-4 h-4 text-blue-600',
              !isComplete && 'animate-spin'
            )}
          />
          <span className="text-sm font-semibold text-gray-700">
            {isComplete ? 'Actualización completa' : 'Actualizando sistemas'}
          </span>
        </div>
        <div className="text-sm font-bold text-blue-600">
          {current}/{total}
        </div>
      </div>

      {/* Barra de progreso */}
      <div className="w-full bg-gray-200 rounded-full h-2 overflow-hidden">
        <div
          className={cn(
            'h-full rounded-full transition-all duration-500 ease-out',
            isComplete ? 'bg-success-500' : 'bg-blue-500'
          )}
          style={{ width: `${percentage}%` }}
        />
      </div>

      {/* Porcentaje */}
      <div className="text-right mt-1">
        <span className="text-xs font-medium text-gray-500">
          {percentage}%
        </span>
      </div>
    </div>
  );
}
