package handler

import (
    "database/sql"
    "encoding/json"
    "net/http"
    "strconv"
    "log"

    "main/models"
    "main/storage"
)

// GetCatalogCategories - для получения категорий каталога товаров
func GetCatalogCategories(w http.ResponseWriter, r *http.Request) {
    log.Println("GetCatalogCategories handler started")
    defer log.Println("GetCatalogCategories handler completed")

    db := storage.GetDB()
    if db == nil {
        log.Println("Database connection is nil")
        http.Error(w, "Database connection error", http.StatusInternalServerError)
        return
    }

    rows, err := db.Query(`
        SELECT id, name, has_subcategories 
        FROM categories
        ORDER BY name
    `)
    if err != nil {
        log.Printf("Database query error: %v", err)
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    var categories []models.Category
    for rows.Next() {
        var c models.Category
        if err := rows.Scan(&c.ID, &c.Name, &c.HasSubcategories); err != nil {
            log.Printf("Row scan error: %v", err)
            http.Error(w, "Internal Server Error", http.StatusInternalServerError)
            return
        }
        categories = append(categories, c)
    }

    w.Header().Set("Content-Type", "application/json")
    if err := json.NewEncoder(w).Encode(categories); err != nil {
        log.Printf("JSON encode error: %v", err)
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
    }
}

// GetCatalogSubcategories - для получения подкатегорий каталога товаров
func GetCatalogSubcategories(w http.ResponseWriter, r *http.Request) {
    // В вашей структуре БД подкатегории не привязаны к категориям
    // Поэтому просто возвращаем все подкатегории
    
    rows, err := storage.DB.Query(`
        SELECT id, name 
        FROM subcategories
        ORDER BY name
    `)
    if err != nil {
        log.Printf("Database query error: %v", err)
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    var subcategories []models.Subcategory
    for rows.Next() {
        var s models.Subcategory
        if err := rows.Scan(&s.ID, &s.Name); err != nil {
            log.Printf("Row scan error: %v", err)
            http.Error(w, "Internal Server Error", http.StatusInternalServerError)
            return
        }
        subcategories = append(subcategories, s)
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(subcategories)
}

// GetCatalogProducts - для получения товаров каталога
func GetCatalogProducts(w http.ResponseWriter, r *http.Request) {
    categoryID := r.URL.Query().Get("category_id")
    subcategoryID := r.URL.Query().Get("subcategory_id")

    var query string
    var args []interface{}

    if subcategoryID != "" {
        sID, err := strconv.Atoi(subcategoryID)
        if err != nil {
            http.Error(w, "Invalid subcategory_id", http.StatusBadRequest)
            return
        }
        query = `
            SELECT p.id, p.name, p.price, p.image_path, p.in_stock, p.category_id, p.subcategory_id
            FROM products p
            WHERE p.subcategory_id = ?
        `
        args = append(args, sID)
    } else if categoryID != "" {
        cID, err := strconv.Atoi(categoryID)
        if err != nil {
            http.Error(w, "Invalid category_id", http.StatusBadRequest)
            return
        }
        query = `
            SELECT p.id, p.name, p.price, p.image_path, p.in_stock, p.category_id, p.subcategory_id
            FROM products p
            WHERE p.category_id = ?
        `
        args = append(args, cID)
    } else {
        http.Error(w, "Either category_id or subcategory_id must be provided", http.StatusBadRequest)
        return
    }

    rows, err := storage.DB.Query(query, args...)
    if err != nil {
        log.Printf("Database query error: %v", err)
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    var products []models.Product
    for rows.Next() {
        var p models.Product
        if err := rows.Scan(&p.ID, &p.Name, &p.Price, &p.ImageURL, &p.InStock, &p.CategoryID, &p.SubcategoryID); err != nil {
            log.Printf("Row scan error: %v", err)
            http.Error(w, "Internal Server Error", http.StatusInternalServerError)
            return
        }
        products = append(products, p)
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(products)
}

// GetProductDetails - для получения полной информации о товаре
func GetProductDetails(w http.ResponseWriter, r *http.Request) {
    productID := r.URL.Query().Get("id")
    if productID == "" {
        http.Error(w, "Product ID is required", http.StatusBadRequest)
        return
    }

    db := storage.GetDB()
    var product models.Product
    var subcatID sql.NullInt64 // Для обработки NULL в subcategory_id

    err := db.QueryRow(`
        SELECT id, name, description, price, image_path, in_stock, 
               category_id, subcategory_id
        FROM products 
        WHERE id = ?
    `, productID).Scan(
        &product.ID, &product.Name, &product.Description, &product.Price,
        &product.ImageURL, &product.InStock, &product.CategoryID,
        &subcatID,
    )

    if err != nil {
        log.Printf("Error fetching product details: %v", err)
        http.Error(w, "Product not found", http.StatusNotFound)
        return
    }

    // Обработка NULL значения для подкатегории
    if subcatID.Valid {
        subcatIDInt := int(subcatID.Int64)
        product.SubcategoryID = &subcatIDInt
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(product)
}

// AddToCart - добавление товара в корзину
func AddToCart(w http.ResponseWriter, r *http.Request) {
    userID := GetCurrentUserID(r)
    if userID == 0 {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    var request struct {
        ProductID int `json:"product_id"`
    }
    
    if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
        http.Error(w, "Invalid request", http.StatusBadRequest)
        return
    }

    // Проверяем существование товара
    var product models.Product
    err := storage.DB.QueryRow(`
        SELECT id, name, price FROM products WHERE id = ?
    `, request.ProductID).Scan(&product.ID, &product.Name, &product.Price)
    
    if err != nil {
        http.Error(w, "Product not found", http.StatusNotFound)
        return
    }

    // Добавляем в корзину
    _, err = storage.DB.Exec(`
        INSERT INTO orders (user_id, product_id, product_name, product_price)
        VALUES (?, ?, ?, ?)
    `, userID, product.ID, product.Name, product.Price)
    
    if err != nil {
        log.Printf("Error adding to cart: %v", err)
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(map[string]string{"status": "added"})
}

