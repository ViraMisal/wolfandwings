// Волчица и Крылья — API.
// Copyright (C) 2026 Грёзов Саярин Аквилович (https://github.com/ViraMisal)
//
// SPDX-License-Identifier: AGPL-3.0-only
//
// Эта программа — свободное ПО: её можно распространять и/или изменять
// на условиях GNU Affero General Public License v3.0 (исключительно этой
// версии), опубликованной Free Software Foundation;
// см. https://www.gnu.org/licenses/agpl-3.0.html и файл LICENSE в корне.
//
// Программа распространяется в надежде, что будет полезной, но БЕЗ КАКИХ-ЛИБО
// ГАРАНТИЙ — без подразумеваемых гарантий товарности и пригодности для
// конкретных целей. Подробности: https://www.gnu.org/licenses/agpl-3.0.html

package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	"github.com/wolfandwings/api/internal/admin"
	"github.com/wolfandwings/api/internal/app"
	"github.com/wolfandwings/api/internal/catalog"
	"github.com/wolfandwings/api/internal/config"
	"github.com/wolfandwings/api/internal/notify"
	"github.com/wolfandwings/api/internal/order"
	"github.com/wolfandwings/api/internal/platform"
	"github.com/wolfandwings/api/internal/preorder"
	"github.com/wolfandwings/api/internal/store"
)

func main() {
	_ = godotenv.Load(".env", "../.env")

	cfg, err := config.Load()
	if err != nil {
		slog.Error("config", "err", err)
		os.Exit(1)
	}

	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: cfg.LogLevel}))
	slog.SetDefault(log)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	db, err := platform.Open(ctx, cfg, log)
	if err != nil {
		log.Error("db open", "err", err)
		os.Exit(1)
	}
	defer db.Close()

	st := store.New(db, cfg.Dialect)
	if err := st.SeedIfEmpty(ctx, cfg.AdminEmail, cfg.AdminPassword, log); err != nil {
		log.Error("seed", "err", err)
		os.Exit(1)
	}

	tg := &notify.Telegram{BotToken: cfg.TGBotToken, ChatID: cfg.TGNotifyChat, Log: log}

	deps := app.Deps{
		Cfg:      cfg,
		Store:    st,
		Catalog:  &catalog.Handlers{Store: st},
		Preorder: &preorder.Handlers{Store: st, Notify: tg},
		Order:    &order.Handlers{Store: st},
		Admin:    admin.NewHandlers(st, cfg),
	}
	handler := app.New(deps, log)

	srv := &http.Server{
		Addr:         cfg.APIAddr,
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Info("api listening", "addr", cfg.APIAddr, "dialect", cfg.Dialect, "env", cfg.Env)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("listen", "err", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	log.Info("shutting down")
	shutdownCtx, c := context.WithTimeout(context.Background(), 10*time.Second)
	defer c()
	_ = srv.Shutdown(shutdownCtx)
}
