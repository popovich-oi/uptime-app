# Next.js frontend

Чистый стартовый проект на Next.js 16 с TypeScript, App Router и Tailwind CSS.

## Разработка

```sh
npm install
npm run dev
```

Откройте [http://localhost:3000](http://localhost:3000).

По умолчанию auth-запросы `/api/auth/*` проксируются на `http://localhost:8080`. Для другого адреса задайте `BACKEND_URL`. Refresh-токен хранится в HttpOnly cookie, а access-токен — только в памяти текущей вкладки и восстанавливается через `/api/auth/refresh`.
