package event

type Code string

const (
	SignUp         Code = "sign-up"
	UpdatePassword Code = "update-password"
)

//TODO: REFACTOR THIS, TWO SOURCE OF TRUTH, CONFIG AND CODE ??
var CodeMap = map[Code]bool{
	SignUp:         true,
	UpdatePassword: true,
}
