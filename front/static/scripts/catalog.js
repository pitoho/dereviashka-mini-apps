document.addEventListener('DOMContentLoaded', function() {
    const catalogTitle = document.querySelector('.catalog-container h1');
    const categoriesContainer = document.getElementById('categoriesContainer');
    const subcategoriesContainer = document.getElementById('subcategoriesContainer');
    const productsContainer = document.getElementById('productsContainer');
    const backToCategoriesBtn = document.getElementById('backToCategories');
    const backToSubcategoriesBtn = document.createElement('button');
    
    // Настройка кнопки "Назад к подкатегориям"
    backToSubcategoriesBtn.className = 'back-button';
    backToSubcategoriesBtn.textContent = '← Назад к подкатегориям';
    backToSubcategoriesBtn.style.display = 'none';
    productsContainer.before(backToSubcategoriesBtn);
    
    // Загрузка категорий при открытии страницы
    loadCatalogCategories();
    
    // Обработчики кнопок "Назад"
    backToCategoriesBtn.addEventListener('click', function() {
        catalogTitle.style.display = 'block';
        subcategoriesContainer.style.display = 'none';
        categoriesContainer.style.display = 'flex';
        categoriesContainer.style.flexDirection = 'column';
        productsContainer.innerHTML = '';
        backToSubcategoriesBtn.style.display = 'none';
        categoriesContainer.style.width = '250px';
    });
    
    backToSubcategoriesBtn.addEventListener('click', function() {
        catalogTitle.style.display = 'block';
        productsContainer.innerHTML = '';
        productsContainer.style.display = 'none';
        subcategoriesContainer.style.display = 'block';
        backToSubcategoriesBtn.style.display = 'none';
    });
    
    // Функция загрузки категорий
    function loadCatalogCategories() {
        fetch('/api/catalog/categories')
            .then(response => {
                if (!response.ok) {
                    throw new Error(`Ошибка загрузки: статус ${response.status}`);
                }
                return response.json();
            })
            .then(categories => {
                if (!categories || categories.length === 0) {
                    throw new Error('Категории не найдены');
                }
                
                categoriesContainer.innerHTML = '';
                
                categories.forEach(category => {
                    const categoryCard = document.createElement('div');
                    categoryCard.className = 'category-card';
                    
                    const categoryButton = document.createElement('button');
                    categoryButton.className = 'category-button';
                    categoryButton.textContent = category.name;
                    
                    categoryButton.addEventListener('click', function() {
                        catalogTitle.style.display = 'block';
                        if (category.has_subcategories) {
                            loadCatalogSubcategories(category.id);
                        } else {
                            loadCatalogProducts(category.id);
                        }
                    });
                    
                    categoryCard.appendChild(categoryButton);
                    categoriesContainer.appendChild(categoryCard);
                });
            })
            .catch(error => {
                console.error('Error loading categories:', error);
                categoriesContainer.innerHTML = `
                    <div class="error-container">
                        <p class="error-message">${error.message}</p>
                    </div>
                `;
            });
    }
    
    // Функция загрузки подкатегорий
    function loadCatalogSubcategories(categoryId) {
        fetch(`/api/catalog/subcategories?category_id=${categoryId}`)
            .then(response => {
                if (!response.ok) {
                    throw new Error(`Ошибка загрузки: статус ${response.status}`);
                }
                return response.json();
            })
            .then(subcategories => {
                if (!subcategories || subcategories.length === 0) {
                    throw new Error('Подкатегории не найдены');
                }
                
                catalogTitle.style.display = 'block';
                categoriesContainer.style.display = 'none';
                subcategoriesContainer.style.display = 'block';
                productsContainer.style.display = 'none';
                backToSubcategoriesBtn.style.display = 'none';
                
                const subcategoriesList = document.createElement('div');
                subcategoriesList.className = 'subcategories-list';
                
                subcategories.forEach(subcategory => {
                    const subcategoryButton = document.createElement('button');
                    subcategoryButton.className = 'subcategory-button';
                    subcategoryButton.textContent = subcategory.name;
                    
                    subcategoryButton.addEventListener('click', function() {
                        catalogTitle.style.display = 'block';
                        loadCatalogProducts(subcategory.id, true);
                    });
                    
                    subcategoriesList.appendChild(subcategoryButton);
                });
                
                const existingList = document.querySelector('.subcategories-list');
                if (existingList) {
                    subcategoriesContainer.removeChild(existingList);
                }
                subcategoriesContainer.appendChild(subcategoriesList);
            })
            .catch(error => {
                console.error('Error loading subcategories:', error);
                subcategoriesContainer.innerHTML = `
                    <div class="error-container">
                        <p class="error-message">${error.message}</p>
                        <button onclick="loadCatalogCategories()" class="back-button">← Назад к категориям</button>
                    </div>
                `;
            });
    }
    
    // Функция загрузки товаров
    function loadCatalogProducts(categoryOrSubcategoryId, isSubcategory = false) {
        const endpoint = isSubcategory ? 
            `/api/catalog/products?subcategory_id=${categoryOrSubcategoryId}` :
            `/api/catalog/products?category_id=${categoryOrSubcategoryId}`;
        
        fetch(endpoint)
            .then(response => {
                if (!response.ok) {
                    throw new Error(`Ошибка загрузки: статус ${response.status}`);
                }
                return response.json();
            })
            .then(products => {
                catalogTitle.style.display = 'block';
                subcategoriesContainer.style.display = 'none';
                productsContainer.style.display = 'block';
                productsContainer.innerHTML = '';
                
                if (!products || products.length === 0) {
                    productsContainer.innerHTML = `
                        <div class="empty-container">
                            <p class="empty-message">Товары не найдены</p>
                        </div>
                    `;
                    backToSubcategoriesBtn.style.display = isSubcategory ? 'block' : 'none';
                    return;
                }
                
                backToSubcategoriesBtn.style.display = isSubcategory ? 'block' : 'none';
                
                const productsGrid = document.createElement('div');
                productsGrid.className = 'products-grid';
                
                products.forEach(product => {
                    const productCard = document.createElement('div');
                    productCard.className = 'product-card clickable';
                    
                    const isInStock = product.in_stock === true || product.in_stock === 1 || product.in_stock === '1';
                    
                    if (!isInStock) {
                        productCard.classList.add('out-of-stock');
                    }
                    
                    const productImage = document.createElement('img');
                    productImage.src = product.image_url || '../static/images/no-image.png';
                    productImage.alt = product.name;
                    
                    const productInfo = document.createElement('div');
                    productInfo.className = 'product-info';
                    
                    const productName = document.createElement('h3');
                    productName.textContent = product.name;
                    
                    const productPrice = document.createElement('p');
                    productPrice.className = 'price';
                    productPrice.textContent = `${product.price.toLocaleString('ru-RU')} ₽`;
                    
                    const stockStatus = document.createElement('p');
                    stockStatus.className = isInStock ? 'in-stock' : 'out-of-stock';
                    stockStatus.textContent = isInStock ? 'В наличии' : 'Нет в наличии';
                    
                    productInfo.appendChild(productName);
                    productInfo.appendChild(productPrice);
                    productInfo.appendChild(stockStatus);
                    
                    productCard.appendChild(productImage);
                    productCard.appendChild(productInfo);
                    
                    productCard.addEventListener('click', function(e) {
                        if (!e.target.closest('.add-to-cart')) {
                            window.location.href = `/product?id=${product.id}`;
                        }
                    });
                    
                    productsGrid.appendChild(productCard);
                });
                
                productsContainer.appendChild(productsGrid);
            })
            .catch(error => {
                console.error('Error loading products:', error);
                productsContainer.innerHTML = `
                    <div class="error-container">
                        <p class="error-message">${error.message}</p>
                        <button onclick="${isSubcategory ? 'loadCatalogSubcategories(' + categoryOrSubcategoryId + ')' : 'loadCatalogCategories()'}" 
                                class="back-button">
                            ← ${isSubcategory ? 'Назад к подкатегориям' : 'Назад к категориям'}
                        </button>
                    </div>
                `;
            });
    }

});