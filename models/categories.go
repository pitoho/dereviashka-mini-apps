package models

type Category struct {
	ID               int   `gorm:"primaryKey" json:"id"`
	Name             string `gorm:"unique;not null" json:"name"`
	HasSubcategories bool   `gorm:"default:false" json:"has_subcategories"`
}

type Subcategory struct {
	ID         int   `gorm:"primaryKey" json:"id"`
	Name       string `gorm:"unique;not null" json:"name"`
	CategoryID int   `json:"category_id"`
}