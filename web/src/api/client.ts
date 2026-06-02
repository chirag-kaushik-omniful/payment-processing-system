import type { ApiError, AuditEvent, AuthResponse, Payment, PaymentMethod, Wallet } from "./types";

const API_BASE = import.meta.env.VITE_API_BASE ?? "/api";

function getToken(): string | null {
  return localStorage.getItem("access_token");
}

export function setTokens(auth: AuthResponse) {
  localStorage.setItem("access_token", auth.access_token);
  localStorage.setItem("refresh_token", auth.refresh_token);
  localStorage.setItem("user_id", auth.user_id);
  localStorage.setItem("email", auth.email);
}

export function clearTokens() {
  localStorage.removeItem("access_token");
  localStorage.removeItem("refresh_token");
  localStorage.removeItem("user_id");
  localStorage.removeItem("email");
}

export function getStoredUser(): { userId: string; email: string } | null {
  const userId = localStorage.getItem("user_id");
  const email = localStorage.getItem("email");
  if (!userId || !email) return null;
  return { userId, email };
}

async function request<T>(
  path: string,
  options: RequestInit = {},
  idempotencyKey?: string
): Promise<T> {
  const headers: Record<string, string> = {
    "Content-Type": "application/json",
    ...(options.headers as Record<string, string>),
  };
  const token = getToken();
  if (token) headers.Authorization = `Bearer ${token}`;
  if (idempotencyKey) headers["Idempotency-Key"] = idempotencyKey;

  const res = await fetch(`${API_BASE}${path}`, { ...options, headers });
  const text = await res.text();
  let body: unknown = null;
  if (text) {
    try {
      body = JSON.parse(text);
    } catch {
      body = { message: text };
    }
  }

  if (!res.ok) {
    const err = body as ApiError;
    throw new Error(err.message || `Request failed (${res.status})`);
  }
  return body as T;
}

export const api = {
  health: () => request<{ status: string }>("/health"),

  signup: (email: string, password: string) =>
    request<AuthResponse>("/auth/signup", {
      method: "POST",
      body: JSON.stringify({ email, password }),
    }),

  login: (email: string, password: string) =>
    request<AuthResponse>("/auth/login", {
      method: "POST",
      body: JSON.stringify({ email, password }),
    }),

  refresh: (refreshToken: string) =>
    request<AuthResponse>("/auth/refresh", {
      method: "POST",
      body: JSON.stringify({ refresh_token: refreshToken }),
    }),

  listPayments: () => request<Payment[]>("/payments"),

  getPayment: (id: string) => request<Payment>(`/payments/${id}`),

  createPayment: (
    data: {
      amount: number;
      currency: string;
      provider?: string;
      payment_method_id?: string;
      capture_mode?: string;
    },
    idempotencyKey?: string
  ) =>
    request<Payment>("/payments", {
      method: "POST",
      body: JSON.stringify(data),
    }, idempotencyKey),

  capturePayment: (id: string) =>
    request<Payment>(`/payments/${id}/capture`, { method: "POST" }),

  createPaymentMethod: (data: { provider: string; last_four: string; brand: string }) =>
    request<PaymentMethod>("/payment-methods", {
      method: "POST",
      body: JSON.stringify(data),
    }),

  getWallet: () => request<Wallet>("/wallet"),

  createRefund: (paymentId: string, amount: number) =>
    request<Payment>("/refunds", {
      method: "POST",
      body: JSON.stringify({ payment_id: paymentId, amount }),
    }),

  adminListPayments: async () => {
    const res = await request<{ payments: Payment[] }>("/admin/payments");
    return res.payments ?? [];
  },

  adminAudit: async () => {
    const res = await request<{ events: AuditEvent[] }>("/admin/audit");
    return res.events ?? [];
  },

  adminReplaySaga: (paymentId: string) =>
    request<{ status: string }>(`/admin/payments/${paymentId}/replay-saga`, {
      method: "POST",
    }),
};
