package services

import (
	"context"
	"errors"
	"time"

	"go.uber.org/zap"
)

type User struct {
	UID           string     `json:"uid"`
	DisplayName   string     `json:"display_name"`
	Email         string     `json:"email"`
	CreatedAt     time.Time  `json:"created_at"`
	ShowAnonymous bool       `json:"show_anonymous"`
	IsActive      bool       `json:"is_active"`
	IsBanned      bool       `json:"is_banned"`
	BannedReason  *string    `json:"banned_reason,omitempty"`
	ExpPoints     int        `json:"exp_points"`
	Level         int        `json:"level"`
	ReportsCount  int        `json:"reports_count"`
	LastReportAt  *time.Time `json:"last_report_at,omitempty"`
	AvatarURL     *string    `json:"avatar_url,omitempty"`
	Bio           *string    `json:"bio,omitempty"`
}

func GetUserByUID(ctx context.Context, uid string) (*User, error) {
	query := `
		SELECT uid, display_name, email, created_at, show_anonymous,
		       is_active, is_banned, banned_reason, exp_points, level,
		       reports_count, last_report_at, avatar_url, bio
		FROM users
		WHERE uid = $1
	`

	row := DB.QueryRow(ctx, query, uid)

	var u User
	if err := row.Scan(
		&u.UID,
		&u.DisplayName,
		&u.Email,
		&u.CreatedAt,
		&u.ShowAnonymous,
		&u.IsActive,
		&u.IsBanned,
		&u.BannedReason,
		&u.ExpPoints,
		&u.Level,
		&u.ReportsCount,
		&u.LastReportAt,
		&u.AvatarURL,
		&u.Bio,
	); err != nil {
		zap.L().Error("failed to fetch user from db", zap.String("uid", uid), zap.Error(err))
		return nil, errors.New("user not found in database")
	}

	return &u, nil
}
