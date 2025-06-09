package handler

import "github.com/sz-po/go-distributed-kvm-switch/internal/pkg/process"

type FindResponseProcess struct {
	Name          process.Name          `json:"name"`
	Specification process.Specification `json:"specification"`
	Status        process.Status        `json:"status"`
}

type FindResponse struct {
	Processes []FindResponseProcess `json:"processes"`
	Count     int                   `json:"count"`
}

type CreateRequest struct {
	Specification process.Specification `json:"specification"`
}

type CreateResponse struct {
	Specification process.Specification `json:"specification"`
	Status        process.Status        `json:"status"`
}

type StartResponse struct {
	Status process.Status `json:"status"`
}

type GetByNameResponse struct {
	Name          process.Name          `json:"name"`
	Specification process.Specification `json:"specification"`
	Status        process.Status        `json:"status"`
}
