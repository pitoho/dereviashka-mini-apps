package handler

import (
	"net/http"
	"encoding/json"

	"main/models"
	"main/storage"

)
func GetCategory(w http.ResponseWriter, r *http.Request) {
	rows, err := storage.DB.Query("SELECT name, has_subcategories FROM categories ORDER BY name")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var categories []models.Category

	for rows.Next() {
		var c models.Category
		if err := rows.Scan(&c.Name, &c.HasSubcategories); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		categories = append(categories, c)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(categories)
}

func GetSubcategory(w http.ResponseWriter, r *http.Request) {
	// Для новой структуры БД (без category_id в subcategories)
	rows, err := storage.DB.Query("SELECT name FROM subcategories ORDER BY name")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var subcategories []models.Subcategory

	for rows.Next() {
		var s models.Subcategory
		if err := rows.Scan(&s.Name); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		subcategories = append(subcategories, s)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(subcategories)
}
