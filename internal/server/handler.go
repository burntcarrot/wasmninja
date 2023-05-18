package server

import (
	"fmt"

	"github.com/burntcarrot/wasmninja/internal/module"
	"github.com/valyala/fasthttp"

	jsoniter "github.com/json-iterator/go"
)

var (
	json = jsoniter.ConfigFastest
)

type Handler struct {
	moduleLoader  *module.ModuleLoader
	moduleInvoker *module.ModuleInvoker
}

func NewHandler(moduleLoader *module.ModuleLoader, moduleInvoker *module.ModuleInvoker) *Handler {
	return &Handler{
		moduleLoader:  moduleLoader,
		moduleInvoker: moduleInvoker,
	}
}

func (h *Handler) Handle(ctx *fasthttp.RequestCtx) {
	reqBody := InvokeRequest{}
	if err := json.Unmarshal(ctx.Request.Body(), &reqBody); err != nil {
		ctx.Error(fmt.Sprintf("Failed to parse request body: %v", err), fasthttp.StatusBadRequest)
		return
	}

	module, err := h.moduleLoader.LoadModule(reqBody.Module)
	if err != nil {
		ctx.Error(fmt.Sprintf("Failed to load module: %v", err), fasthttp.StatusInternalServerError)
		return
	}

	output, err := h.moduleInvoker.InvokeModuleWaZero(module.Bytes, reqBody.Data)
	if err != nil {
		ctx.Error(fmt.Sprintf("Failed to invoke module: %v", err), fasthttp.StatusInternalServerError)
		return
	}

	response := InvokeResponse{
		Result: string(output),
	}

	responseJSON, err := json.Marshal(response)
	if err != nil {
		ctx.Error(fmt.Sprintf("Failed to marshal response: %v", err), fasthttp.StatusInternalServerError)
		return
	}

	ctx.SetContentType("application/json")
	_, err = ctx.Write(responseJSON)
	if err != nil {
		ctx.Error(fmt.Sprintf("Failed to write response: %v", err), fasthttp.StatusInternalServerError)
		return
	}
}

func (h *Handler) Health(ctx *fasthttp.RequestCtx) {
	ctx.SetStatusCode(fasthttp.StatusOK)
	ctx.SetBody([]byte("OK"))
}
