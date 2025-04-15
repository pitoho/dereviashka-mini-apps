package models

import "time"

type UserInfo struct {
    TelegramID      int64
    TelegramLogin   string
    FirstName       string
    LastName        string
    Token           string
    TokenExpiration *time.Time
    IsAdmin         bool
}