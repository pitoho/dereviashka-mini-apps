document.addEventListener('DOMContentLoaded', function() {
    loadUserOrders();
    
    document.getElementById('checkoutBtn').addEventListener('click', checkoutOrder);

    async function loadUserOrders() {
        try {
            const response = await fetch('/api/orders', {
                credentials: 'include',
                headers: {
                    'Accept': 'application/json'
                }
            });
    
            if (!response.ok) {
                throw new Error(`Server error: ${response.status}`);
            }
    
            const data = await response.json();
            console.log("Received data:", data);
            
            // Проверка на пустой или невалидный ответ
            if (!data || (Array.isArray(data) && data.length === 0)) {
                showEmptyCart();
                return;
            }
            
            if (!Array.isArray(data)) {
                throw new Error("Invalid data format: expected array");
            }
    
            renderOrders(data);
            updateCheckoutButton();
            
        } catch (error) {
            console.error('Loading error:', error);
            showError(error.message);
        }
    }

    function showEmptyCart() {
        const ordersList = document.getElementById('ordersList');
        const emptyMessage = document.getElementById('emptyCartMessage');
        
        ordersList.innerHTML = '';
        emptyMessage.style.display = 'block';
        updateCheckoutButton();
    }

    document.getElementById('ordersList').addEventListener('click', function(e) {
        if (e.target.classList.contains('remove-order') || 
            e.target.parentElement.classList.contains('remove-order')) {
            const button = e.target.classList.contains('remove-order') ? 
                          e.target : e.target.parentElement;
            removeOrder(button);
        }
    });
    
    function renderOrders(orders) {
        const ordersList = document.getElementById('ordersList');
        const emptyMessage = document.getElementById('emptyCartMessage');
        
        ordersList.innerHTML = '';
        
        if (!orders || orders.length === 0) {
            showEmptyCart();
            return;
        }
        
        emptyMessage.style.display = 'none';

        orders.forEach(order => {
            const product = order.product || {
                name: "Неизвестный товар",
                price: 0,
                image_url: "/static/images/no-image.png",
                description: ""
            };
    
            const orderItem = document.createElement('div');
            orderItem.className = 'order-item';

            const formattedPrice = new Intl.NumberFormat('ru-RU', {
                style: 'currency',
                currency: 'RUB',
                minimumFractionDigits: 0
            }).format(product.price || 0);
    
            orderItem.innerHTML = `
                <div class="order-image">
                    <img src="${product.image_url || '/static/images/no-image.png'}" 
                         alt="${product.name}" 
                         onerror="this.src='/static/images/no-image.png'">
                </div>
                <div class="order-info">
                    <h3>${product.name || 'Без названия'}</h3>
                    ${product.description ? `<p class="product-description">${product.description}</p>` : ''}
                    <div class="order-price">${formattedPrice}</div>
                </div>
                <button class="remove-order" data-order-id="${order.id}">×</button>
            `;
            
            ordersList.appendChild(orderItem);
        });
        
        updateCheckoutButton();
    }

    function updateCheckoutButton() {
        const checkoutBtn = document.getElementById('checkoutBtn');
        const isEmpty = document.getElementById('ordersList').children.length === 0;
        checkoutBtn.disabled = isEmpty;
        checkoutBtn.classList.toggle('disabled', isEmpty);
    }

    function showError(message) {
        const ordersList = document.getElementById('ordersList');
        ordersList.innerHTML = `
            <div class="error-message">
                <p>⚠️ Произошла ошибка при загрузке корзины</p>
                <p><small>${message || 'Неизвестная ошибка'}</small></p>
                <button onclick="location.reload()">Попробовать снова</button>
            </div>
        `;
        updateCheckoutButton();
    }
    
    async function checkoutOrder() {
        if (this.disabled) return;
        
        if (!confirm('Вы уверены, что хотите оформить заказ?')) return;
        
        try {
            const response = await fetch('/api/orders/checkout', {
                method: 'POST',
                credentials: 'include'
            });

            if (!response.ok) {
                const errorData = await response.json().catch(() => null);
                throw new Error(errorData?.message || 'Ошибка оформления заказа');
            }

            alert('Заказ успешно оформлен!');
            showEmptyCart();
            
        } catch (error) {
            console.error('Ошибка оформления:', error);
            alert(`Ошибка: ${error.message}`);
        }
    }

    async function removeOrder(button) {
        if (!confirm('Удалить товар из корзины?')) return;
        
        try {
            const orderId = button.getAttribute('data-order-id');
            const response = await fetch(`/api/orders/${orderId}`, {
                method: 'DELETE',
                credentials: 'include'
            });

            if (!response.ok) {
                throw new Error('Ошибка удаления товара');
            }

            button.closest('.order-item').remove();
            
            const ordersList = document.getElementById('ordersList');
            if (ordersList.children.length === 0) {
                showEmptyCart();
            } else {
                updateCheckoutButton();
            }
            
        } catch (error) {
            console.error('Ошибка удаления:', error);
            alert(`Ошибка: ${error.message}`);
        }
    }

    
});