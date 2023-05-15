package server

type InvokeRequest struct {
	Module string `json:"module"`
	Data   string `json:"data"`
}
