import type { ListParams, ListResult, Stats, User, UserInput, UserStatus } from '../types';

export const SERVER_UNREACHABLE = 'Unable to reach the server. Please try again later.';
export const DUPLICATE_EMAIL = 'A user with this email already exists.';

const BASE = (import.meta.env.VITE_API_BASE_URL ?? '').replace(/\/$/, '');

export class ApiError extends Error {
  constructor(
    message: string,
    public status: number,
    public details: Record<string, string> = {},
  ) {
    super(message);
  }
}

async function request<T>(path: string, init?: RequestInit): Promise<{ data: T; meta?: ListResult['meta'] }> {
  let res: Response;
  try {
    res = await fetch(`${BASE}${path}`, {
      ...init,
      headers: { 'Content-Type': 'application/json', ...(init?.headers ?? {}) },
    });
  } catch {
    throw new ApiError(SERVER_UNREACHABLE, 0);
  }
  let body: any = null;
  try {
    body = await res.json();
  } catch {
    /* non-JSON body */
  }
  if (!res.ok) {
    if (res.status === 503 || res.status >= 502) throw new ApiError(SERVER_UNREACHABLE, res.status);
    if (res.status === 409) throw new ApiError(DUPLICATE_EMAIL, 409);
    throw new ApiError(body?.error?.message ?? 'Request failed.', res.status, body?.error?.details ?? {});
  }
  return { data: body?.data as T, meta: body?.meta };
}

export async function listUsers(p: ListParams): Promise<ListResult> {
  const qs = new URLSearchParams({ page: String(p.page), pageSize: String(p.pageSize) });
  if (p.search) qs.set('search', p.search);
  if (p.status) qs.set('status', p.status);
  const { data, meta } = await request<User[]>(`/api/v1/users?${qs}`);
  return { users: data ?? [], meta: meta ?? { page: p.page, pageSize: p.pageSize, total: 0 } };
}

export const getStats = async () => (await request<Stats>('/api/v1/users/stats')).data;
export const getUser = async (id: string) => (await request<User>(`/api/v1/users/${id}`)).data;
export const createUser = async (u: UserInput) =>
  (await request<User>('/api/v1/users', { method: 'POST', body: JSON.stringify(u) })).data;
export const updateUser = async (id: string, u: UserInput) =>
  (await request<User>(`/api/v1/users/${id}`, { method: 'PUT', body: JSON.stringify(u) })).data;
export const setUserStatus = async (id: string, status: UserStatus) =>
  (await request<User>(`/api/v1/users/${id}/status`, { method: 'PATCH', body: JSON.stringify({ status }) })).data;
export const deleteUser = async (id: string) => {
  await request<unknown>(`/api/v1/users/${id}`, { method: 'DELETE' });
};
