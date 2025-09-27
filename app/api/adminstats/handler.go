package adminstatsapi

import adminstatssvc "student-services-platform-backend/app/services/adminstats"

type Handler struct {
	svc *adminstatssvc.Service
}

func New(s *adminstatssvc.Service) *Handler {
	return &Handler{svc: s}
}