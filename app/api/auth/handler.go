package authapi

import "student-services-platform-backend/app/services/auth"

type Handler struct {
	svc *auth.Service
}

func New(s *auth.Service) *Handler {
	return &Handler{svc: s}
}