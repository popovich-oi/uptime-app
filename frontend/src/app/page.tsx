"use client";

import { FormEvent, useEffect, useRef, useState } from "react";
import Link from "next/link";
import Image from "next/image";
import { useAuth } from "./auth";
import { api } from "./api";

type Monitor = { id: number; url: string; interval_seconds: number; created_at: string };
type IntervalUnit = "seconds" | "minutes" | "hours";
const multipliers: Record<IntervalUnit, number> = { seconds: 1, minutes: 60, hours: 3600 };

export default function Home() {
  const { user, loading, error, register, login, logout, clearError } = useAuth();
  const [mode, setMode] = useState<"login" | "register">("login");
  const [username, setUsername] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [formError, setFormError] = useState<string | null>(null);
  const [profileOpen, setProfileOpen] = useState(false);
  const [logoutPending, setLogoutPending] = useState(false);
  if (loading && !user) return <main className="auth-shell"><p>Проверяем сессию…</p></main>;
  if (user) {
    const handleLogout = async () => {
      setLogoutPending(true);
      await logout();
      setLogoutPending(false);
    };

    return <main className="dashboard"><header className="dashboard__header"><Link className="brand" href="/">Uptime App</Link><div className="profile"><div className="profile__actions"><button type="button" className="profile__trigger" aria-expanded={profileOpen} aria-haspopup="menu" onClick={() => setProfileOpen((open) => !open)}><span className="profile__avatar">{user.avatar_url ? <Image src={user.avatar_url} alt="" width={96} height={96} unoptimized /> : user.username.charAt(0).toUpperCase()}</span><span className="profile__name">{user.username}</span><span className="profile__chevron" aria-hidden="true">{profileOpen ? "⌃" : "⌄"}</span></button><button type="button" className="profile__logout profile__logout--inline" onClick={() => void handleLogout()} disabled={logoutPending}>{logoutPending ? "Выходим…" : "Выйти"}</button></div>{profileOpen && <div className="profile__menu" role="menu"><div className="profile__details"><strong>{user.username}</strong><span>{user.email}</span></div><Link className="profile__link" href="/profile">Профиль</Link></div>}</div></header><MonitorDashboard /></main>;
  }
  const submit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault(); clearError(); setFormError(null);
    if (mode === "register" && (username.trim().length < 3 || username.trim().length > 32)) return setFormError("Имя пользователя должно содержать от 3 до 32 символов.");
    if (!/^\S+@\S+\.\S+$/.test(email)) return setFormError("Введите корректный email.");
    if (mode === "register" && password.length < 8) return setFormError("Пароль должен содержать минимум 8 символов.");
    await (mode === "register" ? register(username, email, password) : login(email, password));
  };
  return <main className="auth-shell"><section className="card"><p className="eyebrow">Uptime App</p><h1>{mode === "login" ? "Войдите в аккаунт" : "Создайте аккаунт"}</h1><p className="muted">Мониторинг сервисов начинается с безопасной сессии.</p><div className="tabs"><button type="button" className={mode === "login" ? "tab active" : "tab"} onClick={() => { setMode("login"); clearError(); }}>Войти</button><button type="button" className={mode === "register" ? "tab active" : "tab"} onClick={() => { setMode("register"); clearError(); }}>Регистрация</button></div><form onSubmit={submit}>{mode === "register" && <label>Имя пользователя<input value={username} onChange={(event) => setUsername(event.target.value)} autoComplete="username" required /></label>}<label>Email<input type="email" value={email} onChange={(event) => setEmail(event.target.value)} autoComplete="email" required /></label><label>Пароль<input type="password" value={password} onChange={(event) => setPassword(event.target.value)} autoComplete={mode === "login" ? "current-password" : "new-password"} required /></label>{(formError || error) && <p className="error">{formError || error}</p>}<button className="button" type="submit" disabled={loading}>{loading ? "Подождите…" : mode === "login" ? "Войти" : "Создать аккаунт"}</button></form></section></main>;
}

