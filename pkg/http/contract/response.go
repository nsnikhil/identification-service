package contract

type APIResponse struct {
	Data    interface{} `json:"data,omitempty"`
	Error   *Error      `json:"error,omitempty"`
	Success bool        `json:"success"`
}

type Error struct {
	Message string `json:"message,omitempty"`
}

func NewSuccessResponse(data interface{}) APIResponse {
	return APIResponse{
		Data:    data,
		Success: true,
	}
}

func NewFailureResponse(description string) APIResponse {
	return APIResponse{
		Error: &Error{
			Message: description,
		},
		Success: false,
	}
}
