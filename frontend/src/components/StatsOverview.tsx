import { Server, CheckCircle, AlertTriangle, XCircle } from 'lucide-react';
import type { SystemStats } from '../types/system';

interface StatsOverviewProps {
  stats: SystemStats;
}

export function StatsOverview({ stats }: StatsOverviewProps) {
  return (
    <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
      <StatCard
        icon={<Server className="w-5 h-5" />}
        label="Total Sistemas"
        value={stats.total}
        className="bg-gray-50 border-gray-200 text-gray-700"
      />
      <StatCard
        icon={<CheckCircle className="w-5 h-5" />}
        label="Online"
        value={stats.online}
        percentage={stats.onlinePercentage}
        className="bg-success-50 border-success-200 text-success-700"
      />
      <StatCard
        icon={<AlertTriangle className="w-5 h-5" />}
        label="Warnings"
        value={stats.warning}
        percentage={stats.warningPercentage}
        className="bg-warning-50 border-warning-200 text-warning-700"
      />
      <StatCard
        icon={<XCircle className="w-5 h-5" />}
        label="Errores"
        value={stats.error}
        percentage={stats.errorPercentage}
        className="bg-error-50 border-error-200 text-error-700"
      />
    </div>
  );
}

// Helper component (co-located)
interface StatCardProps {
  icon: React.ReactNode;
  label: string;
  value: number;
  percentage?: number;
  className: string;
}

function StatCard({ icon, label, value, percentage, className }: StatCardProps) {
  return (
    <div
      className={`rounded-xl border-2 p-5 shadow-md transition-all hover:shadow-lg hover:scale-[1.02] ${className}`}
    >
      <div className="flex items-start gap-4">
        <div className="p-3 rounded-lg bg-white/60 backdrop-blur-sm">
          {icon}
        </div>
        <div className="flex-1">
          <p className="text-sm font-semibold opacity-80 uppercase tracking-wide mb-1">{label}</p>
          <p className="text-3xl font-extrabold">
            {value}
          </p>
          {percentage !== undefined && (
            <p className="text-sm font-medium opacity-70 mt-1">
              {percentage}% del total
            </p>
          )}
        </div>
      </div>
    </div>
  );
}
