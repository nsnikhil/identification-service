package liberr

type Kind string

const (
	ValidationError         Kind = "validationError"
	InternalError           Kind = "internalError"
	AuthenticationError     Kind = "authenticationError"
	InvalidCredentialsError Kind = "invalidCredentialsError"
	ResourceNotFound        Kind = "resourceNotFound"
)
