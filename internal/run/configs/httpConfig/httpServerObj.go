package httpConfig

import (
	"Demonstration-Service/internal/Application/Contracts/OrdersServices"
	"Demonstration-Service/internal/Presentation/Servers/HTTP"
	"go.uber.org/zap"
	"net/http"
	"net/http/pprof"
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

	mux := http.NewServeMux()

	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

	mux.HandleFunc("/", handler.ServeHTTP)

	return &Server{
		server:  server,
		logger:  logger,
		handler: mux,
	}
}

func (s *Server) Start(address string) error {
	s.logger.Info("HTTP server is starting", zap.String("address", address))
	return http.ListenAndServe(address, s.handler)
}

func (s *Server) Stop() {
	s.logger.Info("HTTP server stopped")
}
