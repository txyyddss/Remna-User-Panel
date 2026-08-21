package httpapi

import "github.com/go-chi/chi/v5"

func (s *Server) mountMemberOperations(router chi.Router) {
	router.Get("/api/v1/me/traffic-reset-automation", s.trafficResetAutomation)
	router.Put("/api/v1/me/traffic-reset-automation", s.updateTrafficResetAutomation)
	router.Get("/api/v1/purchases/{id}/traffic-reset", s.trafficResetQuote)
	router.Post("/api/v1/purchases/{id}/traffic-reset", s.trafficReset)
	router.Get("/api/v1/purchases/{id}/refund", s.memberRefundQuote)
	router.Post("/api/v1/purchases/{id}/refund", s.memberRefund)
	router.Post("/api/v1/subscription/connections", s.requestConnections)
	router.Get("/api/v1/subscription/connections/{id}", s.pollConnections)
	router.Post("/api/v1/subscription/connections/drop", s.dropConnection)
	router.Get("/api/v1/subscription/ip-blocks", s.memberIPBlocks)
	router.Post("/api/v1/subscription/ip-blocks/{blockId}/unblock", s.memberUnblockIP)
	router.Get("/api/v1/operations/{id}", s.memberOperation)
}
