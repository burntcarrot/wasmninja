package module

import (
	"bytes"
	"context"

	"github.com/burntcarrot/wasmninja/internal/runtime"

	"github.com/tetratelabs/wazero"
)

type ModuleInvoker struct {
	runtime *runtime.Runtime
}

func NewModuleInvoker(runtime *runtime.Runtime) *ModuleInvoker {
	return &ModuleInvoker{
		runtime: runtime,
	}
}

func (m *ModuleInvoker) InvokeModuleWaZero(module []byte, data string) ([]byte, error) {
	// Create a new WebAssembly instance from the module
	var stdoutBuf bytes.Buffer
	config := wazero.NewModuleConfig().WithStdout(&stdoutBuf)

	env := map[string]string{
		"WASMNINJA_DATA": data,
	}

	for k, v := range env {
		config = config.WithEnv(k, v)
	}

	ctx := context.Background()

	// Instantiate the module. This invokes the _start function by default.
	_, err := m.runtime.InstantiateWithConfig(ctx, module, config)
	if err != nil {
		return []byte{}, err
	}

	return stdoutBuf.Bytes(), nil
}
