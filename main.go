package main

import (
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

	// Проверяем подключение к базе
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

	// Настраиваем получение сообщений
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	// Главный цикл: ждём сообщения и отвечаем
	for update := range updates {
		if update.Message == nil {
			continue
		}

		text := update.Message.Text
		chatID := update.Message.Chat.ID

		// Разрезаем сообщение на слова
		parts := strings.Fields(text)

		// Сообщение должно быть хотя бы из двух слов: "/add" и название
		if len(parts) < 2 {
			msg := tgbotapi.NewMessage(chatID, "Напиши так: /add пить воду")
			bot.Send(msg)
			continue
		}

		if parts[0] == "/add" {
			// Склеиваем все слова после /add обратно в название привычки
			name := strings.Join(parts[1:], " ")

			// Сохраняем в базу
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
		}
	}
}
