package HTTP

import (
	"go.uber.org/zap"
	"net/http"
	"runtime/debug"
	"time"
)

func Logging(logger *zap.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, req)
		duration := time.Since(start)
		logger.Info("request",
			zap.String("method", req.Method),
			zap.String("uri", req.RequestURI),
			zap.Duration("duration", duration),
		)
	})
}

func PanicRecovery(logger *zap.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				logger.Error("panic recovery",
					zap.Any("error", err),
					zap.String("stack", string(debug.Stack())),
				)
			}
		}()
		next.ServeHTTP(w, req)
	})
}
