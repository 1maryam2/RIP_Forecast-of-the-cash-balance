package ds

import (
	"time"
)

type LoginRequest struct {
	Login    string `json:"login" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type RegisterRequest struct {
	Login       string `json:"login" binding:"required"`
	Password    string `json:"password" binding:"required"`
	IsModerator bool   `json:"is_moderator,omitempty"`
}

type UpdateUserRequest struct {
	Login       string `json:"login,omitempty"`
	IsModerator *bool  `json:"is_moderator,omitempty"`
}

type CreateAccountRequest struct {
	Code        string `json:"code" binding:"required"`
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
	Type        string `json:"type" binding:"required"`
	Category    string `json:"category"`
}

type UpdateAccountRequest struct {
	Code        string `json:"code,omitempty"`
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	Type        string `json:"type,omitempty"`
	Category    string `json:"category,omitempty"`
}

type AccountsFilter struct {
	Search   string `json:"search,omitempty"`
	Type     string `json:"type,omitempty"`
	Category string `json:"category,omitempty"`
}

type AddToFundsApplicationRequest struct {
	AccountID uint `json:"account_id" binding:"required"`
}

type UpdateFundsApplicationRequest struct {
	CompanyName string  `json:"company_name,omitempty"`
	INN         string  `json:"inn,omitempty"`
	OGRN        string  `json:"ogrn,omitempty"`
	InitialSum  float64 `json:"initial_sum,omitempty"`
	Quarter     int     `json:"quarter,omitempty"`
}

type UpdateFundsApplicationItemRequest struct {
	Amount  float64 `json:"amount" binding:"required"`
	Comment string  `json:"comment,omitempty"`
}

type ApplicationFilter struct {
	Status   FundsApplicationStatus `json:"status,omitempty"`
	DateFrom *time.Time             `json:"date_from,omitempty"`
	DateTo   *time.Time             `json:"date_to,omitempty"`
}
