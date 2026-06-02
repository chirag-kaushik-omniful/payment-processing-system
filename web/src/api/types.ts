export interface AuthResponse {
  access_token: string;
  refresh_token: string;
  expires_in: number;
  user_id: string;
  email: string;
}

export interface Payment {
  id: string;
  user_id: string;
  amount: number;
  currency: string;
  status: string;
  provider?: string;
}

export interface PaymentMethod {
  id: string;
  provider: string;
  last_four: string;
  brand: string;
}

export interface Wallet {
  user_id: string;
  balance: number;
  version: number;
  currency?: string;
}

export interface AuditEvent {
  id: string;
  actor_id?: string;
  action: string;
  resource_type: string;
  resource_id: string;
  created_at?: string;
}

export interface ApiError {
  code?: string;
  message: string;
}
