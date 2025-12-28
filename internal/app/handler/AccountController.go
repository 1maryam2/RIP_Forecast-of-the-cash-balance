package handler

import (
	"lab_1/internal/app/repository"
)

type Handler struct {
	Repository   *repository.Repository
	EventManager *EventManager
}

func NewHandler(r *repository.Repository, em *EventManager) *Handler {
	return &Handler{
		Repository:   r,
		EventManager: em,
	}
}
