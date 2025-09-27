package cannedapi

import cannedsvc "student-services-platform-backend/app/services/canned"

type Handler struct {
	svc *cannedsvc.Service
}

func New(s *cannedsvc.Service) *Handler {
	return &Handler{svc: s}
}