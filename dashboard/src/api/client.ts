const API_BASE = '/v1/api';

interface RequestOptions {
  method?: string;
  body?: unknown;
  headers?: Record<string, string>;
}

interface ApiResponse<T> {
  data: T;
  message?: string;
  meta?: { total: number; page: number; page_size: number };
}

class ApiError extends Error {
  status: number;
  constructor(message: string, status: number) {
    super(message);
    this.status = status;
  }
}

async function request<T>(path: string, opts: RequestOptions = {}): Promise<ApiResponse<T>> {
  const { method = 'GET', body, headers = {} } = opts;

  const res = await fetch(`${API_BASE}${path}`, {
    method,
    credentials: 'include', // httpOnly cookie
    headers: {
      'Content-Type': 'application/json',
      ...headers,
    },
    body: body ? JSON.stringify(body) : undefined,
  });

  if (!res.ok) {
    const err = await res.json().catch(() => ({ message: res.statusText }));
    throw new ApiError(err.message || res.statusText, res.status);
  }

  return res.json();
}

// Auth
export const auth = {
  login: (email: string, password: string) =>
    request<{ access_token: string; refresh_token: string }>('/auth/login', { method: 'POST', body: { email, password } }),
  register: (email: string, password: string, full_name: string) =>
    request<{ access_token: string }>('/auth/register', { method: 'POST', body: { email, password, full_name } }),
  refresh: (refresh_token: string) =>
    request<{ access_token: string }>('/auth/refresh', { method: 'POST', body: { refresh_token } }),
  logout: () => request('/auth/logout', { method: 'POST' }),
  me: () => request<User>('/users/me'),
};

// Tenants
export const tenants = {
  myTenants: () => request<Tenant[]>('/tenants/me'),
  get: (id: string) => request<Tenant>(`/tenants/${id}`),
  update: (id: string, data: Partial<Tenant>) =>
    request<Tenant>(`/tenants/${id}`, { method: 'PUT', body: data }),
  create: (name: string, slug: string) =>
    request<Tenant>('/tenants', { method: 'POST', body: { name, slug } }),
};

// Members
export const members = {
  list: (tenantId: string) => request<TenantMember[]>(`/tenants/${tenantId}/members`),
  invite: (tenantId: string, email: string, role_id?: string) =>
    request(`/tenants/${tenantId}/invitations`, { method: 'POST', body: { email, role_id } }),
};

// Roles
export const roles = {
  list: () => request<Role[]>('/roles'),
};

// Apps
export const apps = {
  list: (tenantId: string) => request<App[]>(`/tenants/${tenantId}/apps`),
  get: (appId: string) => request<App>(`/apps/${appId}`),
  create: (tenantId: string, name: string) =>
    request<App>(`/tenants/${tenantId}/apps`, { method: 'POST', body: { name } }),
  update: (appId: string, data: Partial<App>) =>
    request<App>(`/apps/${appId}`, { method: 'PUT', body: data }),
  delete: (appId: string) =>
    request(`/apps/${appId}`, { method: 'DELETE' }),
  metrics: (appId: string) => request<AppMetrics>(`/apps/${appId}/metrics`),
  timeseries: (appId: string) => request<TimeSeriesPoint[]>(`/apps/${appId}/metrics/timeseries`),
  environments: (appId: string) => request<Environment[]>(`/apps/${appId}/environments`),
  apiKeys: (envId: string) => request<ApiKey[]>(`/environments/${envId}/api-keys`),
  webhooks: (appId: string) => request<Webhook[]>(`/apps/${appId}/webhooks`),
};

// Pricing Plans
export const plans = {
  list: () => request<PricingPlan[]>('/pricing-plans'),
};

// GDPR
export const gdpr = {
  export: (tenantId: string, userId: string, categories?: string[]) =>
    request(`/tenants/${tenantId}/gdpr/export`, { method: 'POST', body: { user_id: userId, categories } }),
  erase: (tenantId: string, userId: string, categories?: string[]) =>
    request(`/tenants/${tenantId}/gdpr/erase`, { method: 'POST', body: { user_id: userId, categories } }),
};

// Types
export interface User {
  id: string;
  email: string;
  full_name?: string;
  is_root_admin: boolean;
}

export interface Tenant {
  id: string;
  name: string;
  slug: string;
  status: string;
  pricing_plan?: PricingPlan;
}

export interface TenantMember {
  id: string;
  user_id: string;
  tenant_id: string;
  role?: Role;
  user?: User;
  created_at: string;
}

export interface Role {
  id: string;
  name: string;
  slug: string;
  permissions: Record<string, unknown>;
}

export interface App {
  id: string;
  name: string;
  tenant_id: string;
  status: string;
  created_at: string;
}

export interface AppMetrics {
  total_messages: number;
  total_threads: number;
  total_subscribers: number;
  messages_today: number;
}

export interface TimeSeriesPoint {
  date: string;
  count: number;
}

export interface Environment {
  id: string;
  name: string;
  slug: string;
  app_id: string;
  rate_limit_rpm: number;
  rate_limit_daily: number;
}

export interface ApiKey {
  id: string;
  key_prefix: string;
  environment_id: string;
  is_active: boolean;
  created_at: string;
}

export interface Webhook {
  id: string;
  url: string;
  events: string[];
  is_active: boolean;
}

export interface PricingPlan {
  id: string;
  name: string;
  slug: string;
  price: number;
  currency: string;
  monthly_credits: number;
  max_apps: number;
  max_members: number;
  max_workflows: number;
  max_messages_per_month: number;
  rate_limit_rpm: number;
  is_active: boolean;
}
