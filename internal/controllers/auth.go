package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"nammablr/pothole/internal/services"
	"net/http"
	"time"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"go.uber.org/zap"
	"google.golang.org/api/option"
)

type CreateUserRequest struct {
	UID           string  `json:"uid"` // Required: Firebase UID
	DisplayName   string  `json:"display_name"`
	AvatarURL     *string `json:"avatar_url,omitempty"`
	ShowAnonymous *bool   `json:"show_anonymous,omitempty"`
	IsActive      *bool   `json:"is_active,omitempty"`
}

type CreateUserResponse struct {
	User      *services.User `json:"user"`
	Timestamp time.Time      `json:"timestamp"`
	Message   string         `json:"message"`
}

type LoginRequest struct {
	IDToken string `json:"idToken"`
}

type LoginResponse struct {
	UID           string         `json:"uid"`
	DisplayName   string         `json:"display_name"`
	Email         string         `json:"email"`
	CreatedAt     time.Time      `json:"created_at"`
	ShowAnonymous bool           `json:"show_anonymous"`
	IsActive      bool           `json:"is_active"`
	IsBanned      bool           `json:"is_banned"`
	BannedReason  string         `json:"banned_reason,omitempty"`
	ExpPoints     int            `json:"exp_points"`
	Level         int            `json:"level"`
	ReportsCount  int            `json:"reports_count"`
	LastReportAt  time.Time      `json:"last_report_at,omitempty"`
	AvatarURL     string         `json:"avatar_url,omitempty"`
	Bio           string         `json:"bio,omitempty"`
	Timestamp     time.Time      `json:"timestamp"`
	Message       string         `json:"message"`
	UserExists    bool           `json:"user_exist"`
	User          *services.User `json:"user"`
}

var firebaseAuth *auth.Client

func InitFirebase() error {
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

	token, err := firebaseAuth.VerifyIDToken(r.Context(), req.IDToken)
	if err != nil {
		http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
		return
	}

	userRecord, err := firebaseAuth.GetUser(r.Context(), token.UID)
	if err != nil {
		http.Error(w, "Failed to fetch user", http.StatusInternalServerError)
		return
	}

	userExists := true

	dbUser, err := services.GetUserByUID(r.Context(), userRecord.UID)

	if err != nil {
		zap.L().Warn("user not found in DB, using Firebase record only", zap.String("uid", userRecord.UID))
		dbUser = &services.User{
			UID:           userRecord.UID,
			DisplayName:   userRecord.DisplayName,
			Email:         userRecord.Email,
			CreatedAt:     time.Now(),
			ShowAnonymous: true,
			IsActive:      true,
			IsBanned:      false,
			ExpPoints:     0,
			Level:         1,
			ReportsCount:  0,
		}
		userExists = false
	}

	resp := LoginResponse{
		UID:         userRecord.UID,
		User:        dbUser,
		Email:       userRecord.Email,
		DisplayName: dbUser.DisplayName,
		Timestamp:   time.Now(),
		Message:     "Login successful",
		UserExists:  userExists,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func CreateUser(w http.ResponseWriter, r *http.Request) {
	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if req.UID == "" || req.DisplayName == "" {
		http.Error(w, "UID and DisplayName are required", http.StatusBadRequest)
		return
	}

	// Set default values if nil
	showAnonymous := true
	if req.ShowAnonymous != nil {
		showAnonymous = *req.ShowAnonymous
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	createdAt := time.Now()

	// Insert into DB using pgx/v5
	query := `
		INSERT INTO users (uid, display_name, email, created_at, show_anonymous, is_active, is_banned, exp_points, level, reports_count)
		VALUES ($1, $2, $3, $4, $5, $6, false, 0, 1, 0)
		ON CONFLICT (uid) DO UPDATE
		SET display_name = EXCLUDED.display_name,
		    show_anonymous = EXCLUDED.show_anonymous,
		    is_active = EXCLUDED.is_active
		RETURNING uid, display_name, email, created_at, show_anonymous,
		          is_active, is_banned, banned_reason, exp_points, level,
		          reports_count, last_report_at, avatar_url, bio
	`

	email := ""
	row := services.DB.QueryRow(context.Background(), query,
		req.UID,
		req.DisplayName,
		email,
		createdAt,
		showAnonymous,
		isActive,
		req.UID, // uid again for conflict
	)

	var user services.User
	if err := row.Scan(
		&user.UID,
		&user.DisplayName,
		&user.Email,
		&user.CreatedAt,
		&user.ShowAnonymous,
		&user.IsActive,
		&user.IsBanned,
		&user.BannedReason,
		&user.ExpPoints,
		&user.Level,
		&user.ReportsCount,
		&user.LastReportAt,
		&user.AvatarURL,
		&user.Bio,
	); err != nil {
		zap.L().Error("failed to insert or update user", zap.Error(err))
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	resp := CreateUserResponse{
		User:      &user,
		Timestamp: time.Now(),
		Message:   "User created successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
