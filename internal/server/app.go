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
)

type App struct {
	server *http.Server
}

func NewApp() (*App, error) {
	configFile := flag.String("config", "config.yaml", "path to the config file")
	flag.Parse()

	cfg, err := config.NewConfig(*configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %v", err)
	}

	cache := cache.NewConnection(cfg.Cache)
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

	if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}

	return nil
}

func setupServer(cfg config.ServerConfig, moduleLoader *module.ModuleLoader, moduleInvoker *module.ModuleInvoker) *http.Server {
	mux := http.NewServeMux()

	handler := NewHandler(moduleLoader, moduleInvoker)
	mux.HandleFunc("/invoke", handler.Handle)
	mux.HandleFunc("/health", handler.Health)

	log.Printf("Starting server on %s....\n", fmt.Sprintf("%s:%d", cfg.Host, cfg.Port))

	return &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
}

func (a *App) shutdownServer() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := a.server.Shutdown(ctx); err != nil {
		log.Println("Server shutdown error:", err)
		return
	}
}
