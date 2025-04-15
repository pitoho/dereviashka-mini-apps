document.addEventListener('DOMContentLoaded', function() {
    const urlParams = new URLSearchParams(window.location.search);
    const categoryName = urlParams.get('name');
    
    if (categoryName) {
        document.getElementById('category-title').textContent = categoryName;
        loadCategoryProducts(categoryName);
    }
});

function loadCategoryProducts(categoryName) {
    fetch(`/api/category-products?name=${encodeURIComponent(categoryName)}`)
        .then(response => response.json())
        .then(products => {
            const container = document.getElementById('products-container');
            container.innerHTML = '';
            
            if (products.length === 0) {
                container.innerHTML = '<p>Товары не найдены</p>';
                return;
            }
            
            products.forEach(product => {
                const productCard = createProductCard(product);
                container.appendChild(productCard);
            });
        })
        .catch(error => {
            console.error('Ошибка загрузки товаров:', error);
            document.getElementById('products-container').innerHTML = 
                '<p>Произошла ошибка при загрузке товаров</p>';
        });
}

function createProductCard(product) {
    const card = document.createElement('div');
    card.className = 'product-card';
    
    card.innerHTML = `
        <div class="product-image">
            <img src="${product.image_path}" alt="${product.name}">
        </div>
        <div class="product-info">
            <h3>${product.name}</h3>
            <p class="price">${product.price.toFixed(2)} руб.</p>
            <p class="description">${product.description}</p>
            <button class="add-to-cart">В корзину</button>
        </div>
    `;
    
    return card;
}

