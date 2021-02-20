package liberr

type Kind string

const (
	ValidationError         Kind = "validationError"
	InternalError           Kind = "internalError"
	AuthenticationError     Kind = "authenticationError"
	InvalidCredentialsError Kind = "invalidCredentialsError"
	InvalidArgsError        Kind = "invalidArgsError"
	InitializationError     Kind = "initializationError"
	//TODO: SUFFIX WITH ERROR
	ResourceNotFound     Kind = "resourceNotFound"
	DuplicateRecordError Kind = "duplicateRecordError"
	ProducerError        Kind = "producerError"
	ConsumerError        Kind = "consumerError"
)
