# Uptime App

Структура состоит из двух независимых проектов:

- `backend/` — минимальный HTTP-сервер на Go.
- `frontend/` — стартовое приложение Next.js 16 с TypeScript и App Router.

Реализация функций мониторинга пока не добавлена.

## Запуск frontend и backend

Установите зависимости фронтенда:

```sh
cd frontend && npm install
```

Затем из корня проекта запустите оба приложения одной командой:

```sh
./scripts/dev.sh
```

Frontend будет доступен на `http://localhost:3000`, backend — на `http://localhost:8080`. Остановка через `Ctrl+C` завершает оба процесса.
Скрипт также запускает PostgreSQL через Docker Compose и ждёт готовности базы данных.
