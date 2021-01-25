package contract

const LogoutSuccessfulMessage = "Logout Successful"

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (lr LoginRequest) IsValid() error {
	return isValid("LoginRequest.IsValid",
		pair{name: "email", data: lr.Email},
		pair{name: "password", data: lr.Password},
	)
}

type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type RefreshTokenResponse struct {
	AccessToken string `json:"access_token"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func (lr LogoutRequest) IsValid() error {
	return isValid("LogoutRequest.IsValid",
		pair{name: "refresh token", data: lr.RefreshToken},
	)
}

type LogoutResponse struct {
	Message string `json:"message"`
}
