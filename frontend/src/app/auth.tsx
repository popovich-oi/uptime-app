"use client";

import { createContext, useCallback, useContext, useEffect, useState } from "react";
import { api, setAccessToken as setApiAccessToken } from "./api";

export type User = { id: number; username: string; email: string; avatar_url: string };
type AuthResponse = { access_token: string; user: User };
type AuthContextValue = { user: User | null; accessToken: string | null; loading: boolean; error: string | null; register: (username: string, email: string, password: string) => Promise<boolean>; login: (email: string, password: string) => Promise<boolean>; logout: () => Promise<void>; updateProfile: (username: string, avatar?: File) => Promise<boolean>; clearError: () => void };
const AuthContext = createContext<AuthContextValue | null>(null);

export function AuthProvider({ children }: Readonly<{ children: React.ReactNode }>) {
  const [user, setUser] = useState<User | null>(null);
  const [accessToken, setAccessToken] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const applyResponse = (response: AuthResponse | null) => { setApiAccessToken(response?.access_token ?? null); setAccessToken(response?.access_token ?? null); setUser(response?.user ?? null); };
  const refresh = useCallback(async () => { try { applyResponse((await api.post<AuthResponse>("auth/refresh")).data); } catch { applyResponse(null); } finally { setLoading(false); } }, []);
  useEffect(() => { const timer = window.setTimeout(() => void refresh(), 0); return () => window.clearTimeout(timer); }, [refresh]);
  const authenticate = async (path: string, body: object) => { setLoading(true); setError(null); try { applyResponse((await api.post<AuthResponse>(path, body)).data); return true; } catch (cause) { setError(cause instanceof Error ? cause.message : "Не удалось выполнить запрос"); return false; } finally { setLoading(false); } };
  const register = (username: string, email: string, password: string) => authenticate("auth/register", { username, email, password });
  const login = (email: string, password: string) => authenticate("auth/login", { email, password });
  const logout = async () => { try { await api.post("auth/logout"); } finally { applyResponse(null); } };
  const updateProfile = async (username: string, avatar?: File) => {
    if (!accessToken) return false;
    setError(null);
    try {
      const form = new FormData();
      if (avatar) { form.set("avatar", avatar); form.set("username", username); }
      const response = avatar ? await api.post<{ user: User }>("auth/avatar", form) : await api.patch<{ user: User }>("auth/profile", { username });
      if (response.data.user) setUser(response.data.user);
      return true;
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : "Не удалось сохранить профиль");
      return false;
    }
  };
  const value = { user, accessToken, loading, error, register, login, logout, updateProfile, clearError: () => setError(null) };
  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth() { const context = useContext(AuthContext); if (!context) throw new Error("useAuth must be used inside AuthProvider"); return context; }
