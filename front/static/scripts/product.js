// Функция для показа уведомлений
function showNotification(message, type = 'success') {
    const container = document.getElementById('notificationContainer');
    const notification = document.createElement('div');
    notification.className = `notification ${type}`;
    notification.textContent = message;
    container.appendChild(notification);
    
    setTimeout(() => {
        notification.classList.add('show');
    }, 10);
    
    setTimeout(() => {
        notification.classList.remove('show');
        setTimeout(() => {
            container.removeChild(notification);
        }, 300);
    }, 2000);
}

// Функция для обработки 401 ошибки (неавторизован)
function handleUnauthorized() {
    showNotification('Для выполнения действия необходимо авторизоваться', 'error');
}

// Функция проверки наличия товара в корзине
async function isProductInCart(productId) {
    try {
        const response = await fetch('/api/orders', {
            credentials: 'include',
            headers: {
                'Accept': 'application/json'
            }
        });

        if (response.status === 401) {
            handleUnauthorized();
            return Promise.reject('Unauthorized');
        }

        if (!response.ok) {
            throw new Error('Ошибка при проверке корзины');
        }

        const cartItems = await response.json();
        return Array.isArray(cartItems) && cartItems.some(item => item.product_id === productId);
    } catch (error) {
        console.error('Error checking cart:', error);
        return false;
    }
}

document.addEventListener('DOMContentLoaded', function() {
    const urlParams = new URLSearchParams(window.location.search);
    const productId = urlParams.get('id');
    const addToCartBtn = document.getElementById('addToCart');

    addToCartBtn.addEventListener('click', async function() {
        // Проверяем наличие товара в корзине
        try {
            const alreadyInCart = await isProductInCart(parseInt(productId));
            
            if (alreadyInCart) {
                showNotification('Этот товар уже есть в вашей корзине', 'warning');
                return;
            }

            // Если товара нет в корзине, добавляем его
            const response = await fetch('/api/cart/add', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                credentials: 'include',
                body: JSON.stringify({ product_id: parseInt(productId) })
            });

            if (response.status === 401) {
                handleUnauthorized();
                return;
            }

            if (!response.ok) {
                const errorData = await response.json();
                throw new Error(errorData.message || 'Ошибка добавления в корзину');
            }

            showNotification('Товар успешно добавлен в корзину', 'success');
        } catch (error) {
            if (error !== 'Unauthorized') {
                console.error('Error:', error);
                showNotification(error.message || 'Произошла ошибка при добавлении в корзину', 'error');
            }
        }
    });
    
    if (!productId) {
        showNotification('Не указан ID товара', 'error');
        setTimeout(() => {
            window.location.href = '/catalog';
        }, 2000);
        return;
    }

    fetch(`/api/catalog/product?id=${productId}`)
        .then(response => {
            if (response.status === 401) {
                return Promise.reject('Unauthorized');
            }
            if (!response.ok) {
                return response.json().then(err => {
                    throw new Error(err.message || 'Товар не найден');
                });
            }
            return response.json();
        })
        .then(product => {
            document.getElementById('productName').textContent = product.name;
            document.getElementById('productNameTitle').textContent = product.name;
            document.getElementById('productDescription').textContent = 
                product.description || 'Нет описания';
            document.getElementById('productPrice').textContent = 
                `${product.price.toLocaleString('ru-RU')} ₽`;
            
            const stockElement = document.getElementById('productStock');
            stockElement.textContent = product.in_stock ? 'В наличии' : 'Нет в наличии';
            stockElement.classList.add(product.in_stock ? 'in-stock' : 'out-of-stock');
            
            addToCartBtn.disabled = !product.in_stock;
            
            const productImage = document.getElementById('productImage');
            productImage.src = product.image_url || '../static/images/no-image.png';
            productImage.alt = product.name;
            
            if (product.category_id) {
                fetch(`/api/catalog/categories`)
                    .then(response => {
                        if (response.status === 401) {
                            return Promise.reject('Unauthorized');
                        }
                        if (!response.ok) {
                            throw new Error('Ошибка загрузки категорий');
                        }
                        return response.json();
                    })
                    .then(categories => {
                        const category = categories.find(c => c.id === product.category_id);
                        if (category) {
                            document.getElementById('productCategory').textContent = category.name;
                        }
                    })
                    .catch(error => {
                        if (error === 'Unauthorized') {
                            handleUnauthorized();
                        } else {
                            console.error('Error loading categories:', error);
                            document.getElementById('productCategory').textContent = 'Неизвестная категория';
                        }
                    });
            }
        })
        .catch(error => {
            if (error !== 'Unauthorized') {
                console.error('Error:', error);
                showNotification(error.message || 'Произошла ошибка при загрузке товара', 'error');
                setTimeout(() => {
                    window.location.href = '/catalog';
                }, 2000);
            }
        });
});