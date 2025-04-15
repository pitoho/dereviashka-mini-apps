const menuToggle = document.querySelector('.menu-toggle');
const menu = document.querySelector('.menu');

menuToggle.addEventListener('click', () => {
    menuToggle.classList.toggle('active');
    menu.classList.toggle('active');
});


const slides = document.querySelector('.slides');
        const prevBtn = document.querySelector('.prev');
        const nextBtn = document.querySelector('.next');
        const indicators = document.querySelectorAll('.indicator');
        const slideElements = document.querySelectorAll('.slide');
        let currentSlide = 0;
        const slideCount = slideElements.length;
        
        // Клонируем первый и последний слайды для бесконечности
const firstClone = slideElements[0].cloneNode(true);
const lastClone = slideElements[slideCount - 1].cloneNode(true);
        
slides.appendChild(firstClone);
slides.insertBefore(lastClone, slideElements[0]);
        
        // Корректируем текущий слайд после добавления клонов
currentSlide = 1;
slides.style.transform = `translateX(-${currentSlide * 100 / slideCount}%)`;
        
function goToSlide(slideIndex) {
        slides.style.transition = 'transform 0.5s ease-in-out';
        currentSlide = slideIndex;
        slides.style.transform = `translateX(-${currentSlide * 100 / slideCount}%)`;
            
        updateIndicators();
    }
        
function updateIndicators() {
    let activeIndex;
    if (currentSlide === 0) {
        activeIndex = slideCount - 1;
    } else if (currentSlide === slideCount + 1) {
        activeIndex = 0;
    } else {
        activeIndex = currentSlide - 1;
    }
            
    indicators.forEach((ind, index) => {
        ind.classList.toggle('active', index === activeIndex);
    });
}
        
function nextSlide() {
    if (currentSlide >= slideCount + 1) return;
    currentSlide++;
    slides.style.transform = `translateX(-${currentSlide * 100 / slideCount}%)`;
            
    if (currentSlide === slideCount + 1) {
        setTimeout(() => {
            slides.style.transition = 'none';
            currentSlide = 1;
            slides.style.transform = `translateX(-${currentSlide * 100 / slideCount}%)`;
            setTimeout(() => {
                slides.style.transition = 'transform 0.5s ease-in-out';
            }, 10);
        }, 500);
    }
            
    updateIndicators();
}
        
function prevSlide() {
    if (currentSlide <= 0) return;
    currentSlide--;
    slides.style.transform = `translateX(-${currentSlide * 100 / slideCount}%)`;
            
    if (currentSlide === 0) {
        setTimeout(() => {
            slides.style.transition = 'none';
            currentSlide = slideCount;
            slides.style.transform = `translateX(-${currentSlide * 100 / slideCount}%)`;
            setTimeout(() => {
                slides.style.transition = 'transform 0.5s ease-in-out';
            }, 10);
        }, 500);
    }
            
    updateIndicators();
}
        
prevBtn.addEventListener('click', prevSlide);
nextBtn.addEventListener('click', nextSlide);
        
indicators.forEach((indicator, index) => {
    indicator.addEventListener('click', () => {
        goToSlide(index + 1);
    });
});
        
        // Автопрокрутка с бесконечностью
let interval = setInterval(nextSlide, 5000);
        
        // Пауза при наведении
carousel.addEventListener('mouseenter', () => {
    clearInterval(interval);
});
        
carousel.addEventListener('mouseleave', () => {
    interval = setInterval(nextSlide, 5000);
});

