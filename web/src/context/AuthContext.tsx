import {
  createContext,
  useCallback,
  useContext,
  useMemo,
  useState,
  type ReactNode,
} from "react";
import { api, clearTokens, getStoredUser, setTokens } from "../api/client";
import type { AuthResponse } from "../api/types";

interface AuthState {
  userId: string;
  email: string;
}

interface AuthContextValue {
  user: AuthState | null;
  isAuthenticated: boolean;
  login: (email: string, password: string) => Promise<void>;
  signup: (email: string, password: string) => Promise<void>;
  logout: () => void;
}

const AuthContext = createContext<AuthContextValue | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<AuthState | null>(() => getStoredUser());

  const applyAuth = useCallback((auth: AuthResponse) => {
    setTokens(auth);
    setUser({ userId: auth.user_id, email: auth.email });
  }, []);

  const login = useCallback(
    async (email: string, password: string) => {
      const auth = await api.login(email, password);
      applyAuth(auth);
    },
    [applyAuth]
  );

  const signup = useCallback(
    async (email: string, password: string) => {
      const auth = await api.signup(email, password);
      applyAuth(auth);
    },
    [applyAuth]
  );

  const logout = useCallback(() => {
    clearTokens();
    setUser(null);
  }, []);

  const value = useMemo(
    () => ({
      user,
      isAuthenticated: !!user,
      login,
      signup,
      logout,
    }),
    [user, login, signup, logout]
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth() {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error("useAuth must be used within AuthProvider");
  return ctx;
}
