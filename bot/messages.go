package bot

import (
	"log"
	"fmt"
    "strconv"
    "os"


	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
    "github.com/joho/godotenv"
)

// Глобальные переменные для хранения состояния диалогов
var (
    activeUserDialogs    = make(map[int64]int64)    // map[userChatID]managerChatID
    activeManagerDialogs = make(map[int64]int64)    // map[managerChatID]userChatID
    managerGroupID         int64   // ID группы менеджеров
    waitingForQuestion   = make(map[int64]bool)     // map[userChatID]bool - ожидаем вопрос от пользователя
)

// Запуск диалога с менеджером
func startManagerDialog(message *tgbotapi.Message) {
    // Загрузка .env файла
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}


    managerGroupID, err = strconv.ParseInt(os.Getenv("TELEGRAM_MANAGER_GROUP_ID"), 10, 64)
	if err != nil {
		log.Fatal("Invalid TELEGRAM_MANAGER_GROUP_ID in .env file")
	}

    userChatID := message.Chat.ID

    // Проверяем, не является ли сообщение текстом кнопки главного меню
    if isMainMenuButton(message.Text) {
        return
    }

    // Создаем сообщение в группе менеджеров
    msgText := fmt.Sprintf("Новый вопрос от @%s (%s %s):\n%s", 
        message.From.UserName, 
        message.From.FirstName, 
        message.From.LastName,
        message.Text)

    msg := tgbotapi.NewMessage(managerGroupID, msgText)

    sentMsg, err := Bot.Send(msg)
    if err != nil {
        log.Printf("Ошибка при создании диалога с менеджером: %v", err)
        reply := tgbotapi.NewMessage(userChatID, "Не удалось соединиться с менеджером. Попробуйте позже.")
        Bot.Send(reply)
        return
    }

    // Сохраняем информацию о диалоге
    activeUserDialogs[userChatID] = sentMsg.Chat.ID
    activeManagerDialogs[sentMsg.Chat.ID] = userChatID

    // Отправляем пользователю сообщение с инструкцией
    reply := tgbotapi.NewMessage(userChatID, "Ваш вопрос передан менеджеру. \nТеперь все ваши сообщения будут переправляться человеку. Вам ответит первый освободившийся сотрудник. \nДля возвращения в главное меню нажмите кнопку \"Завершить диалог\". Вы сможете задать следующий вопрос после завершения диалога, нажав на кнопку \"Задать вопрос менеджеру\" в главном меню")
    reply.ReplyMarkup = createUserDialogKeyboard()
    Bot.Send(reply)
}

// Обработка сообщений от менеджеров
func HandleManagerMessage(message *tgbotapi.Message) {
    // Игнорируем сообщения с текстом кнопок главного меню
    if isMainMenuButton(message.Text) {
        return
    }

    // Если это ответ на сообщение
    if message.ReplyToMessage != nil {
        // Ищем пользователя по ID чата
        if userChatID, exists := activeManagerDialogs[message.Chat.ID]; exists {
            // Отправляем ответ пользователю
            reply := tgbotapi.NewMessage(userChatID, fmt.Sprintf(message.Text))
            reply.ReplyMarkup = createUserDialogKeyboard()
            if _, err := Bot.Send(reply); err != nil {
                log.Printf("Ошибка при отправке ответа пользователю: %v", err)
            }
        }
        return
    }

    // Обработка кнопки завершения диалога
    if message.Text == "Завершить диалог" {
        if userChatID, exists := activeManagerDialogs[message.Chat.ID]; exists {
            // Уведомляем пользователя
            sendMainMenu(userChatID)
            
            // Удаляем информацию о диалоге
            delete(activeUserDialogs, userChatID)
            delete(activeManagerDialogs, message.Chat.ID)
           
        } 
    }
}

// Создание клавиатуры для пользователя в режиме диалога
func createUserDialogKeyboard() tgbotapi.ReplyKeyboardMarkup {
    keyboard := tgbotapi.NewReplyKeyboard(
        tgbotapi.NewKeyboardButtonRow(
            tgbotapi.NewKeyboardButton("Завершить диалог"),
        ),
    )
    keyboard.ResizeKeyboard = true
    return keyboard
}

// Отправка главного меню
func sendMainMenu(chatID int64) {
    mainKeyboard := tgbotapi.NewReplyKeyboard(
        tgbotapi.NewKeyboardButtonRow(
            tgbotapi.NewKeyboardButton("Войти в личный кабинет на сайте"),
        ),
        tgbotapi.NewKeyboardButtonRow(
            tgbotapi.NewKeyboardButton("Задать вопрос менеджеру"),
        ),
        tgbotapi.NewKeyboardButtonRow(
            tgbotapi.NewKeyboardButton("ℹ Помощь"),
        ),
    )
    mainKeyboard.ResizeKeyboard = true

    msg := tgbotapi.NewMessage(chatID, "Выберите действие:")
    msg.ReplyMarkup = mainKeyboard
    if _, err := Bot.Send(msg); err != nil {
        log.Printf("Ошибка при отправке главного меню: %v", err)
    }
}

// Отправка справки
func sendHelp(chatID int64) {
    msg := tgbotapi.NewMessage(chatID, "Здесь будет текст помощи...")
    if _, err := Bot.Send(msg); err != nil {
        log.Printf("Ошибка при отправке справки: %v", err)
    }
}

func isMainMenuButton(text string) bool {
    switch text {
    case "Войти в личный кабинет на сайте", 
         "Задать вопрос менеджеру", 
         "ℹ Помощь",
         "Главное меню":
        return true
    }
    return false
}