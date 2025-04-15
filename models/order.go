package models

import "time"

type Order struct {
    ID          int       `json:"id"`
    UserID      int       `json:"user_id"`
    ProductID   int       `json:"product_id"`
    Quantity    int       `json:"quantity"`
    IsActive    bool      `json:"is_active"`
    CreatedAt   time.Time `json:"created_at"`
    
    // Связь с продуктом
    Product     *Product  `json:"product,omitempty" gorm:"foreignKey:ProductID"`
}