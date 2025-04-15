package handler
import (
	"fmt"
	"strings"

	"main/models"
)

func FormatOrderMessage(user *models.UserInfo, orders []models.Order) string {
	var total float64
	var message strings.Builder

	message.WriteString(fmt.Sprintf("🛒 <b>Новый заказ</b>\n\n"))
	message.WriteString(fmt.Sprintf("👤 <b>Пользователь:</b> %s %s (@%s)\n",
		user.FirstName, user.LastName, user.TelegramLogin))
	message.WriteString("<b>Товары:</b>\n")

	for _, order := range orders {
		if order.Product != nil {
			product := order.Product
			message.WriteString(fmt.Sprintf("— %s\n", product.Name))
			message.WriteString(fmt.Sprintf("  Цена: %.2f ₽\n", product.Price))
			message.WriteString("  Ссылка:\n")
			message.WriteString(fmt.Sprintf("  http://localhost:8080/product?id=%d\n\n", product.ID))
			total += product.Price
		}
	}

	message.WriteString(fmt.Sprintf("\n<b>Итого:</b> %.2f ₽", total))
	return message.String()
}
