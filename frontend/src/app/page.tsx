"use client";

import { FormEvent, useEffect, useState } from "react";
import Link from "next/link";
import Image from "next/image";
import { Monitor, useAuth } from "./auth";

export default function Home() {
  const { user, loading, error, register, login, logout, clearError, listMonitors, createMonitor } = useAuth();
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

    return <main className="dashboard"><header className="dashboard__header"><Link className="brand" href="/">Uptime App</Link><div className="profile"><div className="profile__actions"><button type="button" className="profile__trigger" aria-expanded={profileOpen} aria-haspopup="menu" onClick={() => setProfileOpen((open) => !open)}><span className="profile__avatar">{user.avatar_url ? <Image src={user.avatar_url} alt="" width={96} height={96} unoptimized /> : user.username.charAt(0).toUpperCase()}</span><span className="profile__name">{user.username}</span><span className="profile__chevron" aria-hidden="true">{profileOpen ? "⌃" : "⌄"}</span></button><button type="button" className="profile__logout profile__logout--inline" onClick={() => void handleLogout()} disabled={logoutPending}>{logoutPending ? "Выходим…" : "Выйти"}</button></div>{profileOpen && <div className="profile__menu" role="menu"><div className="profile__details"><strong>{user.username}</strong><span>{user.email}</span></div><Link className="profile__link" href="/profile">Профиль</Link></div>}</div></header><MonitorDashboard listMonitors={listMonitors} createMonitor={createMonitor} error={error} clearError={clearError} /></main>;
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

function MonitorDashboard({ listMonitors, createMonitor, error, clearError }: { listMonitors: () => Promise<Monitor[]>; createMonitor: (url: string, value: number, unit: string) => Promise<boolean>; error: string | null; clearError: () => void }) {
  const [monitors, setMonitors] = useState<Monitor[]>([]);
  const [url, setURL] = useState("");
  const [intervalValue, setIntervalValue] = useState(5);
  const [intervalUnit, setIntervalUnit] = useState("seconds");
  const [open, setOpen] = useState(false);
  const [pending, setPending] = useState(false);
  useEffect(() => { void listMonitors().then(setMonitors); }, [listMonitors]);
  const submit = async (event: FormEvent<HTMLFormElement>) => { event.preventDefault(); clearError(); setPending(true); if (await createMonitor(url, intervalValue, intervalUnit)) { setURL(""); setOpen(false); setMonitors(await listMonitors()); } setPending(false); };
  return <section className="monitor-content"><div className="monitor-heading"><div><p className="eyebrow">Мониторинг</p><h1>Ваши сайты</h1><p className="muted">Создайте точку мониторинга и задайте период опроса.</p></div><button className="button monitor-add" type="button" onClick={() => setOpen((value) => !value)}>{open ? "Закрыть" : "Добавить сайт"}</button></div>{open && <form className="monitor-form" onSubmit={submit}><label>URL сайта<input type="url" value={url} onChange={(event) => setURL(event.target.value)} placeholder="https://example.com" required /></label><div className="interval-fields"><label>Число<input type="number" min={1} value={intervalValue} onChange={(event) => setIntervalValue(Number(event.target.value))} required /></label><label>Единица<select value={intervalUnit} onChange={(event) => setIntervalUnit(event.target.value)}><option value="seconds">Секунды</option><option value="minutes">Минуты</option><option value="hours">Часы</option></select></label></div>{error && <p className="error">{error}</p>}<button className="button" type="submit" disabled={pending}>{pending ? "Создаём…" : "Создать"}</button></form>}{!open && error && <p className="error">{error}</p>}{monitors.length === 0 ? <div className="monitor-empty"><h2>Пока нет точек мониторинга</h2><p className="muted">Добавьте сайт, чтобы он появился в этом списке.</p></div> : <div className="monitor-list">{monitors.map((monitor) => <article className="monitor-card" key={monitor.id}><div><strong>{monitor.url}</strong><span>Проверка: {formatInterval(monitor.interval_seconds)}</span></div><span className="monitor-status">Создано</span></article>)}</div>}</section>;
}

function formatInterval(seconds: number) { if (seconds % 3600 === 0) return `каждые ${seconds / 3600} ч.`; if (seconds % 60 === 0) return `каждые ${seconds / 60} мин.`; return `каждые ${seconds} сек.`; }
