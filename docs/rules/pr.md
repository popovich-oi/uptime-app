# PR инструкция

- Создавай PR из текущей task-ветки в `master`; не создавай PR из `master` в `master`.
- Перед PR проверь `git status`, `git log master..HEAD`, `git diff —check` и убедись, что в PR не попадут незакоммиченные или нерелевантные файлы.
- Перед отправкой выполни проверки из локальных “AGENTS.md” для каждого затронутого приложения и укажи команды и результат в описании PR.
- Отправляй ветку командой `git push -u origin <branch-name>`. Если remote `origin` отсутствует, добавь его только после проверки целевого URL.
- Создавай PR через GitHub CLI: `gh pr create —base master --head <branch-name> --title “<conventional-commit-title>” --body-file <shody-files>`
- Название PR должно соответствовать Conventional Commits, быть коротким, в повелительном наклонении и не превышать 72 символа.
- Описание PR оформляй разделами `Summary`, `Changes`, `API / Configuration / Migrations`, `Verification`, `Risks / Follow-ups` и `Screenshots` при наличии UI-изменений.
- В `Summary` опиши пользовательский или системный эффект, в `Changes` - существенные изменения по приложениям, в `Verification` - фактически выполненные проверки.
- Явно указывай миграции, новые переменные окружения, изменения АРІ-контрактов, ограничения и известные проблемы сборки.