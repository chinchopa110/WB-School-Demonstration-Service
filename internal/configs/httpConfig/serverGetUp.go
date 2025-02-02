package httpConfig

import (
	"Demonstration-Service/internal/Application/Contracts/OrdersServices"
	"Demonstration-Service/internal/Presentation/Servers/HTTP"
	"Demonstration-Service/internal/configs"
	"log"
	"net/http"

	"go.uber.org/zap"
)

func ServerGetUp(service OrdersServices.IGetService) {
	server := HTTP.NewServer(service)

	logger, err := configs.InitLogger("app.log")
	if err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}
	defer func() {
		if err := logger.Sync(); err != nil {
			logger.Error("failed to sync logger", zap.Error(err))
		}
	}()

	logger.Info("HTTP server listening", zap.String("address", ":8080"))
	var handlerAPI http.Handler = http.HandlerFunc(server.ServeHTTP)
	handlerAPI = HTTP.Logging(logger, handlerAPI)
	handlerAPI = HTTP.PanicRecovery(logger, handlerAPI)

	if err := http.ListenAndServe(":8080", handlerAPI); err != nil {
		logger.Fatal("failed to start http server", zap.Error(err))
	}
}
