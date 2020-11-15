package contract

const ClientRevokeSuccessful = "client revoked successfully"

type CreateClientRequest struct {
	Name              string `json:"name"`
	AccessTokenTTL    int    `json:"access_token_ttl"`
	SessionTTL        int    `json:"session_ttl"`
	MaxActiveSessions int    `json:"max_active_sessions"`
}

type CreateClientResponse struct {
	PublicKey string `json:"public_key"`
	Secret    string `json:"secret"`
}

type ClientRevokeRequest struct {
	ID string `json:"id"`
}

type ClientRevokeResponse struct {
	Message string `json:"message"`
}
