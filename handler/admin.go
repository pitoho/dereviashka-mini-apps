package handler
import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"strings"

	// "main/handler"
	"main/models"

)
func ServeAdminProfile(w http.ResponseWriter, r *http.Request, user models.UserInfo) {
    // Читаем шаблон страницы
    path := filepath.Join("front", "pages", "admin_profile.html")
    content, err := os.ReadFile(path)
    if err != nil {
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        log.Printf("Error reading admin profile template: %v", err)
        return
    }

    // Проверяем параметр success
    successMsg := ""
    if r.URL.Query().Get("success") == "product_added" {
        successMsg = `<div class="success">Товар успешно добавлен!</div>`
    }

    // Заменяем плейсхолдеры в шаблоне
    html := string(content)
    html = strings.Replace(html, "<!-- SUCCESS_MESSAGE -->", successMsg, 1)

    w.Header().Set("Content-Type", "text/html")
    w.WriteHeader(http.StatusOK)
    w.Write([]byte(html))
}