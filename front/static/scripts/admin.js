// Загрузка категорий при открытии страницы
document.addEventListener('DOMContentLoaded', function() {
    const accordions = document.querySelectorAll('.accordion');
            
    accordions.forEach(accordion => {
        accordion.addEventListener('click', function() {
            // Закрываем все открытые панели
            document.querySelectorAll('.panel.active').forEach(panel => {
                if (panel !== this.nextElementSibling) {
                    panel.classList.remove('active');
                    panel.style.maxHeight = null;
                    panel.previousElementSibling.classList.remove('active');
                }
            });
            
            // Переключаем текущую панель
            this.classList.toggle('active');
            const panel = this.nextElementSibling;
            
            if (panel.classList.contains('active')) {
                panel.classList.remove('active');
                panel.style.maxHeight = null;
            } else {
                panel.classList.add('active');
                updatePanelHeight(panel);
                
                // Если это панель добавления товара, добавляем обработчик для категорий
                if (panel.querySelector('#category')) {
                    setupCategoryHandlers(panel);
                }
            }
        });
    });

    // Функция для обновления высоты панели
    function updatePanelHeight(panel) {
        const contentHeight = panel.scrollHeight;
        panel.style.maxHeight = contentHeight + 'px';
        
        // Обработчики для изображений
        const images = panel.querySelectorAll('img');
        images.forEach(img => {
            if (!img.hasAttribute('data-resize-listener')) {
                img.addEventListener('load', function() {
                    panel.style.maxHeight = 'none';
                    const newHeight = panel.scrollHeight;
                    panel.style.maxHeight = newHeight + 'px';
                });
                img.setAttribute('data-resize-listener', 'true');
            }
        });
    }

    // Настройка обработчиков для категорий и подкатегорий
    function setupCategoryHandlers(panel) {
        const categorySelect = panel.querySelector('#category');
        const subcategoryGroup = panel.querySelector('#subcategoryGroup');
        
        if (categorySelect && subcategoryGroup) {
            categorySelect.addEventListener('change', function() {
                // Даем время для загрузки/отображения подкатегорий
                setTimeout(() => {
                    updatePanelHeight(panel);
                }, 100);
            });
        }
    }

    // Обработчик для загрузки изображений в форме редактирования
    const editImageInput = document.getElementById('editImage');
    if (editImageInput) {
        editImageInput.addEventListener('change', function(e) {
            const file = e.target.files[0];
            if (file) {
                const reader = new FileReader();
                reader.onload = function(event) {
                    const img = document.getElementById('currentImage');
                    img.src = event.target.result;
                    
                    let panel = img.closest('.panel');
                    if (panel && panel.classList.contains('active')) {
                        setTimeout(() => {
                            updatePanelHeight(panel);
                        }, 100);
                    }
                };
                reader.readAsDataURL(file);
            }
        });
    }

    // Обработчик для загрузки изображений в форме добавления
    const addImageInput = document.getElementById('image');
    if (addImageInput) {
        addImageInput.addEventListener('change', function(e) {
            const file = e.target.files[0];
            if (file) {
                const panel = this.closest('.panel');
                if (panel && panel.classList.contains('active')) {
                    setTimeout(() => {
                        updatePanelHeight(panel);
                    }, 100);
                }
            }
        });
    }
    loadCategories();
    // Добавляем обработчик изменения категории для формы добавления товара
    document.getElementById('category').addEventListener('change', function() {
        const category = this.value;
        const subcategoryGroup = document.getElementById('subcategoryGroup');
        const subcategorySelect = document.getElementById('subcategory');
    
        if (category === 'Фурнитура') {
            fetch('/api/subcategories')
                .then(response => response.json())
                .then(subcategories => {
                    subcategorySelect.innerHTML = 
                        '<option value="">-- Выберите подкатегорию --</option>' +
                        subcategories.map(s => 
                            `<option value="${s.name}">${s.name}</option>`
                        ).join('');
                    subcategoryGroup.style.display = 'block';
                })
                .catch(error => {
                    console.error('Ошибка загрузки подкатегорий:', error);
                    subcategoryGroup.style.display = 'none';
                });
        } else {
            subcategoryGroup.style.display = 'none';
            subcategorySelect.value = '';
        }
    });

});



// Функция для загрузки категорий
function loadCategories() {
    fetch('/api/categories')
        .then(response => response.json())
        .then(categories => {
            const select = document.getElementById('category');
            const editSelect = document.getElementById('editCategory');
            
            select.innerHTML = categories.map(c => 
                `<option value="${c.name}">${c.name}</option>`
            ).join('');
            
            editSelect.innerHTML = categories.map(c => 
                `<option value="${c.name}">${c.name}</option>`
            ).join('');
        });
}

