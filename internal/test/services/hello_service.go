package services

import "go-api-template/internal/test/models"

// HelloService handles business logic for hello endpoints
type HelloService struct{}

// NewHelloService creates a new instance of the hello service
func NewHelloService() *HelloService {
	return &HelloService{}
}

// GetHelloWorld returns an example message
func (s *HelloService) GetHelloWorld() *models.HelloData {
	return &models.HelloData{
		Message: "Hello World! This is an example endpoint",
		Example: true,
	}
}
