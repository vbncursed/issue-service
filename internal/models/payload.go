package models

import "time"

type PayloadPass struct {
	ID         string         `json:"id"`
	Type       string         `json:"type"`
	Level      string         `json:"level"`
	Scopes     []string       `json:"scopes"`
	OneTime    bool           `json:"one_time"`
	NBF        time.Time      `json:"nbf"`
	EXP        time.Time      `json:"exp"`
	Attrs      map[string]any `json:"attrs"`
	HolderHint string         `json:"holder_hint"`
}

type PayloadMeta struct {
	OrgID         string    `json:"org_id"`
	PolicyID      string    `json:"policy_id"`
	ZoneContext   string    `json:"zone_context"`
	IssuedAt      time.Time `json:"issued_at"`
	Nonce         []byte    `json:"nonce"`
	SchemaVersion int       `json:"schema_version"`
}

type SignedPayload struct {
	V           int         `json:"v"`
	Pass        PayloadPass `json:"pass"`
	Meta        PayloadMeta `json:"meta"`
	IssuerKeyID string      `json:"issuer_key_id"`
}
