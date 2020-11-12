package resperr

type ResponseError struct {
	statusCode  int
	description string
}

func (re ResponseError) StatusCode() int {
	return re.statusCode
}

func (re ResponseError) Description() string {
	return re.description
}

func NewResponseError(statusCode int, description string) ResponseError {
	return ResponseError{
		statusCode:  statusCode,
		description: description,
	}
}
