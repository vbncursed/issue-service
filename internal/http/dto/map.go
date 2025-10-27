package dto

import (
	im "github.com/vbncursed/vkr/issue-service/internal/models"
	issvc "github.com/vbncursed/vkr/issue-service/internal/service"
)

// ToCommand преобразует CreatePassRequest в команду use case
func (r CreatePassRequest) ToCommand() issvc.IssuePassCommand {
	return issvc.IssuePassCommand{
		OrgID:       r.OrgID,
		PolicyID:    r.PolicyID,
		SubjectName: r.SubjectName,
		ZoneID:      r.ZoneID,
		NBF:         r.NBF,
		EXP:         r.EXP,
		OneTime:     r.OneTime,
		Attrs:       r.Attrs,
	}
}

// FromIssueResult формирует ответ по результату use case
func FromIssueResult(res issvc.IssuePassResult) CreatePassResponse {
	return CreatePassResponse{
		ID:          res.ID,
		Status:      string(im.StatusActive),
		IssuerKeyID: res.IssuerKeyID,
		Payload:     res.Payload,
	}
}

// Revoke
func RevokeResponseOK(id string) RevokeResponse {
	return RevokeResponse{ID: id, Status: string(im.StatusRevoked)}
}

// Approve
func FromApproveResult(id string, r issvc.ApproveResult) ApproveResponse {
	return ApproveResponse{ID: id, PickupToken: r.Token, ExpiresAt: r.ExpiresAt}
}

// Pickup
func FromPickupResult(r issvc.PickupResult) PickupResponse {
	return PickupResponse{Payload: r.Payload, IssuerKeyID: r.IssuerKeyID}
}
