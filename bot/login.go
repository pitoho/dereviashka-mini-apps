package bot

import (
	"database/sql"
	// "main/models"
	"time"
	"sync"
	"fmt"
	"crypto/rand"
	"encoding/hex"
	"log"
	"html"

	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
	_ "github.com/go-sql-driver/mysql"
)

var (
	Bot *tgbotapi.BotAPI
	DB  *sql.DB
	tokensMutex sync.Mutex
	tokenLifetime = 24 * time.Hour
	cleanupInterval = 1 * time.Hour
)

func InitDB(dataSourceName string) error {
	var err error
	DB, err = sql.Open("mysql", dataSourceName)
	if err != nil {
		return err
	}
	
	err = DB.Ping()
	if err != nil {
		return err
	}
	
	return nil
}

func HandleLoginCommand(msg *tgbotapi.Message) {
    token := GenerateToken()
    chatID := msg.Chat.ID

    // Проверяем, есть ли пользователь в базе и является ли он админом
    var isAdmin bool
    err := DB.QueryRow("SELECT is_admin FROM users WHERE telegram_id = ?", chatID).Scan(&isAdmin)
    if err != nil && err != sql.ErrNoRows {
        log.Printf("Ошибка при проверке пользователя: %v", err)
        sendErrorMessage(chatID)
        return
    }

    expiration := time.Now().Add(tokenLifetime)
    
    // Добавляем/обновляем пользователя в базе
    _, err = DB.Exec("CALL upsert_user(?, ?, ?, ?, ?, ?)", 
        chatID, 
        msg.From.UserName, 
        msg.From.FirstName, 
        msg.From.LastName, 
        token, 
        expiration)
    
    if err != nil {
        log.Printf("Ошибка при сохранении пользователя: %v", err)
        sendErrorMessage(chatID)
        return
    }

	response := fmt.Sprintf(
		"Ваш токен для входа: \n\n<code>%s</code>\n\n" +
		"Нажмите на него, чтобы скопировать, затем перейдите обратно на " +
		"<a href=\"http://localhost:8080/profile\">сайт</a> " +
		"и введите этот токен в соответствующее поле.\n" +
		"Токен действителен 24 часа.",
		html.EscapeString(token))
	
	reply := tgbotapi.NewMessage(chatID, response)
	reply.ParseMode = "HTML"
    
    // Вместо клавиатуры с одной кнопкой отправляем главное меню
    sendMainMenu(chatID)
    Bot.Send(reply)
}

func sendErrorMessage(chatID int64) {
	msg := tgbotapi.NewMessage(chatID, "Произошла ошибка при обработке вашего запроса. Пожалуйста, попробуйте позже.")
	Bot.Send(msg)
}

func GenerateToken() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func TokenCleanup() {
	for {
		time.Sleep(cleanupInterval)
		_, err := DB.Exec("CALL cleanup_expired_tokens()")
		if err != nil {
			log.Printf("Ошибка при очистке токенов: %v", err)
		}
	}
}