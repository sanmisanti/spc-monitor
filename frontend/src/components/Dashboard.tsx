import { useState } from 'react';
import { Header } from './Header';
import { StatsOverview } from './StatsOverview';
import { SystemCard } from './SystemCard';
import { useSystems } from '../hooks/useSystems';
import { calculateSystemStats } from '../lib/utils';

export function Dashboard() {
  const {
    systems,
    loading,
    error,
    lastUpdate,
    sseConnected,
    refreshAll,
  } = useSystems();

  const [refreshing, setRefreshing] = useState(false);

  const stats = calculateSystemStats(systems);

  const handleRefresh = async () => {
    setRefreshing(true);
    try {
      await refreshAll();
    } finally {
      // Mantener spinner por mínimo 1 segundo para feedback visual
      setTimeout(() => setRefreshing(false), 1000);
    }
  };

  if (loading) {
    return (
      <div className="min-h-screen bg-gray-50">
        <Header
          lastUpdate={null}
          sseConnected={false}
          onRefresh={() => {}}
          refreshing={false}
        />
        <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
          <LoadingSkeleton />
        </main>
      </div>
    );
  }

  if (error) {
    return (
      <div className="min-h-screen bg-gray-50">
        <Header
          lastUpdate={lastUpdate}
          sseConnected={sseConnected}
          onRefresh={handleRefresh}
          refreshing={refreshing}
        />
        <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
          <div className="bg-error-50 border-2 border-error-200 rounded-lg p-6 text-center">
            <p className="text-error-700 font-semibold text-lg mb-2">
              Error al cargar sistemas
            </p>
            <p className="text-error-600 text-sm">{error}</p>
          </div>
        </main>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50">
      <Header
        lastUpdate={lastUpdate}
        sseConnected={sseConnected}
        onRefresh={handleRefresh}
        refreshing={refreshing}
      />

      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8 space-y-6">
        {/* Stats overview */}
        <StatsOverview stats={stats} />

        {/* System cards */}
        <div className="space-y-4">
          {systems.map((system) => (
            <SystemCard key={system.id} system={system} />
          ))}
        </div>

        {/* Empty state */}
        {systems.length === 0 && (
          <div className="bg-white border-2 border-gray-200 rounded-lg p-12 text-center">
            <p className="text-gray-500 text-lg">
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
    <div className="space-y-6 animate-pulse">
      {/* Stats skeleton */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        {[1, 2, 3, 4].map((i) => (
          <div key={i} className="bg-white rounded-lg border-2 border-gray-200 p-4 h-24" />
        ))}
      </div>

      {/* Cards skeleton */}
      <div className="space-y-4">
        {[1, 2, 3].map((i) => (
          <div
            key={i}
            className="bg-white rounded-lg border-2 border-gray-200 p-6 h-32"
          />
        ))}
      </div>
    </div>
  );
}
