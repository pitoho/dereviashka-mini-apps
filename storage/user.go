package storage

import (
	"main/models"
	"database/sql"
	"fmt"
)

// GetUserByID возвращает информацию о пользователе по его ID
func GetUserByID(userID int64) (*models.UserInfo, error) {
	var user models.UserInfo
	var (
		tokenExpiration sql.NullTime
		isAdmin         bool
	)
	
	query := `
		SELECT 
			telegram_id, 
			telegram_login, 
			first_name, 
			last_name, 
			token, 
			token_expiration, 
			is_admin
		FROM users
		WHERE id = ?
	`
	
	row := DB.QueryRow(query, userID)
	
	err := row.Scan(
		&user.TelegramID,
		&user.TelegramLogin,
		&user.FirstName,
		&user.LastName,
		&user.Token,
		&tokenExpiration,
		&isAdmin,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("error getting user: %v", err)
	}
	
	// Handle nullable token_expiration
	if tokenExpiration.Valid {
		user.TokenExpiration = &tokenExpiration.Time
	}
	
	user.IsAdmin = isAdmin
	
	return &user, nil
}