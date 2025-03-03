package middlewares

import (
	"net/http"
)

type ErrorHandler func(w http.ResponseWriter, r *http.Request)

var (
	Handle400Error ErrorHandler
	Handle500Error ErrorHandler
)

func SetErrorHandlers(h400, h500 ErrorHandler) {
	Handle400Error = h400
	Handle500Error = h500
}

func ErrorMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Créer un wrapper pour ResponseWriter pour capturer le code de statut
		rw := &responseWriter{w, http.StatusOK}
		// Appeler le handler suivant
		next.ServeHTTP(rw, r)

		// Vérifier le code de statut et rediriger si nécessaire
		switch rw.status {
		case http.StatusBadRequest:
			if Handle400Error != nil {
				Handle400Error(w, r)
			}
		case http.StatusInternalServerError:
			if Handle500Error != nil {
				Handle500Error(w, r)
			}
		}
	})
}

// Wrapper pour ResponseWriter pour capturer le code de statut
type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(status int) {
	rw.status = status
	rw.ResponseWriter.WriteHeader(status)
}
