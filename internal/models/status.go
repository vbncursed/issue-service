package models

// PassStatus — доменный статус пропуска
type PassStatus string

const (
	StatusActive  PassStatus = "Active"
	StatusRevoked PassStatus = "Revoked"
	StatusExpired PassStatus = "Expired"
)
