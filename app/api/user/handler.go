package userapi

import usersvc "student-services-platform-backend/app/services/user"

type Handler struct {
	svc *usersvc.Service
}

func New(s *usersvc.Service) *Handler {
	return &Handler{svc: s}
}