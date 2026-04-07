package models

import "time"

type User struct {
	ID           int       `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
}

type State struct {
	ID           int     `json:"id"`
	UF           string  `json:"uf"`
	Nome         string  `json:"nome"`
	Regiao       string  `json:"regiao"`
	Capital      string  `json:"capital"`
	Populacao    int64   `json:"populacao"`
	PIBPerCapita float64 `json:"pib_per_capita"`
	Filiais      *int    `json:"filiais_ativas,omitempty"`
}

type Branch struct {
	ID        int        `json:"id"`
	Nome      string     `json:"nome"`
	Cidade    string     `json:"cidade"`
	StateID   int        `json:"state_id"`
	UF        string     `json:"uf,omitempty"`
	Lat       *float64   `json:"lat,omitempty"`
	Lng       *float64   `json:"lng,omitempty"`
	OpenedAt  string     `json:"opened_at"`
	CreatedAt time.Time  `json:"created_at"`
	DeletedAt *time.Time `json:"-"`
}

type PaginatedBranches struct {
	Data       []Branch `json:"data"`
	Page       int      `json:"page"`
	Limit      int      `json:"limit"`
	Total      int      `json:"total"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type BranchRequest struct {
	Nome     string   `json:"nome"`
	Cidade   string   `json:"cidade"`
	UF       string   `json:"uf"`
	Lat      *float64 `json:"lat"`
	Lng      *float64 `json:"lng"`
	OpenedAt string   `json:"opened_at"`
}

type ErrorResponse struct {
	Error  string            `json:"error"`
	Fields map[string]string `json:"fields,omitempty"`
}