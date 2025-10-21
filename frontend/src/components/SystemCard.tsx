import { useState, useEffect } from 'react';
import { ChevronDown, Clock, Zap } from 'lucide-react';
import * as Accordion from '@radix-ui/react-accordion';
import type { System } from '../types/system';
import {
  getStatusClasses,
  getStatusIcon,
  getEnvironmentBadge,
  formatRelativeTime,
  formatResponseTime,
  getAverageResponseTime,
  cn,
} from '../lib/utils';
import { DataSourceBadge } from './DataSourceBadge';

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
      const timer = setTimeout(() => setIsUpdating(false), 2000); // Duración de animaciones
      return () => clearTimeout(timer);
    }
  }, [system.last_check, system.source]);

  return (
    <Accordion.Root type="single" collapsible className="w-full">
      <Accordion.Item value="item-1" className="border-none">
        <div
          className={cn(
            'rounded-lg border-2 shadow-sm transition-all hover:shadow-md animate-fadeIn',
            statusClasses.bg,
            statusClasses.border,
            // Animaciones cuando se actualiza por SSE
            isUpdating && 'animate-flash animate-pulse-ring animate-fade-highlight'
          )}
        >
          {/* Header (siempre visible) */}
          <Accordion.Header>
            <Accordion.Trigger className="w-full px-6 py-4 flex items-center justify-between hover:opacity-80 transition-opacity group">
              <div className="flex items-center gap-4 flex-1">
                {/* Icono de estado */}
                <div className="text-3xl">{getStatusIcon(system.status)}</div>

                {/* Info del sistema */}
                <div className="flex-1 text-left">
                  <div className="flex items-center gap-2 mb-1 flex-wrap">
                    <h3 className={cn('text-xl font-bold', statusClasses.text)}>
                      {system.name}
                    </h3>
                    <span
                      className={cn(
                        'px-2 py-0.5 rounded-md text-xs font-semibold border',
                        envBadge.className
                      )}
                    >
                      {envBadge.label}
                    </span>
                    {/* Badge de origen de datos (Cache/En vivo/Actualizando) */}
                    <DataSourceBadge
                      source={system.source}
                      sseConnected={sseConnected}
                      isRefreshing={isRefreshing}
                    />
                  </div>
                  <div className="flex items-center gap-4 text-sm text-gray-600">
                    <span className="flex items-center gap-1">
                      <Clock className="w-4 h-4" />
                      {formatRelativeTime(system.last_check)}
                    </span>
                    <span className="flex items-center gap-1">
                      <Zap className="w-4 h-4" />
                      {formatResponseTime(avgResponseTime)}
                    </span>
                    <span>
                      {system.checks.length} check{system.checks.length !== 1 ? 's' : ''}
                    </span>
                  </div>
                </div>

                {/* Indicador expandir */}
                <ChevronDown className="w-5 h-5 text-gray-500 transition-transform group-data-[state=open]:rotate-180" />
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

// Helper component (co-located)
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

  return (
    <div
      className={cn(
        'rounded-md border p-3 transition-all',
        statusClasses.bg,
        'border-gray-200'
      )}
    >
      <div className="flex items-start justify-between gap-3">
        <div className="flex-1">
          <div className="flex items-center gap-2 mb-1">
            <span className="text-lg">{getStatusIcon(check.status as any)}</span>
            <span className="font-semibold text-gray-900">{check.name}</span>
            <span className="px-2 py-0.5 rounded text-xs font-medium bg-gray-100 text-gray-700">
              {check.type}
            </span>
          </div>
          <p className="text-sm text-gray-700 mb-2">{check.message}</p>

          {/* Metadata (si existe) */}
          {check.metadata && Object.keys(check.metadata).length > 0 && (
            <div className="mt-2 grid grid-cols-2 sm:grid-cols-3 gap-2 text-xs">
              {Object.entries(check.metadata)
                .filter(([_, value]) => value !== null && value !== undefined)
                .slice(0, 6)
                .map(([key, value]) => (
                  <div key={key} className="bg-white/50 rounded px-2 py-1">
                    <span className="text-gray-500 font-medium">
                      {key.replace(/_/g, ' ')}:
                    </span>{' '}
                    <span className="text-gray-900 font-semibold">
                      {typeof value === 'boolean'
                        ? value
                          ? '✓'
                          : '✗'
                        : String(value)}
                    </span>
                  </div>
                ))}
            </div>
          )}
        </div>

        {/* Response time badge */}
        <div className="text-right">
          <div className="text-xs text-gray-500">Tiempo</div>
          <div className="text-sm font-bold text-gray-900">
            {formatResponseTime(check.response_time_ms)}
          </div>
        </div>
      </div>
    </div>
  );
}
