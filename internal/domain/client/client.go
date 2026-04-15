package client

import "time"

type DocumentType string

const (
	DocumentTypeCPF  DocumentType = "CPF"
	DocumentTypeCNPJ DocumentType = "CNPJ"
)

type Client struct {
	ID             string
	DocumentType   DocumentType
	DocumentNumber string
	Name           string
	Email          *string
	Phone          *string
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      *time.Time
}