function MonitorDashboard() {
  const [monitors, setMonitors] = useState<Monitor[]>([]);
  const [url, setURL] = useState("");
  const [intervalValue, setIntervalValue] = useState(5);
  const [intervalUnit, setIntervalUnit] = useState<IntervalUnit>("seconds");
  const [open, setOpen] = useState(false);
  const [pending, setPending] = useState(false);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [loaded, setLoaded] = useState(false);
  const sequence = useRef(0);
  useEffect(() => {
    const currentSequence = sequence;
    const id = ++currentSequence.current;
    const controller = new AbortController();
    void api.get<{ monitors: Monitor[] }>("monitors", { signal: controller.signal })
      .then((response) => { if (id === sequence.current) { setMonitors(response.data.monitors ?? []); setLoaded(true); } })
      .catch((cause) => { if (id === sequence.current && !controller.signal.aborted) setError(cause instanceof Error ? cause.message : "Не удалось загрузить точки мониторинга"); })
      .finally(() => { if (id === sequence.current) setLoading(false); });
    return () => { controller.abort(); currentSequence.current++; };
  }, []);
  const submit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (!Number.isSafeInteger(intervalValue) || intervalValue < 1 || intervalValue > Math.floor(Number.MAX_SAFE_INTEGER / multipliers[intervalUnit])) {
      setError("Укажите допустимый целый интервал."); return;
    }
    setError(null); setPending(true);
    try {
      await api.post<{ monitor: Monitor }>("monitors", { url, interval_value: intervalValue, interval_unit: intervalUnit });
      setURL(""); setOpen(false);
      const id = ++sequence.current;
      setLoading(true);
      try {
        const response = await api.get<{ monitors: Monitor[] }>("monitors");
        if (id === sequence.current) { setMonitors(response.data.monitors ?? []); setLoaded(true); }
      } catch (cause) {
        if (id === sequence.current) setError(cause instanceof Error ? cause.message : "Не удалось загрузить точки мониторинга");
      } finally { if (id === sequence.current) setLoading(false); }
    } catch (cause) { setError(cause instanceof Error ? cause.message : "Не удалось создать точку мониторинга"); }
    finally { setPending(false); }
  };
  return <section className="monitor-content"><div className="monitor-heading"><div><p className="eyebrow">Мониторинг</p><h1>Ваши сайты</h1><p className="muted">Создайте точку мониторинга и задайте период опроса.</p></div><button className="button monitor-add" type="button" onClick={() => setOpen((value) => !value)}>{open ? "Закрыть" : "Добавить сайт"}</button></div>{open && <form className="monitor-form" onSubmit={submit}><label>URL сайта<input type="url" maxLength={2048} value={url} onChange={(event) => setURL(event.target.value)} placeholder="https://example.com" required /></label><div className="interval-fields"><label>Число<input type="number" min={1} max={Math.floor(Number.MAX_SAFE_INTEGER / multipliers[intervalUnit])} value={intervalValue} onChange={(event) => setIntervalValue(Number(event.target.value))} required /></label><label>Единица<select value={intervalUnit} onChange={(event) => setIntervalUnit(event.target.value as IntervalUnit)}><option value="seconds">Секунды</option><option value="minutes">Минуты</option><option value="hours">Часы</option></select></label></div><button className="button" type="submit" disabled={pending}>{pending ? "Создаём…" : "Создать"}</button></form>}{error && <p className="error" role="alert">{error}</p>}{loading && <p>Загружаем точки мониторинга…</p>}{loaded && monitors.length === 0 && !loading ? <div className="monitor-empty"><h2>Пока нет точек мониторинга</h2><p className="muted">Добавьте сайт, чтобы он появился в этом списке.</p></div> : <div className="monitor-list">{monitors.map((monitor) => <article className="monitor-card" key={monitor.id}><div><strong>{monitor.url}</strong><span>Проверка: {formatInterval(monitor.interval_seconds)}</span></div><span className="monitor-status">Создано</span></article>)}</div>}</section>;
}

function formatInterval(seconds: number) { if (seconds % 3600 === 0) return `каждые ${seconds / 3600} ч.`; if (seconds % 60 === 0) return `каждые ${seconds / 60} мин.`; return `каждые ${seconds} сек.`; }
