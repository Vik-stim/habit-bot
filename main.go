package main

import (
	"fmt"
	"log"
	"os"
	"strings"

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
	}
}
