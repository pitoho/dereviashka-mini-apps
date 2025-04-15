package handler

import (
	"database/sql"
	// "encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	// "time"

	"main/models"
	"main/storage"
)

// Функция для отображения профиля обычного пользователя
func ServeUserProfile(w http.ResponseWriter, r *http.Request, user models.UserInfo) {
	// Читаем шаблон страницы
	path := filepath.Join("front", "pages", "profile.html")
	content, err := os.ReadFile(path)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Printf("Error reading user profile template: %v", err)
		return
	}

	// Заменяем плейсхолдеры в шаблоне
	html := string(content)
	html = strings.ReplaceAll(html, "{{username}}", user.TelegramLogin)
	html = strings.ReplaceAll(html, "{{first_name}}", user.FirstName)
	html = strings.ReplaceAll(html, "{{last_name}}", user.LastName)

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(html))
}

func ProfileHandler(w http.ResponseWriter, r *http.Request, user models.UserInfo) {
	// Проверяем токен из cookie
	tokenCookie, err := r.Cookie("auth_token")
	if err != nil || tokenCookie.Value == "" {
		http.Redirect(w, r, "/auth", http.StatusSeeOther)
		return
	}

	// Проверяем токен в базе данных
	err = storage.DB.QueryRow(`
		SELECT telegram_id, telegram_login, first_name, last_name, is_admin 
		FROM users 
		WHERE token = ? AND token_expiration > NOW()`, 
		tokenCookie.Value).Scan(
			&user.TelegramID,
			&user.TelegramLogin,
			&user.FirstName,
			&user.LastName,
			&user.IsAdmin,
		)

	if err != nil {
		if err == sql.ErrNoRows {
			// Токен не найден или истек
			http.Redirect(w, r, "/auth", http.StatusSeeOther)
		} else {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			log.Printf("Database error: %v", err)
		}
		return
	}

	// Если пользователь авторизован, показываем профиль
	if user.IsAdmin {
		// Для админа показываем страницу с дополнительной кнопкой
		ServeAdminProfile(w, r, user)
	} else {
		// Для обычного пользователя
		ServeUserProfile(w, r, user)
	}
}

// getCurrentUserID извлекает ID пользователя из токена в cookie
func GetCurrentUserID(r *http.Request) int {
	tokenCookie, err := r.Cookie("auth_token")
	if err != nil || tokenCookie.Value == "" {
		return 0
	}

	var userID int
	
	err = storage.DB.QueryRow(`
		SELECT id 
		FROM users 
		WHERE token = ? AND token_expiration > NOW()`, 
		tokenCookie.Value).Scan(&userID)

	if err != nil {
		if err != sql.ErrNoRows {
			log.Printf("Database error in getCurrentUserID: %v", err)
		}
		return 0
	}

	return userID
}