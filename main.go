package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	token := os.Getenv("TELEGRAM_TOKEN")
	if token == "" {
		log.Fatal("Не задана переменная TELEGRAM_TOKEN")
	}

	// Подключение к базе
	db, err := openDB()
	if err != nil {
		log.Fatal("Ошибка базы:", err)
	}
	defer db.Close()
	log.Println("База данных подключена")

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Бот авторизован как @%s", bot.Self.UserName)

	// Получение сообщений
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	// Главный цикл
	for update := range updates {
		if update.Message == nil {
			continue
		}

		text := update.Message.Text
		chatID := update.Message.Chat.ID

		parts := strings.Fields(text)

		// --- Команда /add ---
		if len(parts) > 0 && parts[0] == "/add" {
			if len(parts) < 2 {
				msg := tgbotapi.NewMessage(chatID, "Напиши так: /add пить воду")
				bot.Send(msg)
				continue
			}

			name := strings.Join(parts[1:], " ")

			_, err := db.Exec(
				"INSERT INTO habits (chat_id, name) VALUES (?, ?)",
				chatID, name,
			)
			if err != nil {
				msg := tgbotapi.NewMessage(chatID, "Ошибка сохранения 😢")
				bot.Send(msg)
				continue
			}

			msg := tgbotapi.NewMessage(chatID, "Привычка \""+name+"\" добавлена! ✅")
			bot.Send(msg)
			continue
		}

		// --- Команда /list ---
		if len(parts) > 0 && parts[0] == "/list" {
			rows, err := db.Query(
				"SELECT name FROM habits WHERE chat_id = ?", chatID)
			if err != nil {
				msg := tgbotapi.NewMessage(chatID, "Ошибка чтения базы 😢")
				bot.Send(msg)
				continue
			}
			defer rows.Close()

			var names []string
			for rows.Next() {
				var name string
				rows.Scan(&name)
				names = append(names, name)
			}
			if err := rows.Err(); err != nil {
				msg := tgbotapi.NewMessage(chatID, "Ошибка чтения базы 😢")
				bot.Send(msg)
				continue
			}

			if len(names) == 0 {
				msg := tgbotapi.NewMessage(chatID, "У тебя пока нет привычек. Добавь: /add пить воду")
				bot.Send(msg)
				continue
			}

			answer := "Твои привычки:\n"
			for i, name := range names {
				answer += fmt.Sprintf("%d. %s\n", i+1, name)
			}

			msg := tgbotapi.NewMessage(chatID, answer)
			bot.Send(msg)
			continue
		}

		// --- Любое другое сообщение ---
		msg := tgbotapi.NewMessage(chatID, "Команды: /add <название>, /list")
		bot.Send(msg)

		// --- Команда /done ---
		if len(parts) > 0 && parts[0] == "/done" {
			if len(parts) < 2 {
				msg := tgbotapi.NewMessage(chatID, "Напиши так: /done пить воду")
				bot.Send(msg)
				continue
			}

			name := strings.Join(parts[1:], " ")

			// Находим привычку по названию
			var habitID int
			err := db.QueryRow(
				"SELECT id FROM habits WHERE chat_id = ? AND name = ?",
				chatID, name,
			).Scan(&habitID)
			if err != nil {
				msg := tgbotapi.NewMessage(chatID, "Такой привычки нет. Посмотри список: /list")
				bot.Send(msg)
				continue
			}

			// Сегодняшняя дата в формате 2026-10-20
			today := time.Now().Format("2006-01-02")

			// Вставляем отметку; если уже есть — UNIQUE не даст
			res, err := db.Exec(
				"INSERT OR IGNORE INTO checks (habit_id, date) VALUES (?, ?)",
				habitID, today,
			)
			if err != nil {
				msg := tgbotapi.NewMessage(chatID, "Ошибка сохранения 😢")
				bot.Send(msg)
				continue
			}

			// Сколько строк реально вставилось?
			affected, _ := res.RowsAffected()
			if affected == 0 {
				msg := tgbotapi.NewMessage(chatID, "Сегодня уже отмечено! Возвращайся завтра 💪")
				bot.Send(msg)
				continue
			}

			// Считаем серию: сколько отметок у этой привычки
			var count int
			db.QueryRow(
				"SELECT COUNT(*) FROM checks WHERE habit_id = ?", habitID,
			).Scan(&count)

			msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Отметил! Серия: %d дней 🔥", count))
			bot.Send(msg)
			continue
		}
	}
}
