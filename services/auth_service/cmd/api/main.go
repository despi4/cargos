package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/despi4/cargos/auth_service/internal/adapters/secondary/migrator"
	"github.com/despi4/cargos/auth_service/internal/adapters/secondary/postgres"
	pgRepo "github.com/despi4/cargos/auth_service/internal/adapters/secondary/postgres/repository"
	redisRepo "github.com/despi4/cargos/auth_service/internal/adapters/secondary/redis/repository"
	"github.com/despi4/cargos/auth_service/internal/core/service"
	"github.com/despi4/cargos/auth_service/migrations"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

func main() {
	// 1. Загрузка переменных окружения из .env (если файл существует)
	if err := godotenv.Load(); err != nil {
		// Попробуем загрузить из папки auth_service, если запущено из корня
		_ = godotenv.Load("services/auth_service/.env")
		_ = godotenv.Load("auth_service/.env")
	}

	dsn := os.Getenv("DSN")
	redisAddr := os.Getenv("REDIS_ADDR")
	redisPassword := os.Getenv("REDIS_PASSWORD")
	jwtSecret := os.Getenv("JWT_SECRET")
	port := os.Getenv("HTTP_PORT")

	// 2. Накатываем миграции базы данных перед стартом
	log.Println("Применение миграций базы данных...")
	if err := migrator.RunMigrations(migrations.FS, ".", dsn); err != nil {
		log.Fatalf("Ошибка миграций: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 3. Подключение к PostgreSQL
	db, err := postgres.NewDB(ctx, dsn)
	if err != nil {
		log.Fatalf("Ошибка подключения к PostgreSQL: %v", err)
	}
	defer db.Close()

	// 4. Подключение к Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       0,
	})
	defer func() {
		if err := rdb.Close(); err != nil {
			log.Printf("Ошибка при закрытии Redis: %v", err)
		}
	}()

	pingCtx, pingCancel := context.WithTimeout(ctx, 3*time.Second)
	defer pingCancel()
	if err := rdb.Ping(pingCtx).Err(); err != nil {
		log.Fatalf("Ошибка подключения к Redis: %v", err)
	}
	log.Println("Redis успешно подключен")

	// 5. Инициализация репозиториев (Secondary Adapters)
	userRepository := pgRepo.NewUserRepo(db.Pool)
	sessionRepository := redisRepo.NewRedisSessionRepo(rdb)

	// 6. Инициализация сервисов (Core Domain Services)
	hasher := service.NewPasswordHasher()
	tokenSvc := service.NewJwtService([]byte(jwtSecret), 15*time.Minute)
	authSvc := service.NewAuthSvc(userRepository, hasher, tokenSvc, sessionRepository)

	_ = authSvc // Готов для передачи в HTTP хендлеры

	// 7. Настройка базового HTTP-сервера с проверкой здоровья (Healthcheck)
	router := gin.Default()

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"messaage": "OK",
		})
	})

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 8. Запуск сервера в отдельной горутине
	go func() {
		log.Printf("Auth сервис запущен на порту :%s", port)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Ошибка HTTP сервера: %v", err)
		}
	}()

	// 9. Ожидание сигналов ОС для Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Завершение работы Auth сервиса...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Принудительная остановка сервера: %v", err)
	}

	log.Println("Сервис успешно остановлен")
}
