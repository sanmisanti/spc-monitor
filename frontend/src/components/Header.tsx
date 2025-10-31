import { RefreshCw, Activity } from 'lucide-react';
import { formatRelativeTime } from '../lib/utils';

interface HeaderProps {
  lastUpdate: Date | null;
  sseConnected: boolean;
  onRefresh: () => void;
  refreshing: boolean;
}

export function Header({ lastUpdate, sseConnected, onRefresh, refreshing }: HeaderProps) {
  return (
    <header className="bg-white border-b border-gray-200 shadow-sm">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-4">
        <div className="flex items-center justify-between">
          {/* Logo y título */}
          <div className="flex items-center gap-3">
            <div className="w-12 h-12 bg-gradient-to-br from-primary-500 to-primary-700 rounded-xl flex items-center justify-center shadow-lg">
              <Activity className="w-7 h-7 text-white" />
            </div>
            <div>
              <h1 className="text-2xl font-extrabold text-gray-900">
                Monitor de Sistemas
              </h1>
              <p className="text-sm text-gray-600 font-medium">
                Secretaría de Contrataciones
              </p>
            </div>
          </div>

          {/* Acciones */}
          <div className="flex items-center gap-4">
            {/* Estado SSE */}
            <div className="flex items-center gap-2 text-sm">
              <div
                className={`w-2 h-2 rounded-full ${
                  sseConnected ? 'bg-success-500 animate-pulse' : 'bg-gray-400'
                }`}
              />
              <span className="text-gray-600">
                {sseConnected ? 'Conectado' : 'Desconectado'}
              </span>
            </div>

            {/* Última actualización */}
            {lastUpdate && (
              <div className="text-sm text-gray-600">
                Actualizado {formatRelativeTime(lastUpdate.toISOString())}
              </div>
            )}

            {/* Botón refresh */}
            <button
              onClick={onRefresh}
              disabled={refreshing}
              className="flex items-center gap-2 px-5 py-2.5 bg-primary-500 text-white font-semibold rounded-lg hover:bg-primary-600 disabled:opacity-50 disabled:cursor-not-allowed transition-colors shadow-md hover:shadow-lg"
            >
              <RefreshCw
                className={`w-4 h-4 ${refreshing ? 'animate-spin' : ''}`}
              />
              <span>Refrescar</span>
            </button>
          </div>
        </div>
      </div>
    </header>
  );
}
