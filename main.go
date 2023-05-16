package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/wasmerio/wasmer-go/wasmer"
)

var (
	redisClient *redis.Client
	store       *wasmer.Store
	modules     sync.Map
)

func init() {
	// Create a Redis client.
	redisClient = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	// Initialize the wasmer runtime.
	err := initializeRuntime()
	if err != nil {
		log.Fatal("Failed to initialize Wasmer runtime:", err)
	}

	// Preload all WebAssembly modules.
	if err := preloadModules(); err != nil {
		log.Fatal("Failed to preload WebAssembly modules:", err)
	}
}

func main() {

	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	server := &http.Server{
		Addr:         ":8080",
		Handler:      http.HandlerFunc(handler),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-stop
		log.Println("Shutting down server...")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			log.Fatal("Server shutdown error:", err)
		}
	}()

	log.Println("Starting server on port 8080...")
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal("Server error:", err)
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	moduleName, query := parseRequestPath(r.URL.Path)

	module, err := getModule(moduleName)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to load module: %v", err), http.StatusInternalServerError)
		return
	}

	output, err := invokeModule(module, query)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to invoke module: %v", err), http.StatusInternalServerError)
		return
	}

	// Set the module's stdout as the response body
	w.Write(output)
}

func parseRequestPath(path string) (string, string) {
	// TODO: parse request path to get module name and query
	return "example", "42"
}

func preloadModules() error {
	moduleFiles, err := os.ReadDir("wasm_modules")
	if err != nil {
		return fmt.Errorf("failed to read module directory: %w", err)
	}

	for _, file := range moduleFiles {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".wasm" {
			modulePath := filepath.Join("wasm_modules", file.Name())

			wasmBytes, err := os.ReadFile(modulePath)
			if err != nil {
				return fmt.Errorf("failed to read module %s: %w", modulePath, err)
			}

			module, err := wasmer.NewModule(store, wasmBytes)
			if err != nil {
				return fmt.Errorf("failed to load module %s: %w", modulePath, err)
			}

			if err := cacheModule(file.Name(), module); err != nil {
				return fmt.Errorf("failed to cache module %s: %w", modulePath, err)
			}

			modules.Store(file.Name(), module)
		}
	}

	return nil
}

func initializeRuntime() error {
	engine := wasmer.NewEngine()
	store = wasmer.NewStore(engine)

	return nil
}

func getModule(moduleName string) (*wasmer.Module, error) {
	module, err := getModuleFromCache(moduleName)
	if err != nil {
		// Module not found in cache, load it from disk
		modulePath := filepath.Join("wasm_modules", moduleName+".wasm")
		wasmBytes, err := os.ReadFile(modulePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read module %s: %w", modulePath, err)
		}
		module, err = wasmer.NewModule(store, wasmBytes)
		if err != nil {
			return nil, fmt.Errorf("failed to load module %s: %w", modulePath, err)
		}

		if err := cacheModule(moduleName, module); err != nil {
			return nil, fmt.Errorf("failed to cache module %s: %w", modulePath, err)
		}
	}

	return module, nil
}

func getModuleFromCache(moduleName string) (*wasmer.Module, error) {
	moduleBytes, err := redisClient.Get(context.Background(), moduleName).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("module not found in cache")
		}
		return nil, fmt.Errorf("failed to get module from cache: %w", err)
	}
	module, err := wasmer.NewModule(store, moduleBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to load module from cache: %w", err)
	}

	return module, nil
}

func cacheModule(moduleName string, module *wasmer.Module) error {
	moduleBytes, err := module.Serialize()
	if err != nil {
		return fmt.Errorf("failed to serialize module: %w", err)
	}
	err = redisClient.Set(context.Background(), moduleName, moduleBytes, 0).Err()
	if err != nil {
		return fmt.Errorf("failed to cache module: %w", err)
	}

	return nil
}

func invokeModule(module *wasmer.Module, query string) ([]byte, error) {
	// Create a new WebAssembly instance from the module
	wasiEnv, _ := wasmer.NewWasiStateBuilder("wasi-program").Finalize()
	importObject, err := wasiEnv.GenerateImportObject(store, module)
	if err != nil {
		return nil, fmt.Errorf("failed to generate import object: %w", err)
	}

	instance, err := wasmer.NewInstance(module, importObject)
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate module: %w", err)
	}

	start, err := instance.Exports.GetWasiStartFunction()
	if err != nil {
		return nil, fmt.Errorf("failed to run start function: %w", err)
	}

	start()

	// Call the module's exported function with the input
	HelloWorld, err := instance.Exports.GetFunction("HelloWorld")
	if err != nil {
		return nil, fmt.Errorf("module invocation failed: %w", err)
	}

	// Invoke the module's function with the query value
	result, err := HelloWorld()
	if err != nil {
		return nil, fmt.Errorf("module invocation failed with query value %s: %w", query, err)
	}

	return []byte(fmt.Sprintf("%v", result)), nil
}
