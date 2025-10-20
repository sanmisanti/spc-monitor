import { type ClassValue, clsx } from "clsx";
import { twMerge } from "tailwind-merge";
import type { SystemStatus, System, SystemStats } from "../types/system";

/**
 * Utility para combinar clases de Tailwind con soporte para condicionales
 */
export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

/**
 * Obtiene las clases de color según el estado del sistema
 */
export function getStatusClasses(status: SystemStatus) {
  switch (status) {
    case 'online':
      return {
        bg: 'bg-success-50',
        border: 'border-success-200',
        text: 'text-success-700',
        badge: 'bg-success-100 text-success-800 border-success-300',
        icon: 'text-success-500',
      };
    case 'warning':
      return {
        bg: 'bg-warning-50',
        border: 'border-warning-200',
        text: 'text-warning-700',
        badge: 'bg-warning-100 text-warning-800 border-warning-300',
        icon: 'text-warning-500',
      };
    case 'error':
      return {
        bg: 'bg-error-50',
        border: 'border-error-200',
        text: 'text-error-700',
        badge: 'bg-error-100 text-error-800 border-error-300',
        icon: 'text-error-500',
      };
    default:
      return {
        bg: 'bg-gray-50',
        border: 'border-gray-200',
        text: 'text-gray-700',
        badge: 'bg-gray-100 text-gray-800 border-gray-300',
        icon: 'text-gray-500',
      };
  }
}

/**
 * Obtiene el ícono según el estado
 */
export function getStatusIcon(status: SystemStatus): string {
  switch (status) {
    case 'online':
      return '✅';
    case 'warning':
      return '⚠️';
    case 'error':
      return '❌';
    default:
      return '❓';
  }
}

/**
 * Calcula estadísticas agregadas de sistemas
 */
export function calculateSystemStats(systems: System[]): SystemStats {
  const total = systems.length;
  const online = systems.filter(s => s.status === 'online').length;
  const warning = systems.filter(s => s.status === 'warning').length;
  const error = systems.filter(s => s.status === 'error').length;

  return {
    total,
    online,
    warning,
    error,
    onlinePercentage: total > 0 ? Math.round((online / total) * 100) : 0,
    warningPercentage: total > 0 ? Math.round((warning / total) * 100) : 0,
    errorPercentage: total > 0 ? Math.round((error / total) * 100) : 0,
  };
}

/**
 * Formatea un timestamp relativo (e.g., "hace 2 minutos")
 */
export function formatRelativeTime(dateString: string): string {
  const date = new Date(dateString);
  const now = new Date();
  const diffMs = now.getTime() - date.getTime();
  const diffSecs = Math.floor(diffMs / 1000);
  const diffMins = Math.floor(diffSecs / 60);
  const diffHours = Math.floor(diffMins / 60);
  const diffDays = Math.floor(diffHours / 24);

  if (diffSecs < 60) {
    return `hace ${diffSecs} seg`;
  } else if (diffMins < 60) {
    return `hace ${diffMins} min`;
  } else if (diffHours < 24) {
    return `hace ${diffHours}h`;
  } else {
    return `hace ${diffDays}d`;
  }
}

/**
 * Formatea response time (e.g., "127ms")
 */
export function formatResponseTime(ms: number): string {
  if (ms >= 1000) {
    return `${(ms / 1000).toFixed(2)}s`;
  }
  return `${Math.round(ms)}ms`;
}

/**
 * Obtiene el badge de ambiente según environment
 */
export function getEnvironmentBadge(env: string): { label: string; className: string } {
  switch (env) {
    case 'prod':
      return {
        label: 'PROD',
        className: 'bg-red-100 text-red-800 border-red-300',
      };
    case 'preprod':
      return {
        label: 'PREPROD',
        className: 'bg-yellow-100 text-yellow-800 border-yellow-300',
      };
    case 'shared':
      return {
        label: 'SHARED',
        className: 'bg-gray-100 text-gray-800 border-gray-300',
      };
    default:
      return {
        label: env.toUpperCase(),
        className: 'bg-blue-100 text-blue-800 border-blue-300',
      };
  }
}

/**
 * Calcula el promedio de response times de los checks
 */
export function getAverageResponseTime(checks: Array<{ response_time_ms: number }>): number {
  if (checks.length === 0) return 0;
  const sum = checks.reduce((acc, check) => acc + check.response_time_ms, 0);
  return Math.round(sum / checks.length);
}
