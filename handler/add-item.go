package handler
import (

	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"
	"io"
	"strconv"

	"main/storage"

)


func AddItemToCatalog(w http.ResponseWriter, r *http.Request) {
	// Проверяем авторизацию и права администратора
	tokenCookie, err := r.Cookie("auth_token")
	if err != nil || tokenCookie.Value == "" {
		http.Redirect(w, r, "/auth", http.StatusSeeOther)
		return
	}
	
	var isAdmin bool
	err = storage.DB.QueryRow("SELECT is_admin FROM users WHERE token = ? AND token_expiration > NOW()", tokenCookie.Value).Scan(&isAdmin)
	if err != nil || !isAdmin {
		http.Error(w, "Доступ запрещен", http.StatusForbidden)
		return
	}
	
	if r.Method != "POST" {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}
	
	// Обрабатываем форму
	err = r.ParseMultipartForm(10 << 20) // 10 MB max
	if err != nil {
		http.Error(w, "Ошибка при обработке формы", http.StatusBadRequest)
		return
	}
	
	// Получаем данные из формы
	name := r.FormValue("name")
    category := r.FormValue("category")
    subcategory := r.FormValue("subcategory")
    description := r.FormValue("description")
    price := r.FormValue("price")
    inStock := r.FormValue("in_stock") == "on"

 	// Дополнительная валидация на стороне сервера
    if category == "Фурнитура" && subcategory == "" {
        http.Error(w, "Для фурнитуры необходимо указать подкатегорию", http.StatusBadRequest)
        return
    }
    if category != "Фурнитура" && subcategory != "" {
        http.Error(w, "Подкатегории разрешены только для фурнитуры", http.StatusBadRequest)
        return
    }
	
	// Обрабатываем файл изображения
	file, handler, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "Ошибка при загрузке изображения", http.StatusBadRequest)
		return
	}
	defer file.Close()
	
	// Создаем папку для изображений, если ее нет
	imageDir := filepath.Join("front", "static", "images", "products")
	if _, err := os.Stat(imageDir); os.IsNotExist(err) {
		os.MkdirAll(imageDir, 0755)
	}
	
		// Генерируем уникальное имя файла
	ext := filepath.Ext(handler.Filename)
	imageName := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	imagePath := filepath.Join(imageDir, imageName)
	
	// Сохраняем файл
	f, err := os.OpenFile(imagePath, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		http.Error(w, "Ошибка при сохранении изображения", http.StatusInternalServerError)
		return
	}
	defer f.Close()
	io.Copy(f, file)
	
	 // Вызываем хранимую процедуру
	var productID int
	err = storage.DB.QueryRowContext(r.Context(), 
		"CALL AddProduct(?, ?, ?, ?, ?, ?, ?)",
		name,
		category,
		subcategory,
		description,
		price,
		"/static/images/products/"+imageName,
		inStock,
	).Scan(&productID)
 
	if err != nil {
		http.Error(w, "Ошибка при сохранении товара: "+err.Error(), http.StatusInternalServerError)
		return
	}
 
	http.Redirect(w, r, "/profile?success=product_added&id="+strconv.Itoa(productID), http.StatusSeeOther)
}