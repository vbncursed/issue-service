package repo

const (
	tableIssuerKeys   = "issuer_keys"
	tablePasses       = "passes"
	tablePickupTokens = "pickup_tokens"
)

const (
	colID           = "id"
	colStatus       = "status"
	colCreatedAt    = "created_at"
	colKeyID        = "key_id"
	colAlg          = "alg"
	colPublicKey    = "public_key"
	colPrivateKey   = "private_key"
	colIssuerKeyID  = "issuer_key_id"
	colSignature    = "signature"
	colPayload      = "payload"
	colOrgID        = "org_id"
	colPolicyID     = "policy_id"
	colSubjectName  = "subject_name"
	colZoneID       = "zone_id"
	colNbf          = "nbf"
	colExp          = "exp"
	colOneTime      = "one_time"
	colToken        = "token"
	colPassID       = "pass_id"
	colTTLExpiresAt = "ttl_expires_at"
	colUsedAt       = "used_at"
)
