package middleware

import (
	"net/http"

	"go.uber.org/zap"

	"github.com/Benzogang-Tape/Reddit/internal/models/errs"
)

func Panic(next http.Handler, logger *zap.SugaredLogger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				logger.Errorw("panicMiddleware",
					"method", r.Method,
					"remote_addr", r.RemoteAddr,
					"url", r.URL.Path,
				)

				http.Error(w, errs.ErrInternalServerError.Error(), http.StatusInternalServerError)
				logger.Infow("recovered", "cause", err)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
