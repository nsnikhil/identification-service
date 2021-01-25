package contract

import (
	"fmt"
	"identification-service/pkg/liberr"
)

const (
	UserCreationSuccess   = "user created successfully"
	PasswordUpdateSuccess = "password updated successfully"
)

type CreateUserRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type CreateUserResponse struct {
	Message string `json:"message"`
}

type UpdatePasswordRequest struct {
	Email       string `json:"email"`
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

//NOTE: DOMAIN VALIDATION WILL NOT HAPPEN HERE, THIS IS JUST FOR SANITY
func (upr UpdatePasswordRequest) IsValid() error {
	type pair struct {
		name string
		data string
	}

	isNotEmpty := func(pr ...pair) error {
		for _, p := range pr {
			if len(p.data) == 0 {
				return liberr.WithArgs(
					liberr.Operation("UpdatePasswordRequest.IsValid"),
					liberr.ValidationError,
					fmt.Errorf("%s cannot be empty", p.name),
				)
			}
		}

		return nil
	}

	return isNotEmpty(
		pair{name: "email", data: upr.Email},
		pair{name: "old password", data: upr.OldPassword},
		pair{name: "new password", data: upr.NewPassword},
	)
}

type UpdatePasswordResponse struct {
	Message string `json:"message"`
}
