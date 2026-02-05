package initialize

import (
	"log"

	"CONVERDA/internal/r2/controller/http"
	"CONVERDA/internal/r2/service"
)

// InitR2 initializes R2 service and handler
func InitR2() *http.R2Handler {
	// Initialize R2 service
	r2Service, err := service.NewR2Service()
	if err != nil {
		log.Printf("Warning: Failed to initialize R2 service: %v", err)
		return nil
	}

	// Initialize handler
	handler := http.NewR2Handler(r2Service)
	return handler
}
