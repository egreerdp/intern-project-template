package api

import (
	"log"
	"net/http"
)

type Service struct {
	http.Server
	handler *Handler
}

func NewService(handler *Handler) *Service {
	return &Service{
		handler: handler,
	}
}

func (s *Service) Start() {
	if err := http.ListenAndServe("0.0.0.0:3000", s.handler.MountRoutes()); err != nil {
		log.Fatal(err)
	}
}
