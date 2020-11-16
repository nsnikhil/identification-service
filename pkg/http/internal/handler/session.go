package handler

import (
	"identification-service/pkg/http/contract"
	"identification-service/pkg/http/internal/util"
	"identification-service/pkg/liberr"
	"identification-service/pkg/session"
	"net/http"
)

type SessionHandler struct {
	service session.Service
}

func (sh *SessionHandler) Login(resp http.ResponseWriter, req *http.Request) error {
	var data contract.LoginRequest
	if err := util.ParseRequest(req, &data); err != nil {
		return liberr.WithArgs(liberr.Operation("UserHandler.LoginUser"), err)
	}

	accessToken, refreshToken, err := sh.service.LoginUser(req.Context(), data.Email, data.Password)
	if err != nil {
		return liberr.WithOp("SessionHandler.Login", err)
	}

	respData := contract.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	util.WriteSuccessResponse(http.StatusCreated, respData, resp)
	return nil
}

func (sh *SessionHandler) RefreshToken(resp http.ResponseWriter, req *http.Request) error {
	var data contract.RefreshTokenRequest
	if err := util.ParseRequest(req, &data); err != nil {
		return liberr.WithArgs(liberr.Operation("UserHandler.RefreshToken"), err)
	}

	accessToken, err := sh.service.RefreshToken(req.Context(), data.RefreshToken)
	if err != nil {
		return liberr.WithOp("SessionHandler.RefreshToken", err)
	}

	respData := contract.RefreshTokenResponse{
		AccessToken: accessToken,
	}

	util.WriteSuccessResponse(http.StatusOK, respData, resp)
	return nil
}

func (sh *SessionHandler) Logout(resp http.ResponseWriter, req *http.Request) error {
	var data contract.LogoutRequest
	if err := util.ParseRequest(req, &data); err != nil {
		return liberr.WithArgs(liberr.Operation("UserHandler.RefreshToken"), err)
	}

	err := sh.service.LogoutUser(req.Context(), data.RefreshToken)
	if err != nil {
		return liberr.WithOp("SessionHandler.Logout", err)
	}

	respData := contract.LogoutResponse{
		Message: contract.LogoutSuccessfulMessage,
	}

	util.WriteSuccessResponse(http.StatusOK, respData, resp)
	return nil
}

func NewSessionHandler(service session.Service) *SessionHandler {
	return &SessionHandler{
		service: service,
	}
}
