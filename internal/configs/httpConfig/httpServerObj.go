package httpConfig

import (
	"Demonstration-Service/internal/Application/Contracts/OrdersServices"
	"Demonstration-Service/internal/Presentation/Servers/HTTP"
	"net/http"

	"go.uber.org/zap"
)

type Server struct {
	server  *HTTP.Server
	logger  *zap.Logger
	handler http.Handler
}

func NewServer(service OrdersServices.IGetService, logger *zap.Logger) *Server {
	server := HTTP.NewServer(service)

	var handler http.Handler = http.HandlerFunc(server.ServeHTTP)
	handler = HTTP.Logging(logger, handler)
	handler = HTTP.PanicRecovery(logger, handler)

	return &Server{
		server:  server,
		logger:  logger,
		handler: handler,
	}
}

func (s *Server) Start(address string) error {
	s.logger.Info("HTTP server is starting", zap.String("address", address))
	return http.ListenAndServe(address, s.handler)
}

func (s *Server) Stop() {
	s.logger.Info("HTTP server stopped")
}