package server

import (
	"fmt"
	"io"
	"net/http"

	"github.com/burntcarrot/wasmninja/internal/module"

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

func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to read request body: %v", err), http.StatusBadRequest)
		return
	}
	err = r.Body.Close()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to close request body: %v", err), http.StatusBadRequest)
		return
	}

	var reqBody InvokeRequest

	err = json.Unmarshal(body, &reqBody)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to parse request body: %v", err), http.StatusBadRequest)
		return
	}

	module, err := h.moduleLoader.LoadModule(reqBody.Module)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to load module: %v", err), http.StatusInternalServerError)
		return
	}

	output, err := h.moduleInvoker.InvokeModuleWaZero(module.Bytes, reqBody.Data)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to invoke module: %v", err), http.StatusInternalServerError)
		return
	}

	response := InvokeResponse{
		Result: string(output),
	}

	responseJSON, err := json.Marshal(response)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to marshal response: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(responseJSON)

}
