// Command server runs the user_auth HTTP service.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Lanskask/user_auth/internal/auth"
	"github.com/Lanskask/user_auth/internal/config"
	"github.com/Lanskask/user_auth/internal/httpapi"
	"github.com/Lanskask/user_auth/internal/store"
	"github.com/Lanskask/user_auth/web"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	if err := run(log); err != nil {
		log.Error("server stopped", "err", err)
		os.Exit(1)
	}
}

func run(log *slog.Logger) error {
	cfg, err := config.Load(os.Getenv)
	if err != nil {
		return fmt.Errorf("config: %w", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	db, err := store.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("store: %w", err)
	}
	defer db.Close()

	svc, err := auth.NewService(db, cfg.BcryptCost, cfg.SessionTTL)
	if err != nil {
		return err
	}

	static, err := web.Static()
	if err != nil {
		return fmt.Errorf("static files: %w", err)
	}

	go purgeExpiredSessions(ctx, log, db)

	srv := &http.Server{
		Addr: cfg.Addr,
		Handler: httpapi.New(httpapi.Options{
			Auth: svc, DB: db, Logger: log, CookieSecure: cfg.CookieSecure, RateBurst: cfg.RateBurst, Static: static,
		}),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		log.Info("listening", "addr", cfg.Addr, "cookie_secure", cfg.CookieSecure)
		errCh <- srv.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		return fmt.Errorf("listen: %w", err)
	case <-ctx.Done():
		log.Info("shutting down")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("shutdown: %w", err)
	}
	return nil
}

// purgeExpiredSessions deletes dead sessions hourly until ctx is cancelled.
func purgeExpiredSessions(ctx context.Context, log *slog.Logger, db *store.Store) {
	t := time.NewTicker(time.Hour)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			n, err := db.DeleteExpiredSessions(ctx, time.Now())
			if err != nil {
				log.Warn("purge sessions", "err", err)
				continue
			}
			log.Info("purged expired sessions", "count", n)
		}
	}
}
