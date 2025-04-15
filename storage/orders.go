package storage

import (
	"database/sql"
	"fmt"

	// "log"
	"main/models"
)

func GetUserOrders(userID int) ([]models.Order, error) {
    if DB == nil {
        return nil, fmt.Errorf("database connection is not initialized")
    }

    query := `
        SELECT 
            o.id, o.user_id, o.product_id, o.is_active, o.created_at,
            p.id, p.name, p.description, p.price, p.image_path, p.in_stock, 
            p.category_id, p.subcategory_id
        FROM orders o
        JOIN products p ON o.product_id = p.id
        WHERE o.user_id = ? AND o.is_active = TRUE
        ORDER BY o.created_at DESC`

    rows, err := DB.Query(query, userID)
    if err != nil {
        return nil, fmt.Errorf("query execution failed: %v", err)
    }
    defer rows.Close()

    var orders []models.Order

    for rows.Next() {
        var order models.Order
        var product models.Product
        var subcategoryID sql.NullInt64

        err := rows.Scan(
            &order.ID,
            &order.UserID,
            &order.ProductID,
            &order.IsActive,
            &order.CreatedAt,
            &product.ID,
            &product.Name,
            &product.Description,
            &product.Price,
            &product.ImageURL,
            &product.InStock,
            &product.CategoryID,
            &subcategoryID,
        )
        if err != nil {
            return nil, fmt.Errorf("row scan failed: %v", err)
        }

        // Обработка subcategory_id
        if subcategoryID.Valid {
            val := int(subcategoryID.Int64)
            product.SubcategoryID = &val
        }

        // Важно: создаём новый указатель для каждого продукта
        productCopy := product
        order.Product = &productCopy
        
        orders = append(orders, order)
    }

    if err := rows.Err(); err != nil {
        return nil, fmt.Errorf("rows iteration error: %v", err)
    }

    return orders, nil
}
func DeleteOrder(userID, orderID int) error {
    // Используем один запрос с проверкой принадлежности заказа пользователю
    result, err := DB.Exec(`
        DELETE FROM orders 
        WHERE id = ? AND user_id = ?`,
        orderID, userID)
    
    if err != nil {
        return fmt.Errorf("database error: %v", err)
    }

    rowsAffected, err := result.RowsAffected()
    if err != nil {
        return fmt.Errorf("failed to get rows affected: %v", err)
    }

    if rowsAffected == 0 {
        return fmt.Errorf("order not found or access denied")
    }

    return nil
}