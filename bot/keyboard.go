package bot

import (
	"sync"
	"fmt"
	// "main/models"

	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var (
	// Для хранения состояний чатов
	chatStates    = make(map[int64]string)
	statesMutex   sync.RWMutex
)

func HandleMessage(message *tgbotapi.Message) {
    // Если сообщение из группы менеджеров
    if message.Chat.ID == managerGroupID {
        HandleManagerMessage(message)
        return
    }

    userChatID := message.Chat.ID

    // Обработка кнопки "Главное меню"
    if message.Text == "Главное меню" {
        sendMainMenu(userChatID)
        return
    }

    // Если пользователь в режиме ожидания вопроса
    if waitingForQuestion[userChatID] {
        delete(waitingForQuestion, userChatID)
        startManagerDialog(message)
        return
    }

    // Если пользователь в режиме диалога с менеджером
    if _, inDialog := activeUserDialogs[userChatID]; inDialog {
        // Обработка кнопки "Завершить диалог" от пользователя
        if message.Text == "Завершить диалог" {
            // Находим чат менеджера
            if managerChatID, exists := activeUserDialogs[userChatID]; exists {
                // Удаляем информацию о диалоге
                delete(activeUserDialogs, userChatID)
                delete(activeManagerDialogs, managerChatID)
                
                // Уведомляем менеджера
				msg := tgbotapi.NewMessage(managerChatID, fmt.Sprintf("Пользователь @%s завершил диалог", message.From.UserName))
                Bot.Send(msg)
            }
            
            // Отправляем пользователю главное меню
            sendMainMenu(userChatID)
            return
        }
        
        // Пересылаем сообщение менеджеру
        HandleManagerMessage(message)
        return
    }

    // Обработка команд и кнопок главного меню
    switch {
    case message.IsCommand() && message.Command() == "start":
        sendMainMenu(userChatID)

    case message.Text == "Войти в личный кабинет на сайте":
        HandleLoginCommand(message)

    case message.Text == "Задать вопрос менеджеру":
        waitingForQuestion[userChatID] = true
        msg := tgbotapi.NewMessage(userChatID, "Пожалуйста, напишите ваш вопрос:")
        Bot.Send(msg)

    case message.Text == "ℹ Помощь":
        sendHelp(userChatID)
    }
}