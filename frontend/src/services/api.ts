import type { SystemsResponse, RefreshResponse } from '../types/system';

const API_BASE = 'http://localhost:8080/api';

/**
 * Obtiene todos los sistemas desde el cache
 */
export async function getSystems(): Promise<SystemsResponse> {
  const response = await fetch(`${API_BASE}/systems`);

  if (!response.ok) {
    throw new Error(`Failed to fetch systems: ${response.statusText}`);
  }

  return response.json();
}

/**
 * Dispara refresh de todos los sistemas
 * Retorna inmediatamente (202 Accepted), los updates llegan por SSE
 */
export async function refreshAllSystems(): Promise<RefreshResponse> {
  const response = await fetch(`${API_BASE}/refresh`, {
    method: 'POST',
  });

  if (!response.ok) {
    throw new Error(`Failed to refresh systems: ${response.statusText}`);
  }

  return response.json();
}

/**
 * Dispara refresh de un sistema específico
 * Retorna inmediatamente (202 Accepted), el update llega por SSE
 */
export async function refreshSystem(systemId: string): Promise<RefreshResponse> {
  const response = await fetch(`${API_BASE}/systems/${systemId}/refresh`, {
    method: 'POST',
  });

  if (!response.ok) {
    throw new Error(`Failed to refresh system ${systemId}: ${response.statusText}`);
  }

  return response.json();
}
