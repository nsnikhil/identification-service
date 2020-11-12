package contract

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

type UpdatePasswordResponse struct {
	Message string `json:"message"`
}
