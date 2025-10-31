import { useState, useEffect } from 'react';
import {
  ChevronDown,
  Clock,
  Zap,
  CheckCircle2,
  AlertTriangle,
  XCircle,
  HelpCircle,
  Globe,
  Globe2,
  Shield,
  Database,
  Wifi,
  FileSpreadsheet,
  Activity,
  Server,
} from 'lucide-react';
import * as Accordion from '@radix-ui/react-accordion';
import type { System } from '../types/system';
import {
  getStatusClasses,
  getStatusIcon,
  getCheckIconName,
  getCheckBadgeSummary,
  getEnvironmentBadge,
  formatRelativeTime,
  formatResponseTime,
  getAverageResponseTime,
  cn,
} from '../lib/utils';
import { DataSourceBadge } from './DataSourceBadge';

// Mapa de iconos de Lucide
const ICON_MAP: Record<string, any> = {
  CheckCircle2,
  AlertTriangle,
  XCircle,
  HelpCircle,
  Globe,
  Globe2,
  Shield,
  Database,
  Wifi,
  FileSpreadsheet,
  Activity,
  Server,
};

interface SystemCardProps {
  system: System;
  sseConnected: boolean;
  isRefreshing: boolean;
}

export function SystemCard({ system, sseConnected, isRefreshing }: SystemCardProps) {
  const statusClasses = getStatusClasses(system.status);
  const envBadge = getEnvironmentBadge(system.environment);
  const avgResponseTime = getAverageResponseTime(system.checks);

  // Estado para animaciones cuando se actualiza por SSE
  const [isUpdating, setIsUpdating] = useState(false);

  // Detectar cuando el sistema se actualiza por SSE
  useEffect(() => {
    if (system.source === 'sse') {
      setIsUpdating(true);
      const timer = setTimeout(() => setIsUpdating(false), 2000);
      return () => clearTimeout(timer);
    }
  }, [system.last_check, system.source]);

  // Obtener icono de estado
  const StatusIconComponent = ICON_MAP[getStatusIcon(system.status)] || HelpCircle;

  return (
    <Accordion.Root type="single" collapsible className="w-full">
      <Accordion.Item value="item-1" className="border-none">
        <div
          className={cn(
            'bg-white rounded-xl border-2 shadow-md transition-all hover:shadow-lg animate-fadeIn',
            statusClasses.border,
            isUpdating && 'animate-pulse-ring'
          )}
        >
          {/* Header (siempre visible) */}
          <Accordion.Header>
            <Accordion.Trigger className="w-full px-6 py-5 text-left hover:bg-gray-50/50 transition-colors group rounded-t-xl">
              <div className="flex items-start gap-4">
                {/* Icono de estado */}
                <div className={cn('mt-1', statusClasses.icon)}>
                  <StatusIconComponent className="w-8 h-8" strokeWidth={2.5} />
                </div>

                {/* Info del sistema */}
                <div className="flex-1 min-w-0">
                  {/* Título y badges */}
                  <div className="flex items-center gap-2 mb-2 flex-wrap">
                    <h3 className="text-xl font-bold text-gray-900">
                      {system.name}
                    </h3>
                    <span
                      className={cn(
                        'px-2.5 py-0.5 rounded-md text-xs font-bold border uppercase tracking-wide',
                        envBadge.className
                      )}
                    >
                      {envBadge.label}
                    </span>
                    <DataSourceBadge
                      source={system.source}
                      sseConnected={sseConnected}
                      isRefreshing={isRefreshing}
                    />
                  </div>

                  {/* Checks inline como badges */}
                  <div className="flex items-center gap-2 mb-3 flex-wrap">
                    {system.checks.map((check) => (
                      <CheckBadge key={check.id} check={check} />
                    ))}
                  </div>

                  {/* Metadata del sistema */}
                  <div className="flex items-center gap-4 text-sm text-gray-600">
                    <span className="flex items-center gap-1.5">
                      <Clock className="w-4 h-4" />
                      {formatRelativeTime(system.last_check)}
                    </span>
                    <span className="flex items-center gap-1.5">
                      <Zap className="w-4 h-4" />
                      Avg: {formatResponseTime(avgResponseTime)}
                    </span>
                  </div>
                </div>

                {/* Indicador expandir */}
                <ChevronDown className="w-5 h-5 text-gray-400 transition-transform group-data-[state=open]:rotate-180 mt-1 flex-shrink-0" />
              </div>
            </Accordion.Trigger>
          </Accordion.Header>

          {/* Contenido expandible */}
          <Accordion.Content className="overflow-hidden data-[state=open]:animate-slideDown data-[state=closed]:animate-slideUp">
            <div className="px-6 pb-4 pt-2 border-t border-gray-200/50">
              <div className="space-y-3">
                {system.checks.map((check) => (
                  <CheckItem key={check.id} check={check} />
                ))}
              </div>
            </div>
          </Accordion.Content>
        </div>
      </Accordion.Item>
    </Accordion.Root>
  );
}

