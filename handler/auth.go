package handler

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"main/storage"
	"main/models"

)
var tokenLifetime  = 24 * time.Hour

func AuthoriseHeandler(w http.ResponseWriter, r *http.Request){
	if r.Method == "POST" {
		token := r.FormValue("token")
		if token == "" {
			http.Redirect(w, r, "/auth?error=empty_token", http.StatusSeeOther)
			return
		}

		// Проверяем токен в базе данных
		var user models.UserInfo
		err := storage.DB.QueryRow(`
			SELECT telegram_id, telegram_login, first_name, last_name, is_admin 
			FROM users 
			WHERE token = ? AND token_expiration > NOW()`, 
			token).Scan(
				&user.TelegramID,
				&user.TelegramLogin,
				&user.FirstName,
				&user.LastName,
				&user.IsAdmin,
			)

		if err != nil {
			if err == sql.ErrNoRows {
				http.Redirect(w, r, "/auth?error=invalid_token", http.StatusSeeOther)
			} else {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				log.Printf("Database error: %v", err)
			}
			return
		}

		// Устанавливаем cookie с токеном
		http.SetCookie(w, &http.Cookie{
			Name:     "auth_token",
			Value:    token,
			Path:     "/",
			MaxAge:   int(tokenLifetime.Seconds()),
			HttpOnly: true,
			Secure:   false, // В production установите true для HTTPS
		})

		// Перенаправляем на профиль
		http.Redirect(w, r, "/profile", http.StatusSeeOther)
	} else {
		// Получаем параметр error из URL
		errorMsg := r.URL.Query().Get("error")
		var errorHTML string
		  
		// Формируем HTML для отображения ошибки
		switch errorMsg {
		case "empty_token":
			errorHTML = `<div class="step error">Пожалуйста, введите токен</div>`
		case "invalid_token":
			errorHTML = `<div class="step error">Неверный или просроченный токен</div>`
		default:
			errorHTML = ""
		}
  
		// Читаем шаблон страницы
		path := filepath.Join("front", "pages", "auth.html")
		content, err := os.ReadFile(path)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
  
		// Вставляем сообщение об ошибке перед формой
		html := strings.Replace(string(content), 
			"<!-- ERROR_MESSAGE -->", 
			errorHTML, 
			1)
  
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(html))
	}
}