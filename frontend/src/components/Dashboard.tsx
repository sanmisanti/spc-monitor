import { XCircle } from 'lucide-react';
import { Header } from './Header';
import { StatsOverview } from './StatsOverview';
import { SystemCard } from './SystemCard';
import { ProgressBar } from './ProgressBar';
import { useSystems } from '../hooks/useSystems';
import { calculateSystemStats } from '../lib/utils';

export function Dashboard() {
  const {
    systems,
    loading,
    error,
    lastUpdate,
    sseConnected,
    refreshing,
    refreshProgress,
    refreshAll,
  } = useSystems();

  const stats = calculateSystemStats(systems);

  const handleRefresh = async () => {
    await refreshAll();
  };

  if (loading) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-gray-50 to-gray-100">
        <Header
          lastUpdate={null}
          sseConnected={false}
          onRefresh={() => {}}
          refreshing={false}
        />
        <main className="max-w-[1600px] mx-auto px-4 sm:px-6 lg:px-8 py-8">
          <LoadingSkeleton />
        </main>
      </div>
    );
  }

  if (error) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-gray-50 to-gray-100">
        <Header
          lastUpdate={lastUpdate}
          sseConnected={sseConnected}
          onRefresh={handleRefresh}
          refreshing={refreshing}
        />
        <main className="max-w-[1600px] mx-auto px-4 sm:px-6 lg:px-8 py-8">
          <div className="bg-error-50 border-2 border-error-200 rounded-xl p-8 text-center shadow-lg">
            <XCircle className="w-12 h-12 text-error-500 mx-auto mb-3" />
            <p className="text-error-700 font-bold text-xl mb-2">
              Error al cargar sistemas
            </p>
            <p className="text-error-600 text-sm">{error}</p>
          </div>
        </main>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-gray-50 to-gray-100">
      <Header
        lastUpdate={lastUpdate}
        sseConnected={sseConnected}
        onRefresh={handleRefresh}
        refreshing={refreshing}
      />

      <main className="max-w-[1600px] mx-auto px-4 sm:px-6 lg:px-8 py-8 space-y-8">
        {/* Stats overview */}
        <StatsOverview stats={stats} />

        {/* Progress bar */}
        <ProgressBar
          current={refreshProgress.updated}
          total={refreshProgress.total}
          visible={refreshing}
        />

        {/* System cards - Grid de 2 columnas en desktop */}
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          {systems.map((system) => (
            <SystemCard
              key={system.id}
              system={system}
              sseConnected={sseConnected}
              isRefreshing={refreshing}
            />
          ))}
        </div>

        {/* Empty state */}
        {systems.length === 0 && (
          <div className="bg-white border-2 border-gray-200 rounded-xl p-12 text-center shadow-lg col-span-full">
            <p className="text-gray-500 text-lg font-medium">
              No hay sistemas configurados para monitorear
            </p>
          </div>
        )}
      </main>
    </div>
  );
}

// Helper component (co-located)
function LoadingSkeleton() {
  return (
    <div className="space-y-8 animate-pulse">
      {/* Stats skeleton */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        {[1, 2, 3, 4].map((i) => (
          <div key={i} className="bg-white rounded-xl border-2 border-gray-200 p-5 h-28 shadow-md" />
        ))}
      </div>

      {/* Cards skeleton - Grid de 2 columnas */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {[1, 2, 3, 4, 5, 6].map((i) => (
          <div
            key={i}
            className="bg-white rounded-xl border-2 border-gray-200 p-6 h-40 shadow-md"
          />
        ))}
      </div>
    </div>
  );
}
