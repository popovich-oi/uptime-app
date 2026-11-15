"use client";

import { createContext, useCallback, useContext, useEffect, useState } from "react";

export type User = { id: number; username: string; email: string; avatar_url: string };
type AuthResponse = { access_token: string; user: User };
type AuthContextValue = { user: User | null; accessToken: string | null; loading: boolean; error: string | null; register: (username: string, email: string, password: string) => Promise<boolean>; login: (email: string, password: string) => Promise<boolean>; logout: () => Promise<void>; updateProfile: (username: string, avatar?: File) => Promise<boolean>; clearError: () => void };
const AuthContext = createContext<AuthContextValue | null>(null);

export async function request<T>(path: string, init?: RequestInit): Promise<T | null> {
  const response = await fetch(`/api/${path}`, { ...init, credentials: "include", headers: { Accept: "application/json", ...(init?.body instanceof FormData ? {} : { "Content-Type": "application/json" }), ...init?.headers } });
  if (response.status === 204) return null;
  const raw = await response.text();
  let body: T & { error?: string };
  try {
    body = JSON.parse(raw) as T & { error?: string };
  } catch {
    throw new Error(response.ok ? "Сервер вернул некорректный ответ" : `Ошибка сервера (${response.status})`);
  }
  if (!response.ok) throw new Error(body.error || "Не удалось выполнить запрос");
  return body;
}

export function AuthProvider({ children }: Readonly<{ children: React.ReactNode }>) {
  const [user, setUser] = useState<User | null>(null);
  const [accessToken, setAccessToken] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const applyResponse = (response: AuthResponse | null) => { setAccessToken(response?.access_token ?? null); setUser(response?.user ?? null); };
  const refresh = useCallback(async () => { try { applyResponse(await request<AuthResponse>("auth/refresh", { method: "POST" })); } catch { applyResponse(null); } finally { setLoading(false); } }, []);
  useEffect(() => { const timer = window.setTimeout(() => void refresh(), 0); return () => window.clearTimeout(timer); }, [refresh]);
  const authenticate = async (path: string, body: object) => { setLoading(true); setError(null); try { applyResponse(await request<AuthResponse>(path, { method: "POST", body: JSON.stringify(body) })); return true; } catch (cause) { setError(cause instanceof Error ? cause.message : "Не удалось выполнить запрос"); return false; } finally { setLoading(false); } };
  const register = (username: string, email: string, password: string) => authenticate("auth/register", { username, email, password });
  const login = (email: string, password: string) => authenticate("auth/login", { email, password });
  const logout = async () => { try { await request("auth/logout", { method: "POST" }); } finally { applyResponse(null); } };
  const updateProfile = async (username: string, avatar?: File) => {
    if (!accessToken) return false;
    setError(null);
    try {
      const form = new FormData();
      if (avatar) { form.set("avatar", avatar); form.set("username", username); }
      const response = await request<AuthResponse>(avatar ? "auth/avatar" : "auth/profile", { method: avatar ? "POST" : "PATCH", body: avatar ? form : JSON.stringify({ username }), headers: { Authorization: `Bearer ${accessToken}` } });
      if (response?.user) setUser(response.user);
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
