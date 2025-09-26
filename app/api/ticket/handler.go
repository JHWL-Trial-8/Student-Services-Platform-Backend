package ticketapi

import ticketsvc "student-services-platform-backend/app/services/ticket"

type Handler struct {
	svc *ticketsvc.Service
}

func New(s *ticketsvc.Service) *Handler {
	return &Handler{svc: s}
}