package models

type Product struct {
    ID           int     `json:"id"`
    Name         string  `json:"name"`
    Description  string  `json:"description,omitempty"`
    Price        float64 `json:"price"`
    ImageURL     string  `json:"image_url,omitempty"`
    InStock      bool    `json:"in_stock"`
    CategoryID   int     `json:"category_id,omitempty"`
    SubcategoryID *int   `json:"subcategory_id,omitempty"` // Теперь это указатель, чтобы можно было передавать nil
}