# Трекер привычек — Telegram-бот на Go

Бот помогает отслеживать ежедневные привычки и поддерживет серию выполнений (streak).

## Возможности

- `/add &lt;название&gt;` — добавить привычку (например, `/add пить воду`)
- `/list` — показать все привычки
- `/done &lt;название&gt;` — отметить выполнение сегодня
- Защита от повторной отметки в один день
- Счётчик серии дней 🔥

## Технологии

- Go
- SQLite (modernc.org/sqlite)
- [go-telegram-bot-api](https://github.com/go-telegram-bot-api/telegram-bot-api)

## Запуск

```bash
export TELEGRAM_TOKEN="твой_токен"
go run .
