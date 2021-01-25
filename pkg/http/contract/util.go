package contract

import (
	"fmt"
	"identification-service/pkg/liberr"
)

type pair struct {
	name string
	data string
}

func isValid(op string, pr ...pair) error {
	for _, p := range pr {
		if len(p.data) == 0 {
			return liberr.WithArgs(
				liberr.Operation(op),
				liberr.ValidationError,
				fmt.Errorf("%s cannot be empty", p.name),
			)
		}
	}

	return nil
}
