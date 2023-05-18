package server

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/burntcarrot/wasmninja/internal/cache"
	"github.com/burntcarrot/wasmninja/internal/config"
	"github.com/burntcarrot/wasmninja/internal/module"
	"github.com/burntcarrot/wasmninja/internal/runtime"
	"github.com/valyala/fasthttp"
)

type App struct {
	server *fasthttp.Server
	cfg    *config.Config
}

func NewApp() (*App, error) {
	configFile := flag.String("config", "config.yaml", "path to the config file")
	flag.Parse()

	cfg, err := config.NewConfig(*configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %v", err)
	}

	cache, err := cache.NewConnection(cfg.Cache)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to cache: %v", err)
	}

	runtime := runtime.NewRuntime(context.Background())

	moduleCache := module.NewModuleCache(cache.Connection)

	moduleLoader, err := module.NewModuleLoader(cfg.Loader, moduleCache, runtime)
	if err != nil {
		return nil, fmt.Errorf("failed to init module loader: %v", err)
	}

	moduleInvoker := module.NewModuleInvoker(runtime)

	if err := module.Preload(moduleLoader); err != nil {
		return nil, fmt.Errorf("failed to preload WebAssembly module: %v", err)
	}

	server := setupServer(cfg.Server, moduleLoader, moduleInvoker)

	return &App{
		server: server,
		cfg:    cfg,
	}, nil
}

func (a *App) Start() error {
	// Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-stop
		log.Println("Shutting down server...")
		a.shutdownServer()
	}()

	addr := fmt.Sprintf("%s:%d", a.cfg.Server.Host, a.cfg.Server.Port)
	if err := a.server.ListenAndServe(addr); err != nil && err != http.ErrServerClosed {
		return err
	}

	return nil
}

func setupServer(cfg config.ServerConfig, moduleLoader *module.ModuleLoader, moduleInvoker *module.ModuleInvoker) *fasthttp.Server {
	handler := NewHandler(moduleLoader, moduleInvoker)
	mux := requestHandler(handler)

	log.Printf("Starting server on %s....\n", fmt.Sprintf("%s:%d", cfg.Host, cfg.Port))

	return &fasthttp.Server{
		Handler:            mux,
		MaxConnsPerIP:      10000,
		ReadTimeout:        30 * time.Second,
		WriteTimeout:       30 * time.Second,
		IdleTimeout:        60 * time.Second,
		MaxRequestBodySize: 10 * 1024 * 1024, // 10MB
		DisableKeepalive:   true,
	}
}

func requestHandler(h *Handler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		switch string(ctx.Path()) {
		case "/invoke":
			h.Handle(ctx)
		case "/health":
			h.Health(ctx)
		default:
			ctx.Error("Not Found", fasthttp.StatusNotFound)
		}
	}
}

func (a *App) shutdownServer() {
	if err := a.server.Shutdown(); err != nil {
		log.Println("Server shutdown error:", err)
		return
	}
}
