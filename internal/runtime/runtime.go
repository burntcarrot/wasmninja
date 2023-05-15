package runtime

import (
	"context"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

type Runtime struct {
	wazero.Runtime
}

func NewRuntime(ctx context.Context) *Runtime {
	runtime := wazero.NewRuntime(ctx)
	wasi_snapshot_preview1.MustInstantiate(ctx, runtime)

	return &Runtime{
		Runtime: runtime,
	}
}
