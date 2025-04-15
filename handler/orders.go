package handler

import (
	"encoding/json"
    "log"
	"net/http"
    "main/storage"
    "main/models"
    "fmt"
    "strings"
    "strconv"
    "os"

    "github.com/joho/godotenv"
    tgbot "main/bot"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var managerGroupID int64

// GetUserOrders возвращает активные заказы пользователя
func GetUserOrdersHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")

    userID := GetCurrentUserID(r)
    if userID == 0 {
        w.WriteHeader(http.StatusUnauthorized)
        json.NewEncoder(w).Encode(map[string]string{"error": "Unauthorized"})
        return
    }

    orders, err := storage.GetUserOrders(userID)
    if err != nil {
        log.Printf("Database error: %v", err)
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(map[string]string{"error": "Failed to load orders"})
        return
    }

    // Дополнительная проверка перед отправкой
    for i := range orders {
        if orders[i].Product == nil {
            log.Printf("WARNING: Order %d has nil Product, creating empty one", orders[i].ID)
            orders[i].Product = &models.Product{
                Name:     "Unknown product",
                Price:    0,
                ImageURL: "/static/images/no-image.png",
            }
        }
    }

    if err := json.NewEncoder(w).Encode(orders); err != nil {
        log.Printf("Error encoding response: %v", err)
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(map[string]string{"error": "Failed to encode response"})
    }
}
// GetProductImageURL возвращает URL изображения для продукта
func GetProductImageURL(productID int) (string, error) {
    var imageURL string
    err := storage.DB.QueryRow("SELECT image_url FROM products WHERE id = ?", productID).Scan(&imageURL)
    if err != nil {
        return "", fmt.Errorf("error getting image URL for product %d: %v", productID, err)
    }
    
    // Если URL пустой, возвращаем стандартное изображение
    if imageURL == "" {
        return "/static/images/no-image.png", nil
    }
    
    return imageURL, nil
}

// CheckoutUserOrders оформляет все активные заказы пользователя
func CheckoutUserOrders(userID int) error {
    _, err := storage.DB.Exec("UPDATE orders SET is_active = FALSE WHERE user_id = ? AND is_active = TRUE", userID)
    return err
}
func DeleteOrderHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodDelete {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    // Получаем ID пользователя
    userID := GetCurrentUserID(r)
    if userID == 0 {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    // Получаем ID заказа из URL
    orderIDStr := strings.TrimPrefix(r.URL.Path, "/api/orders/")
    orderID, err := strconv.Atoi(orderIDStr)
    if err != nil {
        http.Error(w, "Invalid order ID", http.StatusBadRequest)
        return
    }

    // Удаляем заказ из базы данных
    err = storage.DeleteOrder(userID, orderID)
    if err != nil {
        http.Error(w, "Failed to delete order", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
    w.Write([]byte("Order deleted successfully"))
}


func CheckoutOrder(w http.ResponseWriter, r *http.Request) {

    err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	managerGroupID, err = strconv.ParseInt(os.Getenv("TELEGRAM_MANAGER_GROUP_ID"), 10, 64)
	if err != nil {
		log.Fatal("Invalid TELEGRAM_MANAGER_GROUP_ID in .env file")
	}

    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    // Get user ID from session
    userID := GetCurrentUserID(r)
    if userID == 0 {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    // Get user's orders from database using existing handler function
    orders, err := storage.GetUserOrders(userID)
    if err != nil {
        log.Printf("Database error: %v", err)
        http.Error(w, "Failed to get orders", http.StatusInternalServerError)
        return
    }

    if len(orders) == 0 {
        http.Error(w, "Cart is empty", http.StatusBadRequest)
        return
    }

    // Get user info
    user, err := storage.GetUserByID(int64(userID))
    if err != nil {
        log.Printf("Failed to get user info: %v", err)
        http.Error(w, "Failed to get user info", http.StatusInternalServerError)
        return
    }

    // Format order message for Telegram
    message := FormatOrderMessage(user, orders)

    // Send to manager group
    msg := tgbotapi.NewMessage(managerGroupID, message)
    msg.ParseMode = "HTML"
    if _, err := tgbot.Bot.Send(msg); err != nil {
        log.Printf("Failed to send order to manager: %v", err)
        // Continue processing even if Telegram message fails
    }

    // Checkout orders using existing function
    if err := CheckoutUserOrders(userID); err != nil {
        log.Printf("Failed to checkout user orders: %v", err)
        http.Error(w, "Failed to process order", http.StatusInternalServerError)
        return
    }

    // Respond to client
    w.Header().Set("Content-Type", "application/json")
    if err := json.NewEncoder(w).Encode(map[string]interface{}{
        "success": true,
        "message": "Order processed successfully",
    }); err != nil {
        log.Printf("Error encoding response: %v", err)
        http.Error(w, "Failed to encode response", http.StatusInternalServerError)
    }
}

func GetUserOrders(userID int64) ([]models.Order, error) {
	var orders []models.Order
	query := `
        SELECT o.id, p.id, p.name, p.price, p.description, p.image_url 
        FROM orders o
        JOIN products p ON o.product_id = p.id
        WHERE o.user_id = ? AND o.is_active = true
    `
	rows, err := storage.DB.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var order models.Order
		var product models.Product
		if err := rows.Scan(
			&order.ID,
			&product.ID,
			&product.Name,
			&product.Price,
			&product.Description,
			&product.ImageURL,
		); err != nil {
			return nil, err
		}
		// Assign pointer to the product
		order.Product = &product
		orders = append(orders, order)
	}

	return orders, nil
}

// ClearUserCart removes all orders for a user
func ClearUserCart(userID int64) error {
	_, err := storage.DB.Exec("UPDATE orders SET is_active = false WHERE user_id = ? AND is_active = true", userID)
	return err
}