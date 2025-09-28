package main

import (
	"context"
	"net/http"
	"errors"
	"time"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"os/signal"
	"syscall"

	"github.com/kujoki/loyalty-service/internal/config"
	"github.com/kujoki/loyalty-service/internal/handler"
	"github.com/kujoki/loyalty-service/internal/repository"
)

func main() {
	cfg := config.ParseConfig()
    var sugar zap.SugaredLogger

    if err := run(cfg, sugar); err != nil {
		sugar.Fatalw(err.Error(), "event", "start server")
    }
}

func run(cfg *config.Config, sugar zap.SugaredLogger) error {
	logger, err := zap.NewDevelopment()
	if err != nil {
        return errors.New("failed to create logger")
    }
    defer logger.Sync()
	sugar = *logger.Sugar()

	r := chi.NewRouter()

	r.Use(handler.GzipMiddlewareRequest)
	r.Use(handler.GzipMiddlewareResponse)
	// r.Use() авторизация

	repo, err := repository.NewPostgresRepository(cfg.DatabaseURL)

	if err != nil {
		return errors.New("failed to connect")
	}

	r.Post("/api/user/register", handler.WrapperHandlerAuth(repo)) // регистрация пользователя
	r.Post("/api/user/login", handler.WrapperHandlerLogin(repo)) // аутентификация пользователя
	r.Post("/api/user/orders", func(w http.ResponseWriter, r *http.Request) {}) // загрузка пользователем номера заказа для расчёта
	r.Get("/api/user/orders", func(w http.ResponseWriter, r *http.Request) {}) // получение списка загруженных пользователем номеров заказов, статусов их обработки и информации о начислениях
	r.Get("/api/user/balance", func(w http.ResponseWriter, r *http.Request) {}) // получение текущего баланса счёта баллов лояльности пользователя
	r.Post("/api/user/balance/withdraw", func(w http.ResponseWriter, r *http.Request) {}) // запрос на списание баллов с накопительного счёта в счёт оплаты нового заказа
	r.Get("/api/user/withdrawals", func(w http.ResponseWriter, r *http.Request) {}) // получение информации о выводе средств с накопительного счёта пользователем

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`this request are not allowed!`))
	})

	srv := &http.Server{
		Addr:    cfg.RunAddr,
		Handler: r,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		sugar.Infow("running server", "address", cfg.RunAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			sugar.Fatalw(err.Error(), "event", "start server")
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// repo.Close()
	sugar.Infow("close repository")

    sugar.Infow("shutting down server gracefully")
	return srv.Shutdown(shutdownCtx)
}