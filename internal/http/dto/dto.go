package dto

import "time"

type CreatePassRequest struct {
	OrgID       string         `json:"org_id"`
	PolicyID    string         `json:"policy_id"`
	SubjectName string         `json:"subject_name"`
	ZoneID      string         `json:"zone_id"`
	NBF         time.Time      `json:"nbf"`
	EXP         time.Time      `json:"exp"`
	OneTime     bool           `json:"one_time"`
	Attrs       map[string]any `json:"attrs"`
}

type CreatePassResponse struct {
	ID          string `json:"id"`
	Status      string `json:"status"`
	IssuerKeyID string `json:"issuer_key_id"`
	Payload     string `json:"payload"`
}

type RevokeResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

type ApproveResponse struct {
	ID          string `json:"id"`
	PickupToken string `json:"pickup_token"`
	ExpiresAt   string `json:"expires_at"`
}

type PickupRequest struct {
	Token string `json:"token"`
}

type PickupResponse struct {
	Payload     string `json:"payload"`
	IssuerKeyID string `json:"issuer_key_id"`
}
