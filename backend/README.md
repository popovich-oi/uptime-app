# Go backend

Backend предоставляет регистрацию, вход и обновление access JWT через refresh cookie.

## Запуск

Из корня проекта запустите PostgreSQL:

```sh
docker compose up -d postgres
```

Скопируйте `.env.example` в `.env` и замените `JWT_SECRET` на случайное значение. Затем из `backend/`:

## Запуск

```sh
go run ./cmd/server
```

Сервер доступен на `http://localhost:8080`. Для прямых браузерных запросов CORS разрешает origin `http://localhost:3000` и передачу cookie. Для другого frontend задайте `CORS_ALLOWED_ORIGIN` как точный origin, включая схему и порт, без завершающего `/`; `*` не поддерживается. CORS не заменяет авторизацию и защиту от CSRF. При стандартном запуске frontend использует прокси Next.js.

Эндпоинты:

```sh
curl -i -c cookies.txt -X POST http://localhost:8080/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"username":"alice","email":"alice@example.com","password":"password123"}'

curl -i -c cookies.txt -X POST http://localhost:8080/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"alice@example.com","password":"password123"}'

curl -i -b cookies.txt -c cookies.txt -X POST http://localhost:8080/auth/refresh
```

Refresh-токен хранится в HttpOnly cookie, а access JWT возвращается в JSON-поле `access_token`.

Аватары JPEG и PNG размером до 5 МБ загружаются авторизованным запросом
`POST /auth/avatar` (multipart-поля `avatar` и `username`) и раздаются по
`/avatars/`. По умолчанию файлы хранятся в `backend/uploads/avatars` при
запуске из каталога `backend/`. Путь можно изменить переменной `AVATAR_DIR`,
например `AVATAR_DIR=uploads/avatars`; каталог должен сохраняться между перезапусками.


Произвольные файлы до 25 МБ можно загрузить авторизованным запросом
`POST /auth/files` с multipart-полем `file` и заголовком `Authorization: Bearer <access_token>`.
Ответ содержит URL вида `/files/<идентификатор>`; по нему файл доступен без
авторизации как скачивание. Не загружайте конфиденциальные данные, если URL
может стать известен другим. Файлы хранятся в `uploads/files` относительно
рабочего каталога сервера; путь можно изменить через `UPLOAD_DIR` (например,
`UPLOAD_DIR=uploads/files`). Каталог должен сохраняться между перезапусками и
исключаться из Git. Исходное имя и тип файла не сохраняются.
Аватары используют тот же модуль хранения, но для них действуют отдельные
проверки формата и размера; ранее загруженные аватары остаются доступны.
