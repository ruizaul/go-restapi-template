package httpx

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

// PaginationParams represents pagination parameters
type PaginationParams struct {
	Page   int
	Limit  int
	Offset int
}

// PaginationMetadata contains pagination information
type PaginationMetadata struct {
	CurrentPage int    `json:"current_page" example:"1"`
	PerPage     int    `json:"per_page" example:"20"`
	TotalItems  int    `json:"total_items" example:"847"`
	TotalPages  int    `json:"total_pages" example:"43"`
	HasNext     bool   `json:"has_next" example:"true"`
	HasPrevious bool   `json:"has_previous" example:"false"`
	NextURL     string `json:"next_url,omitempty" example:"/api/v1/orders?page=2&limit=20"`
	PreviousURL string `json:"previous_url,omitempty"`
}

// ParsePaginationParams parses pagination parameters from request
func ParsePaginationParams(r *http.Request) (*PaginationParams, error) {
	// Default values
	page := 1
	limit := 20

	// Parse page
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		p, err := strconv.Atoi(pageStr)
		if err != nil || p < 1 {
			return nil, fmt.Errorf("page debe ser un número entero mayor a 0")
		}
		page = p
	}

	// Parse limit
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		l, err := strconv.Atoi(limitStr)
		if err != nil || l < 1 {
			return nil, fmt.Errorf("limit debe ser un número entero mayor a 0")
		}
		if l > 100 {
			return nil, fmt.Errorf("limit no puede ser mayor a 100")
		}
		limit = l
	}

	// Calculate offset
	offset := (page - 1) * limit

	return &PaginationParams{
		Page:   page,
		Limit:  limit,
		Offset: offset,
	}, nil
}

// BuildPaginationMetadata builds pagination metadata for responses
func BuildPaginationMetadata(page, limit, total int, basePath string) PaginationMetadata {
	totalPages := (total + limit - 1) / limit
	if totalPages < 1 {
		totalPages = 1
	}

	meta := PaginationMetadata{
		CurrentPage: page,
		PerPage:     limit,
		TotalItems:  total,
		TotalPages:  totalPages,
		HasNext:     page < totalPages,
		HasPrevious: page > 1,
	}

	if meta.HasNext {
		meta.NextURL = fmt.Sprintf("%s?page=%d&limit=%d", basePath, page+1, limit)
	}
	if meta.HasPrevious {
		meta.PreviousURL = fmt.Sprintf("%s?page=%d&limit=%d", basePath, page-1, limit)
	}

	return meta
}

// RespondSuccessWithPagination sends a successful JSend response with pagination
func RespondSuccessWithPagination(w http.ResponseWriter, statusCode int, data any, pagination PaginationMetadata) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := map[string]any{
		"status": "success",
		"data": map[string]any{
			"items":      data,
			"pagination": pagination,
		},
	}

	// Use json.NewEncoder directly like other httpx functions
	// Don't try to write error response since headers are already sent
	_ = json.NewEncoder(w).Encode(response)
}
