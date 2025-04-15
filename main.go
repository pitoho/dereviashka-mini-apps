package main

import (

	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"syscall"
	"os/signal"
	"sync"
	"time"
	"strconv"


	"main/handler"
	"main/models"
	"main/storage"

	"github.com/joho/godotenv"
	tgbot "main/bot"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var (
	tokens = make(map[string]models.UserInfo) // token -> expiration time
	pendingLogins  = make(map[int64]string)    // chatID -> token
	tokensMutex    sync.Mutex
	user models.UserInfo
	cleanupInterval = 1 * time.Hour
	managerGroupID int64

)

func main() {
	// Загружаем переменные окружения
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	managerGroupID, err = strconv.ParseInt(os.Getenv("TELEGRAM_MANAGER_GROUP_ID"), 10, 64)
	if err != nil {
		log.Fatal("Invalid TELEGRAM_MANAGER_GROUP_ID in .env file")
	}

	// Инициализация базы данных
	dbConn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"))

	err = tgbot.InitDB(dbConn)
	if err != nil {
		log.Fatalf("Ошибка инициализации БД для бота: %v", err)
	}

	err = storage.InitDB(dbConn)
	if err != nil {
		log.Fatalf("Ошибка инициализации БД для хендлера: %v", err)
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	checkDirectory(filepath.Join("front", "static"))
	checkDirectory(filepath.Join("front", "pages"))

	// Запускаем очистку устаревших токенов
	go tgbot.TokenCleanup()

	// HTTP сервер
	go func() {
		staticHandler := http.StripPrefix("/static/", 
			http.FileServer(http.Dir(filepath.Join("front", "static"))))
		http.Handle("/static/", staticHandler)

		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/" {
				http.NotFound(w, r)
				return
			}
			servePage(w, r, "index.html")
		})
		http.HandleFunc("/services", func(w http.ResponseWriter, r *http.Request) {
			servePage(w, r, "services.html")
		})
		http.HandleFunc("/delivery", func(w http.ResponseWriter, r *http.Request) {
			servePage(w, r, "delivery.html")
		})
		http.HandleFunc("/contacts", func(w http.ResponseWriter, r *http.Request) {
			servePage(w, r, "contacts.html")
		})
		// Обработчик для /profile
		http.HandleFunc("/profile", func(w http.ResponseWriter, r *http.Request) {
			handler.ProfileHandler(w,r,user)
		})

		http.HandleFunc("/auth", func(w http.ResponseWriter, r *http.Request) {
			handler.AuthoriseHeandler(w,r)
		})
		http.HandleFunc("/add-product", func(w http.ResponseWriter, r *http.Request) {
			handler.AddItemToCatalog(w,r)
		})
		http.HandleFunc("/catalog", func(w http.ResponseWriter, r *http.Request) {
			servePage(w, r, "catalog.html")
		})

		// Обработчик для получения категорий
		http.HandleFunc("/api/categories", func(w http.ResponseWriter, r *http.Request) {
		    handler.GetCategory(w,r)
		})

		// Обработчик для получения подкатегорий
		http.HandleFunc("/api/subcategories", func(w http.ResponseWriter, r *http.Request) {
    		handler.GetSubcategory(w,r)
		})

		// Add this with your other routes
		http.HandleFunc("/api/products/search", handler.SearchProduct)

		http.HandleFunc("/api/catalog/categories", handler.GetCatalogCategories)
		http.HandleFunc("/api/catalog/subcategories", handler.GetCatalogSubcategories)
		http.HandleFunc("/api/catalog/products", handler.GetCatalogProducts)

		http.HandleFunc("/api/catalog/product", handler.GetProductDetails)
		http.HandleFunc("/product", func(w http.ResponseWriter, r *http.Request) {
 		   servePage(w, r, "product.html")
		})

		http.HandleFunc("/api/cart/add", handler.AddToCart)
		http.HandleFunc("/api/orders", handler.GetUserOrdersHandler)
		http.HandleFunc("/api/orders/", handler.DeleteOrderHandler)


		http.HandleFunc("/api/orders/checkout", handler.CheckoutOrder)
		port := ":8080"
		fmt.Printf("HTTP server running on port %s\n", port)
		if err := http.ListenAndServe(port, nil); err != nil {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	// Telegram бот
	go func() {
		var err error
		tgbot.Bot, err = tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_BOT_TOKEN"))
		if err != nil {
			log.Printf("Ошибка создания бота: %v", err)
			return
		}

		tgbot.Bot.Debug = true
		log.Printf("Бот авторизован как %s", tgbot.Bot.Self.UserName)

		u := tgbotapi.NewUpdate(0)
		u.Timeout = 60

		updates := tgbot.Bot.GetUpdatesChan(u)

		go tgbot.TokenCleanup()

		for update := range updates {
			if update.Message != nil {
				tgbot.HandleMessage(update.Message)
			}
		}
	}()

	<-done
	fmt.Println("\nПриложение завершает работу...")
}


func checkDirectory(path string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		log.Fatalf("Папка %s не найдена: %v", path, err)
	}
}

func redirectHandler(target string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, target, http.StatusMovedPermanently)
	}
}

func servePage(w http.ResponseWriter, r *http.Request, page string) {
	path := filepath.Join("front", "pages", page)
	
	if _, err := os.Stat(path); os.IsNotExist(err) {
		http.NotFound(w, r)
		log.Printf("Файл не найден: %s", path)
		return
	}
	
	http.ServeFile(w, r, path)
}


