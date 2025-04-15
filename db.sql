CREATE DATABASE dereviashka;
use dereviashka;
CREATE TABLE users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    telegram_id BIGINT UNIQUE,
    telegram_login VARCHAR(255),
    first_name VARCHAR(255),
    last_name VARCHAR(255),
    token VARCHAR(255) UNIQUE,
    token_expiration DATETIME,
    is_admin BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS products (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    category VARCHAR(100) NOT NULL,
    description TEXT NOT NULL,
    price DECIMAL(10, 2) NOT NULL,
    image_path VARCHAR(255) NOT NULL,
    in_stock BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);
ALTER TABLE products ADD COLUMN subcategory VARCHAR(100) AFTER category;

DELIMITER //
CREATE PROCEDURE upsert_user(
    IN p_telegram_id BIGINT,
    IN p_telegram_login VARCHAR(255),
    IN p_first_name VARCHAR(255),
    IN p_last_name VARCHAR(255),
    IN p_token VARCHAR(255),
    IN p_token_expiration DATETIME
)
BEGIN
    INSERT INTO users (telegram_id, telegram_login, first_name, last_name, token, token_expiration)
    VALUES (p_telegram_id, p_telegram_login, p_first_name, p_last_name, p_token, p_token_expiration)
    ON DUPLICATE KEY UPDATE
        telegram_login = p_telegram_login,
        first_name = p_first_name,
        last_name = p_last_name,
        token = p_token,
        token_expiration = p_token_expiration;
END //
DELIMITER ;

DELIMITER //
CREATE PROCEDURE check_token(
    IN p_token VARCHAR(255),
    OUT p_user_id INT,
    OUT p_is_admin BOOLEAN,
    OUT p_is_valid BOOLEAN
)
BEGIN
    DECLARE v_expiration DATETIME;
    
    SELECT id, is_admin, token_expiration INTO p_user_id, p_is_admin, v_expiration
    FROM users
    WHERE token = p_token;
    
    IF p_user_id IS NOT NULL AND v_expiration > NOW() THEN
        SET p_is_valid = TRUE;
    ELSE
        SET p_is_valid = FALSE;
    END IF;
END //
DELIMITER ;

DELIMITER //
CREATE PROCEDURE cleanup_expired_tokens()
BEGIN
    UPDATE users
    SET token = NULL, token_expiration = NULL
    WHERE token_expiration < NOW();
END //
DELIMITER ;


/*товары*/

select * from products;
select * from categories;
select * from subcategories;


-- Таблица категорий
CREATE TABLE IF NOT EXISTS categories (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    has_subcategories BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Таблица подкатегорий (только для фурнитуры)
CREATE TABLE IF NOT EXISTS subcategories (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Таблица продуктов
CREATE TABLE IF NOT EXISTS products (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    category_id INT NOT NULL,
    subcategory_id INT NULL,
    description TEXT NOT NULL,
    price DECIMAL(10, 2) NOT NULL,
    image_path VARCHAR(255) NOT NULL,
    in_stock BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (category_id) REFERENCES categories(id),
    FOREIGN KEY (subcategory_id) REFERENCES subcategories(id)
);
-- Вставка основных категорий
INSERT IGNORE INTO categories (name, has_subcategories) VALUES 
('Натуральный шпон', FALSE),
('Экошпон', FALSE),
('Эмаль', FALSE),
('Фурнитура', TRUE), -- Единственная категория с подкатегориями
('Входные двери', FALSE);

-- Вставка подкатегорий для фурнитуры
INSERT IGNORE INTO subcategories (name) VALUES
('Замки'),
('Дверные ручки'),
('Петли'),
('Доводчики'),
('Другие элементы');

DELIMITER //

CREATE PROCEDURE AddProduct(
    IN p_name VARCHAR(255),
    IN p_category_name VARCHAR(100),
    IN p_subcategory_name VARCHAR(100),
    IN p_description TEXT,
    IN p_price DECIMAL(10, 2),
    IN p_image_path VARCHAR(255),
    IN p_in_stock BOOLEAN
)
BEGIN
    DECLARE v_category_id INT;
    DECLARE v_has_subcategories BOOLEAN;
    DECLARE v_subcategory_id INT DEFAULT NULL;
    DECLARE v_is_furniture BOOLEAN DEFAULT FALSE;
    
    -- Получаем ID категории и информацию о подкатегориях
    SELECT id, has_subcategories, name = 'Фурнитура' 
    INTO v_category_id, v_has_subcategories, v_is_furniture
    FROM categories WHERE name = p_category_name;
    
    -- Если категория не найдена, создаем новую (без подкатегорий)
    IF v_category_id IS NULL THEN
        INSERT INTO categories (name) VALUES (p_category_name);
        SET v_category_id = LAST_INSERT_ID();
        SET v_has_subcategories = FALSE;
        SET v_is_furniture = FALSE;
    END IF;
    
    -- Валидация: подкатегория только для фурнитуры
    IF v_is_furniture THEN
        IF p_subcategory_name IS NULL OR p_subcategory_name = '' THEN
            SIGNAL SQLSTATE '45000' 
            SET MESSAGE_TEXT = 'Для фурнитуры необходимо указать подкатегорию';
        END IF;
        
        -- Ищем подкатегорию
        SELECT id INTO v_subcategory_id FROM subcategories WHERE name = p_subcategory_name;
        
        -- Если не найдена, создаем новую
        IF v_subcategory_id IS NULL THEN
            INSERT INTO subcategories (name) VALUES (p_subcategory_name);
            SET v_subcategory_id = LAST_INSERT_ID();
        END IF;
    ELSE
        -- Для не-фурнитуры подкатегория должна быть NULL
        IF p_subcategory_name IS NOT NULL AND p_subcategory_name != '' THEN
            SIGNAL SQLSTATE '45000' 
            SET MESSAGE_TEXT = 'Подкатегории разрешены только для фурнитуры';
        END IF;
    END IF;
    
    -- Добавляем товар
    INSERT INTO products (
        name, 
        category_id, 
        subcategory_id, 
        description, 
        price, 
        image_path, 
        in_stock
    ) VALUES (
        p_name,
        v_category_id,
        v_subcategory_id,
        p_description,
        p_price,
        p_image_path,
        p_in_stock
    );
    
    SELECT LAST_INSERT_ID() AS new_product_id;
END //

DELIMITER ;

CREATE TABLE orders (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL,
    product_id INT NOT NULL,
    product_name VARCHAR(255) NOT NULL,
    product_price DECIMAL(10,2) NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,  -- Новое поле: TRUE для активных, FALSE для неактивных
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (product_id) REFERENCES products(id)
);
select * from orders;