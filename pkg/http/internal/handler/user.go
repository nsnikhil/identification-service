package handler

import (
	"identification-service/pkg/http/contract"
	"identification-service/pkg/http/internal/util"
	"identification-service/pkg/liberr"
	"identification-service/pkg/user"
	"net/http"
)

type UserHandler struct {
	service user.Service
}

func (uh *UserHandler) SignUp(resp http.ResponseWriter, req *http.Request) error {
	var data contract.CreateUserRequest
	if err := util.ParseRequest(req, &data); err != nil {
		return liberr.WithArgs(liberr.Operation("UserHandler.SignUp"), err)
	}

	//TODO: THINK IF THE VALIDATION SHOULD BE DELEGATED TO SVC LAYER ?
	_, err := uh.service.CreateUser(req.Context(), data.Name, data.Email, data.Password)
	if err != nil {
		return liberr.WithArgs(liberr.Operation("UserHandler.SignUp"), err)
	}

	//TODO: WRITE SUCCESS LOG
	util.WriteSuccessResponse(http.StatusCreated, contract.CreateUserResponse{Message: contract.UserCreationSuccess}, resp)
	return nil
}

func (uh *UserHandler) UpdatePassword(resp http.ResponseWriter, req *http.Request) error {
	var data contract.UpdatePasswordRequest
	if err := util.ParseRequest(req, &data); err != nil {
		return liberr.WithArgs(liberr.Operation("UserHandler.UpdatePassword"), err)
	}

	err := uh.service.UpdatePassword(req.Context(), data.Email, data.OldPassword, data.NewPassword)
	if err != nil {
		return liberr.WithArgs(liberr.Operation("UserHandler.UpdatePassword"), err)
	}

	util.WriteSuccessResponse(http.StatusOK, contract.UpdatePasswordResponse{Message: contract.PasswordUpdateSuccess}, resp)
	return nil
}

func NewUserHandler(svc user.Service) *UserHandler {
	return &UserHandler{
		service: svc,
	}
}
