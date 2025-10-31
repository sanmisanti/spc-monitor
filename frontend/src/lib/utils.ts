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
 * Obtiene el nombre del ícono de Lucide según el estado
 */
export function getStatusIcon(status: SystemStatus): string {
  switch (status) {
    case 'online':
      return 'CheckCircle2';
    case 'warning':
      return 'AlertTriangle';
    case 'error':
      return 'XCircle';
    default:
      return 'HelpCircle';
  }
}

/**
 * Obtiene el nombre del ícono de Lucide según el tipo de check
 */
export function getCheckIconName(type: string): string {
  const lowerType = type.toLowerCase();

  if (lowerType.includes('http') || lowerType.includes('web')) return 'Globe';
  if (lowerType.includes('ssl') || lowerType.includes('certificate')) return 'Shield';
  if (lowerType.includes('database') || lowerType.includes('db') || lowerType.includes('mail')) return 'Database';
  if (lowerType.includes('vpn') || lowerType.includes('network')) return 'Wifi';
  if (lowerType.includes('domain') || lowerType.includes('rdap')) return 'Globe2';
  if (lowerType.includes('sheet') || lowerType.includes('spreadsheet')) return 'FileSpreadsheet';

  return 'Activity';
}

/**
 * Genera un resumen corto del check para mostrar en badge inline
 */
export function getCheckBadgeSummary(check: {
  type: string;
  name: string;
  status: string;
  response_time_ms: number;
  metadata?: Record<string, any>;
}): string {
  const lowerType = check.type.toLowerCase();

  // HTTP check: mostrar response time
  if (lowerType.includes('http')) {
    return formatResponseTime(check.response_time_ms);
  }

  // SSL check: mostrar días restantes
  if (lowerType.includes('ssl') && check.metadata?.ssl_days_remaining !== undefined) {
    return `${check.metadata.ssl_days_remaining}d`;
  }

  // Mail check: mostrar enviados/total
  if (lowerType.includes('mail') && check.metadata?.today_sent !== undefined) {
    return `${check.metadata.today_sent}/${check.metadata.today_total}`;
  }

  // VPN check: mostrar "OK" o "Down"
  if (lowerType.includes('vpn')) {
    return check.status === 'online' ? 'OK' : 'Down';
  }

  // PostgreSQL check: mostrar conteo de usuarios
  if (lowerType.includes('postgresql') && check.metadata?.user_count !== undefined) {
    return `${check.metadata.user_count} users`;
  }

  // Domain check: mostrar días hasta expiración
  if (lowerType.includes('domain') && check.metadata?.days_remaining !== undefined) {
    return `${check.metadata.days_remaining}d`;
  }

  // Google Sheets check: mostrar antigüedad
  if (lowerType.includes('sheet') && check.metadata?.days_old !== undefined) {
    return check.metadata.days_old === 0 ? 'Hoy' : `${check.metadata.days_old}d`;
  }

  // Default: solo response time
  return formatResponseTime(check.response_time_ms);
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
