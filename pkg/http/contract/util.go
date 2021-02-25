package contract

import (
	"fmt"
	"github.com/nsnikhil/erx"
)

type pair struct {
	name string
	data string
}

func isValid(op string, pr ...pair) error {
	for _, p := range pr {
		if len(p.data) == 0 {
			return erx.WithArgs(
				erx.Operation(op),
				erx.ValidationError,
				fmt.Errorf("%s cannot be empty", p.name),
			)
		}
	}

	return nil
}
