package handler
import (

	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"
	"encoding/json"
	"database/sql"
	"io"


	"main/models"
	"main/storage"
)
func SearchProduct(w http.ResponseWriter, r *http.Request) {
    name := r.URL.Query().Get("name")
    if name == "" {
        http.Error(w, "Name parameter is required", http.StatusBadRequest)
        return
    }

    // Create a product instance
    var product models.Product

    err := storage.DB.QueryRow(`
        SELECT 
            p.id, 
            p.name, 
            c.name as category, 
            s.name as subcategory, 
            p.description, 
            p.price, 
            p.image_path, 
            p.in_stock
        FROM products p
        LEFT JOIN categories c ON p.category_id = c.id
        LEFT JOIN subcategories s ON p.subcategory_id = s.id
        WHERE p.name LIKE ? LIMIT 1`, "%"+name+"%").
        Scan(
            &product.ID, 
            &product.Name, 
            &product.CategoryID, 
            &product.SubcategoryID,
            &product.Description, 
            &product.Price, 
            &product.ImageURL, 
            &product.InStock,
        )

    if err != nil {
        if err == sql.ErrNoRows {
            http.Error(w, "Product not found", http.StatusNotFound)
        } else {
            http.Error(w, err.Error(), http.StatusInternalServerError)
        }
        return
    }

    // Create response
    response := map[string]interface{}{
        "id":          product.ID,
        "name":        product.Name,
        "category":    product.CategoryID,
        "subcategory": "",
        "description": product.Description,
        "price":       product.Price,
        "image_path":  product.ImageURL,
        "in_stock":    product.InStock,
    }

    if product.SubcategoryID != nil {
        response["subcategory"] = *product.SubcategoryID
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}

func UpdateProduct(w http.ResponseWriter, r *http.Request) {
	// Проверка авторизации администратора...
	
	err := r.ParseMultipartForm(10 << 20) // 10 MB
	if err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}

	// Получение данных формы
	id := r.FormValue("id")
	name := r.FormValue("name")
	category := r.FormValue("category")
	subcategory := r.FormValue("subcategory")
	description := r.FormValue("description")
	price := r.FormValue("price")
	inStock := r.FormValue("in_stock") == "on"

	// Обработка изображения (если загружено новое)
	var imagePath string
	file, handler, err := r.FormFile("image")
	if err == nil {
		defer file.Close()
		
		// Сохранение нового изображения
		ext := filepath.Ext(handler.Filename)
		imageName := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
		imagePath = filepath.Join("static", "images", "products", imageName)
		
		f, err := os.Create(imagePath)
		if err != nil {
			http.Error(w, "Error saving image", http.StatusInternalServerError)
			return
		}
		defer f.Close()
		
		if _, err := io.Copy(f, file); err != nil {
			http.Error(w, "Error saving image", http.StatusInternalServerError)
			return
		}
		
		imagePath = "/" + imagePath // Для сохранения в БД
	}

	// Обновление товара в БД
	if imagePath != "" {
		_, err = storage.DB.Exec(`
			UPDATE products 
			SET name = ?, category_id = (SELECT id FROM categories WHERE name = ?),
				subcategory_id = (SELECT id FROM subcategories WHERE name = ?),
				description = ?, price = ?, image_path = ?, in_stock = ?
			WHERE id = ?`,
			name, category, subcategory, description, price, imagePath, inStock, id)
	} else {
		_, err = storage.DB.Exec(`
			UPDATE products 
			SET name = ?, category_id = (SELECT id FROM categories WHERE name = ?),
				subcategory_id = (SELECT id FROM subcategories WHERE name = ?),
				description = ?, price = ?, in_stock = ?
			WHERE id = ?`,
			name, category, subcategory, description, price, inStock, id)
	}

	if err != nil {
		http.Error(w, "Error updating product: "+err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/profile?success=product_updated", http.StatusSeeOther)
}