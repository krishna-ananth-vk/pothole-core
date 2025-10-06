package main

import (
	"fmt"
	"log"
	constant_variables "nammablr/pothole/internal/constants"
	"nammablr/pothole/internal/services"
	"nammablr/pothole/pkg/logger"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func main() {
	logPath := "pothole-core.log"
	zapLogger, err := logger.SetupLogger(logPath)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer zapLogger.Sync()
	err = services.InitFirebase()
	if err != nil {
		zapLogger.Sugar().Fatalf("Firebase initialisation failed: %v", err)
	}
	r := chi.NewRouter()
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			next.ServeHTTP(w, r)
			zapLogger.Info("request completed",
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.Duration("duration", time.Since(start)),
			)
		})
	})
	r.Route(constant_variables.CONTEXT, func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("Hello World!"))
		})

		// r.Get("/users", services.GetUsers)
		r.Get("/health", services.GetHealth)

		// login user
		r.Post("/login", services.LoginUser)

		// Image upload
		r.Post("/upload", services.UploadImage)
	})

	fmt.Printf("App has started on %s", constant_variables.PORT)

	zapLogger.Sugar().Infof("App has started on %s", constant_variables.PORT)
	http.ListenAndServe(constant_variables.PORT, r)
}
