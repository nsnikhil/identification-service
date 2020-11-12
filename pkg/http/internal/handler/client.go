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
	var data contract.CreateClientRequest
	if err := util.ParseRequest(req, &data); err != nil {
		return liberr.WithArgs(liberr.Operation("ClientHandler.Register"), err)
	}

	secret, err := ch.service.CreateClient(data.Name, data.AccessTokenTTL, data.SessionTTL)
	if err != nil {
		return liberr.WithOp("ClientHandler.Register", err)
	}

	util.WriteSuccessResponse(http.StatusCreated, contract.CreateClientResponse{Secret: secret}, resp)
	return nil
}

func (ch *ClientHandler) Revoke(resp http.ResponseWriter, req *http.Request) error {
	var data contract.ClientRevokeRequest
	if err := util.ParseRequest(req, &data); err != nil {
		return liberr.WithArgs(liberr.Operation("ClientHandler.Revoke"), err)
	}

	err := ch.service.RevokeClient(data.ID)
	if err != nil {
		return liberr.WithOp("ClientHandler.Revoke", err)
	}

	util.WriteSuccessResponse(http.StatusOK, contract.ClientRevokeResponse{Message: contract.ClientRevokeSuccessful}, resp)
	return nil
}

func NewClientHandler(service client.Service) *ClientHandler {
	return &ClientHandler{
		service: service,
	}
}
