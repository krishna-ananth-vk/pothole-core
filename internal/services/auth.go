package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"google.golang.org/api/option"
)

type LoginRequest struct {
	IDToken string `json:"idToken"`
}

type LoginResponse struct {
	UID       string    `json:"uid"`
	Email     string    `json:"email,omitempty"`
	Name      string    `json:"name,omitempty"`
	Timestamp time.Time `json:"timestamp"`
	Message   string    `json:"message"`
}

var firebaseAuth *auth.Client

func InitFirebase() error {
	// Path to your Firebase Admin SDK service account key JSON
	opt := option.WithCredentialsFile("serviceAccountKey.json")
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		return fmt.Errorf("error initializing app: %v", err)
	}

	client, err := app.Auth(context.Background())
	if err != nil {
		return fmt.Errorf("error getting Auth client: %v", err)
	}

	firebaseAuth = client
	return nil
}

func LoginUser(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Verify Firebase ID token
	token, err := firebaseAuth.VerifyIDToken(r.Context(), req.IDToken)
	if err != nil {
		http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
		return
	}

	// Fetch user info from Firebase
	userRecord, err := firebaseAuth.GetUser(r.Context(), token.UID)
	if err != nil {
		http.Error(w, "Failed to fetch user", http.StatusInternalServerError)
		return
	}

	resp := LoginResponse{
		UID:       userRecord.UID,
		Email:     userRecord.Email,
		Name:      userRecord.DisplayName,
		Timestamp: time.Now(),
		Message:   "Login successful",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
