package handlers

import (
	"net/http"

	"go-api-template/internal/test/services"
	"go-api-template/pkg/response"
)

// HelloHandler handles HTTP requests for hello endpoints
type HelloHandler struct {
	service *services.HelloService
}

// NewHelloHandler creates a new instance of the handler
func NewHelloHandler(service *services.HelloService) *HelloHandler {
	return &HelloHandler{
		service: service,
	}
}

// GetHelloWorld godoc
// @Summary      Hello World Example
// @Description  Returns a hello world message
// @Tags         test
// @Produce      json
// @Success      200  {object}  models.HelloResponse  "Success response with hello data"
// @Failure      500  {object}  models.ErrorResponse  "Internal server error"
// @Router       /test/hello [get]
func (h *HelloHandler) GetHelloWorld(w http.ResponseWriter, _ *http.Request) {
	data := h.service.GetHelloWorld()
	response.Success(w, data)
}
