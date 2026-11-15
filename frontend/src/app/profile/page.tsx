"use client";

import Link from "next/link";
import Image from "next/image";
import { ChangeEvent, FormEvent, useState } from "react";
import { useAuth, User } from "../auth";

export default function ProfilePage() {
  const { user, loading, error, updateProfile, logout, clearError } = useAuth();
  if (loading) return <main className="profile-page"><p>Проверяем сессию…</p></main>;
  if (!user) return <main className="profile-page"><section className="profile-card"><h1>Сессия завершена</h1><p className="muted">Войдите, чтобы открыть профиль.</p><Link className="button profile-card__link" href="/">Перейти ко входу</Link></section></main>;

  return <main className="dashboard"><ProfileHeader user={user} logout={logout} /><ProfileEditor user={user} updateProfile={updateProfile} error={error} clearError={clearError} /></main>;
}

function ProfileHeader({ user, logout }: { user: User; logout: () => Promise<void> }) {
  const [open, setOpen] = useState(false);
  const [pending, setPending] = useState(false);
  const handleLogout = async () => { setPending(true); await logout(); setPending(false); };
  return <header className="dashboard__header"><Link className="brand" href="/">Uptime App</Link><div className="profile"><div className="profile__actions"><button type="button" className="profile__trigger" aria-expanded={open} aria-haspopup="menu" onClick={() => setOpen((value) => !value)}><span className="profile__avatar">{user.avatar_url ? <Image src={user.avatar_url} alt="" width={96} height={96} unoptimized /> : user.username.charAt(0).toUpperCase()}</span><span className="profile__name">{user.username}</span><span className="profile__chevron" aria-hidden="true">{open ? "⌃" : "⌄"}</span></button><button type="button" className="profile__logout profile__logout--inline" onClick={() => void handleLogout()} disabled={pending}>{pending ? "Выходим…" : "Выйти"}</button></div>{open && <div className="profile__menu" role="menu"><div className="profile__details"><strong>{user.username}</strong><span>{user.email}</span></div><Link className="profile__link" href="/profile">Профиль</Link></div>}</div></header>;
}

function ProfileEditor({ user, updateProfile, error, clearError }: { user: User; updateProfile: (username: string, avatar?: File) => Promise<boolean>; error: string | null; clearError: () => void }) {
  const [username, setUsername] = useState(user.username);
  const [saved, setSaved] = useState(false);
  const [uploading, setUploading] = useState(false);

  const submit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setSaved(false);
    clearError();
    if (await updateProfile(username)) setSaved(true);
  };

  const uploadAvatar = async (event: ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    event.target.value = "";
    if (!file) return;
    setSaved(false);
    clearError();
    setUploading(true);
    try {
      await updateProfile(user.username, file);
    } finally {
      setUploading(false);
    }
  };

  return (
    <section className="profile-content">
      <Link className="back-link" href="/">← В dashboard</Link>
      <p className="eyebrow">Профиль</p>
      <h1>Настройки профиля</h1>
      <p className="muted">Управляйте данными своего аккаунта.</p>
      <form className="profile-fields" onSubmit={submit}>
        <div className="profile-fields__details">
          <div className="profile-readonly"><span>Email</span><strong>{user.email}</strong></div>
          <label>Имя пользователя<input value={username} onChange={(event) => setUsername(event.target.value)} minLength={3} maxLength={32} required /></label>
          {error && <p className="error">{error}</p>}
          {saved && <p className="success">Профиль сохранён.</p>}
          <button className="button" type="submit" disabled={uploading}>Сохранить изменения</button>
        </div>
        <div className="profile-fields__avatar">
          <div className="profile-avatar-preview">{user.avatar_url ? <Image src={user.avatar_url} alt="Аватар профиля" width={96} height={96} unoptimized /> : <span>{user.username.charAt(0).toUpperCase()}</span>}</div>
          <label>Загрузить аватар<input type="file" accept="image/jpeg,image/png" onChange={(event) => void uploadAvatar(event)} disabled={uploading} /></label>
          {uploading && <p className="muted" role="status">Загружаем аватар…</p>}
          <p className="muted">JPEG или PNG, до 5 МБ.</p>
        </div>
      </form>
    </section>
  );
}