// Поиск товара
function searchProduct() {
    const productName = document.getElementById('searchProduct').value.trim();
    if (!productName) {
        alert('Пожалуйста, введите название товара');
        return;
    }
    
    fetch(`/api/products/search?name=${encodeURIComponent(productName)}`)
        .then(response => {
            if (!response.ok) {
                throw new Error(response.status === 404 ? 'Товар не найден' : 'Ошибка сервера');
            }
            return response.json();
        })
        .then(product => {
            displayProductForEdit(product);
        })
        .catch(error => {
            console.error('Ошибка поиска:', error);
            alert(error.message);
        });
}
// Отображение найденного товара для редактирования
function displayProductForEdit(product) {
    document.getElementById('productId').value = product.id;
    document.getElementById('editName').value = product.name;
    document.getElementById('editDescription').value = product.description;
    document.getElementById('editPrice').value = product.price;
    document.getElementById('editInStock').checked = product.in_stock;
    
    // Установка категории
    const categorySelect = document.getElementById('editCategory');
    for (let i = 0; i < categorySelect.options.length; i++) {
        if (categorySelect.options[i].value === product.category) {
            categorySelect.selectedIndex = i;
            break;
        }
    }
    
    // Обработка подкатегорий
    const subcategoryGroup = document.getElementById('editSubcategoryGroup');
    const subcategorySelect = document.getElementById('editSubcategory');
    
    if (product.category === 'Фурнитура') {
        fetch('/api/subcategories')
            .then(response => response.json())
            .then(subcategories => {
                subcategorySelect.innerHTML = 
                    '<option value="">-- Выберите подкатегорию --</option>' +
                    subcategories.map(s => 
                        `<option value="${s.name}" ${s.name === product.subcategory ? 'selected' : ''}>${s.name}</option>`
                    ).join('');
                subcategoryGroup.style.display = 'block';
                
                // Убедимся, что подкатегория выбрана правильно
                if (product.subcategory) {
                    for (let i = 0; i < subcategorySelect.options.length; i++) {
                        if (subcategorySelect.options[i].value === product.subcategory) {
                            subcategorySelect.selectedIndex = i;
                            break;
                        }
                    }
                }
            });
    } else {
        subcategoryGroup.style.display = 'none';
        subcategorySelect.value = '';
    }
    
    // Отображение текущего изображения
    if (product.image_path) {
        document.getElementById('currentImage').src = product.image_path;
        document.getElementById('currentImage').style.display = 'block';
    } else {
        document.getElementById('currentImage').style.display = 'none';
    }
    
    // Показываем форму редактирования
    document.getElementById('productDetails').style.display = 'block';
}

// Обработчик изменения категории в форме редактирования
document.getElementById('editCategory').addEventListener('change', function() {
    const category = this.value;
    const subcategoryGroup = document.getElementById('editSubcategoryGroup');
    const subcategorySelect = document.getElementById('editSubcategory');
    
    if (category === 'Фурнитура') {
        fetch('/api/subcategories')
            .then(response => response.json())
            .then(subcategories => {
                subcategorySelect.innerHTML = 
                    '<option value="">-- Выберите подкатегорию --</option>' +
                    subcategories.map(s => 
                        `<option value="${s.name}">${s.name}</option>`
                    ).join('');
                subcategoryGroup.style.display = 'block';
            })
            .catch(error => {
                console.error('Ошибка загрузки подкатегорий:', error);
                subcategoryGroup.style.display = 'none';
            });
    } else {
        subcategoryGroup.style.display = 'none';
        subcategorySelect.value = '';
    }
});

// Удаление товара
function deleteProduct() {
    const productId = document.getElementById('productId').value;
    if (!productId || !confirm('Вы уверены, что хотите удалить этот товар?')) return;
    
    fetch(`/api/products/delete?id=${productId}`, { method: 'DELETE' })
        .then(response => {
            if (response.ok) {
                alert('Товар успешно удален');
                document.getElementById('productDetails').style.display = 'none';
                document.getElementById('searchProduct').value = '';
            } else {
                alert('Ошибка при удалении товара');
            }
        })
        .catch(error => {
            console.error('Ошибка удаления:', error);
            alert('Ошибка при удалении товара');
        });
}

async function checkoutOrder() {
    if (this.disabled) return;
    
    if (!confirm('Вы уверены, что хотите оформить заказ?')) return;
    
    try {
        const response = await fetch('/api/orders/checkout', {
            method: 'POST',
            credentials: 'include',
            headers: {
                'Content-Type': 'application/json'
            }
        });

        if (!response.ok) {
            const errorData = await response.json().catch(() => null);
            throw new Error(errorData?.message || 'Ошибка оформления заказа');
        }

        const result = await response.json();
        if (result.success) {
            alert('Заказ успешно оформлен! Менеджер свяжется с вами для подтверждения.');
            showEmptyCart();
        } else {
            throw new Error(result.message || 'Неизвестная ошибка');
        }
        
    } catch (error) {
        console.error('Ошибка оформления:', error);
        alert(`Ошибка: ${error.message}`);
    }
}