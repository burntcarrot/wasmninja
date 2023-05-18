package module

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/burntcarrot/wasmninja/internal/config"
	"github.com/burntcarrot/wasmninja/internal/objectstore"
	"github.com/burntcarrot/wasmninja/internal/runtime"

	"github.com/minio/minio-go/v7"
)

const (
	OBJECTSTORE_LOADER = "objectstore"
	FILESYSTEM_LOADER  = "fs"
)

type ModuleLoader struct {
	cache        *ModuleCache
	runtime      *runtime.Runtime
	config       config.LoaderConfig
	objectClient *minio.Client
}

func NewModuleLoader(cfg config.LoaderConfig, cache *ModuleCache, runtime *runtime.Runtime) (*ModuleLoader, error) {
	loader := &ModuleLoader{
		cache:   cache,
		runtime: runtime,
		config:  cfg,
	}

	if cfg.ModuleLoader == OBJECTSTORE_LOADER {
		client, err := objectstore.ConnectMinio(cfg.MinioConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create Minio client: %w", err)
		}
		loader.objectClient = client
	}

	return loader, nil
}

func (m *ModuleLoader) LoadModule(moduleName string) (*Module, error) {
	module, err := m.cache.GetModule(moduleName)
	if err != nil {
		var moduleObject []byte
		var objectPath string

		switch m.config.ModuleLoader {
		case FILESYSTEM_LOADER:
			modulePath := filepath.Join(m.config.ModuleDirectory, moduleName+".wasm")
			moduleObject, err = os.ReadFile(modulePath)
			objectPath = modulePath

		case OBJECTSTORE_LOADER:
			objectPath = moduleName + ".wasm"
			moduleObject, err = m.loadModuleFromObjectStore(objectPath)
		}

		if err != nil {
			return nil, fmt.Errorf("failed to load module %s: %w", objectPath, err)
		}

		if err := m.cache.CacheModule(moduleName, moduleObject); err != nil {
			return nil, fmt.Errorf("failed to cache module %s: %w", objectPath, err)
		}

		module = moduleObject
	}

	return &Module{
		Name:  moduleName,
		Bytes: module,
	}, nil
}

func (m *ModuleLoader) loadModuleFromObjectStore(objectPath string) ([]byte, error) {
	object, err := m.objectClient.GetObject(context.Background(), m.config.MinioConfig.BucketName, objectPath, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	defer object.Close()

	objectData, err := io.ReadAll(object)
	if err != nil {
		return nil, fmt.Errorf("failed to read object data: %v", err)
	}

	return objectData, nil
}
