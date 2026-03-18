package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func Run(ctx context.Context, logger *slog.Logger) error {
	cfg, err := LoadConfig()
	if err != nil {
		return err
	}

	server := NewMCPServer(cfg)
	handler := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server {
		return server
	}, &mcp.StreamableHTTPOptions{
		Stateless:    cfg.Stateless,
		JSONResponse: cfg.JSONResponse,
		Logger:       logger,
	})

	mux := http.NewServeMux()
	mux.Handle("GET /healthz", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintf(w, `{"ok":true,"root":%q,"path":%q}`, cfg.Root, cfg.HTTPPath)
	}))
	mux.Handle(cfg.HTTPPath, WithBearerAuth(handler, cfg.Token))

	httpServer := &http.Server{
		Addr:    cfg.Addr,
		Handler: mux,
	}

	logger.Info("sandbox mcp server starting",
		slog.String("addr", cfg.Addr),
		slog.String("mcp_path", cfg.HTTPPath),
		slog.String("sandbox_root", cfg.Root),
		slog.Bool("auth_enabled", cfg.Token != ""),
		slog.Bool("stateless", cfg.Stateless),
		slog.Bool("json_response", cfg.JSONResponse),
	)

	errCh := make(chan error, 1)
	go func() {
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
			return
		}
		errCh <- nil
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), defaultToolTimeout)
		defer cancel()
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			return err
		}
		if err := <-errCh; err != nil {
			return err
		}
		return ctx.Err()
	case err := <-errCh:
		return err
	}
}
