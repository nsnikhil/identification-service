package contract

const ClientRevokeSuccessful = "client revoked successfully"

type CreateClientRequest struct {
	Name           string `json:"name"`
	AccessTokenTTL int    `json:"access_token_ttl"`
	SessionTTL     int    `json:"session_ttl"`
}

type CreateClientResponse struct {
	Secret string `json:"secret"`
}

type ClientRevokeRequest struct {
	ID string `json:"id"`
}

type ClientRevokeResponse struct {
	Message string `json:"message"`
}
