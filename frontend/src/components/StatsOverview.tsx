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
      className={`rounded-lg border-2 p-4 shadow-sm transition-all hover:shadow-md ${className}`}
    >
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          {icon}
          <div>
            <p className="text-sm font-medium opacity-80">{label}</p>
            <p className="text-2xl font-bold">
              {value}
              {percentage !== undefined && (
                <span className="text-sm font-normal ml-2">
                  ({percentage}%)
                </span>
              )}
            </p>
          </div>
        </div>
      </div>
    </div>
  );
}
