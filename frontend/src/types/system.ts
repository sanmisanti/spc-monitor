// Tipos basados en los modelos del backend Go

export type SystemStatus = 'online' | 'warning' | 'error' | 'unknown';
export type Environment = 'prod' | 'preprod' | 'shared';
export type CheckType = 'http' | 'database' | 'rdap' | 'google-sheets';

export interface Check {
  id: string;
  type: CheckType;
  name: string;
  status: SystemStatus;
  message: string;
  last_check: string; // ISO date string
  response_time_ms: number;
  metadata?: Record<string, any>;
}

export interface System {
  id: string;
  name: string;
  type: string;
  environment: Environment;
  status: SystemStatus;
  last_check: string; // ISO date string
  checks: Check[];
  // Campos para rastrear origen de datos (frontend only)
  source?: 'cache' | 'sse';
  localUpdatedAt?: Date;
}

export interface SystemsResponse {
  systems: System[];
  cached: boolean;
  count: number;
}

export interface RefreshResponse {
  message: string;
  status: string;
  system_id?: string;
}

// SSE Event types
export interface SSEConnectedEvent {
  client_id: string;
}

export interface SSECheckCompleteEvent {
  message: string;
}

// Tipos auxiliares para UI
export interface SystemStats {
  total: number;
  online: number;
  warning: number;
  error: number;
  onlinePercentage: number;
  warningPercentage: number;
  errorPercentage: number;
}

export interface CheckMetadata {
  [key: string]: string | number | boolean | string[] | undefined;
}
