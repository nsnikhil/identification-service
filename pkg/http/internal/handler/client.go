package handler

import (
	"identification-service/pkg/client"
	"identification-service/pkg/http/contract"
	"identification-service/pkg/http/internal/util"
	"identification-service/pkg/liberr"
	"net/http"
)

type ClientHandler struct {
	service client.Service
}

func (ch *ClientHandler) Register(resp http.ResponseWriter, req *http.Request) error {
	var reqBody contract.CreateClientRequest
	if err := util.ParseRequest(req, &reqBody); err != nil {
		return liberr.WithArgs(liberr.Operation("ClientHandler.Register"), err)
	}

	publicKey, secret, err := ch.service.CreateClient(
		req.Context(),
		reqBody.Name,
		reqBody.AccessTokenTTL,
		reqBody.SessionTTL,
		reqBody.MaxActiveSessions,
		reqBody.SessionStrategy,
	)

	if err != nil {
		return liberr.WithOp("ClientHandler.Register", err)
	}

	respBody := contract.CreateClientResponse{PublicKey: publicKey, Secret: secret}

	util.WriteSuccessResponse(http.StatusCreated, respBody, resp)
	return nil
}

func (ch *ClientHandler) Revoke(resp http.ResponseWriter, req *http.Request) error {
	var reqBody contract.ClientRevokeRequest
	if err := util.ParseRequest(req, &reqBody); err != nil {
		return liberr.WithArgs(liberr.Operation("ClientHandler.Revoke"), err)
	}

	err := ch.service.RevokeClient(req.Context(), reqBody.ID)
	if err != nil {
		return liberr.WithOp("ClientHandler.Revoke", err)
	}

	respBody := contract.ClientRevokeResponse{Message: contract.ClientRevokeSuccessful}

	util.WriteSuccessResponse(http.StatusOK, respBody, resp)
	return nil
}

func NewClientHandler(service client.Service) *ClientHandler {
	return &ClientHandler{
		service: service,
	}
}