// Helper component: Badge inline para checks (co-located)
interface CheckBadgeProps {
  check: {
    id: string;
    type: string;
    name: string;
    status: string;
    response_time_ms: number;
    metadata?: Record<string, any>;
  };
}

function CheckBadge({ check }: CheckBadgeProps) {
  const statusClasses = getStatusClasses(check.status as any);
  const CheckIconComponent = ICON_MAP[getCheckIconName(check.type)] || Activity;
  const summary = getCheckBadgeSummary(check);

  // Nombre corto del check (primeras 2 palabras max)
  const shortName = check.name.split(' ').slice(0, 2).join(' ');

  return (
    <div
      className={cn(
        'inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg border text-sm font-medium transition-all',
        statusClasses.bg,
        statusClasses.border,
        'hover:scale-105'
      )}
    >
      <CheckIconComponent className="w-4 h-4 flex-shrink-0" strokeWidth={2} />
      <span className={cn('font-semibold', statusClasses.text)}>{shortName}</span>
      <span className="text-gray-600 font-mono text-xs">{summary}</span>
    </div>
  );
}

// Helper component: Item expandido con detalles (co-located)
interface CheckItemProps {
  check: {
    id: string;
    type: string;
    name: string;
    status: string;
    message: string;
    response_time_ms: number;
    metadata?: Record<string, any>;
  };
}

function CheckItem({ check }: CheckItemProps) {
  const statusClasses = getStatusClasses(check.status as any);
  const CheckIconComponent = ICON_MAP[getCheckIconName(check.type)] || Activity;
  const StatusIconComponent = ICON_MAP[getStatusIcon(check.status as any)] || HelpCircle;

  return (
    <div
      className={cn(
        'rounded-lg border-2 p-4 transition-all hover:shadow-sm',
        statusClasses.bg,
        statusClasses.border
      )}
    >
      <div className="flex items-start justify-between gap-3">
        <div className="flex-1">
          {/* Header del check */}
          <div className="flex items-center gap-2.5 mb-2">
            <StatusIconComponent className={cn('w-5 h-5', statusClasses.icon)} strokeWidth={2} />
            <CheckIconComponent className="w-4 h-4 text-gray-600" strokeWidth={2} />
            <span className="font-bold text-gray-900">{check.name}</span>
            <span className="px-2 py-0.5 rounded-md text-xs font-semibold bg-gray-100 text-gray-700 uppercase tracking-wide">
              {check.type}
            </span>
          </div>

          {/* Mensaje */}
          <p className="text-sm text-gray-700 mb-3 leading-relaxed">{check.message}</p>

          {/* Metadata (si existe) */}
          {check.metadata && Object.keys(check.metadata).length > 0 && (
            <div className="mt-3 grid grid-cols-2 sm:grid-cols-3 gap-2 text-xs">
              {Object.entries(check.metadata)
                .filter(([_, value]) => value !== null && value !== undefined)
                .slice(0, 9)
                .map(([key, value]) => (
                  <div key={key} className="bg-white/60 rounded-md px-2.5 py-1.5 border border-gray-200">
                    <span className="text-gray-600 font-semibold uppercase tracking-wide text-[10px]">
                      {key.replace(/_/g, ' ')}
                    </span>
                    <div className="text-gray-900 font-bold mt-0.5">
                      {typeof value === 'boolean'
                        ? value
                          ? '✓ Sí'
                          : '✗ No'
                        : String(value)}
                    </div>
                  </div>
                ))}
            </div>
          )}
        </div>

        {/* Response time badge */}
        <div className="text-right flex-shrink-0">
          <div className="text-xs text-gray-500 font-medium mb-1">Tiempo</div>
          <div className={cn('text-lg font-bold', statusClasses.text)}>
            {formatResponseTime(check.response_time_ms)}
          </div>
        </div>
      </div>
    </div>
  );
}
