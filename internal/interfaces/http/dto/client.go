package dto

type CreateClientRequest struct {
	DocumentType   string  `json:"document_type" binding:"required"`
	DocumentNumber string  `json:"document_number" binding:"required"`
	Name           string  `json:"name" binding:"required"`
	Email          *string `json:"email"`
	Phone          *string `json:"phone"`
}

type ClientResponse struct {
	ID             string  `json:"id"`
	DocumentType   string  `json:"document_type"`
	DocumentNumber string  `json:"document_number"`
	Name           string  `json:"name"`
	Email          *string `json:"email,omitempty"`
	Phone          *string `json:"phone,omitempty"`
}
