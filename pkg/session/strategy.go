package session

import (
	"context"
	"errors"
	"fmt"
	"identification-service/pkg/liberr"
	"math"
)

//TODO: REFACTOR THE ENTIRE FILE
type strategy interface {
	apply(ctx context.Context, userID string, currActiveSessions, maxActiveSessions int) error
}

const (
	revokeOldName = "revoke_old"
)

var strategies = map[string]func(store Store) strategy{
	revokeOldName: func(store Store) strategy { return &revokeOld{store: store} },
}

func strategyFromName(name string, store Store) (strategy, error) {
	strategy, ok := strategies[name]
	if !ok {
		return nil, liberr.WithArgs(
			liberr.Operation("SessionStrategy.FromName"),
			liberr.InvalidArgsError,
			fmt.Errorf("invalid session strategy %s", name),
		)
	}

	return strategy(store), nil
}

type revokeOld struct {
	store Store
}

func (ro *revokeOld) apply(ctx context.Context, userID string, currActiveSessions, maxActiveSessions int) error {
	if currActiveSessions < maxActiveSessions {
		return liberr.WithOp(
			"revokeOld.apply",
			errors.New("current active sessions is less than max active sessions allowed"),
		)
	}

	n := int(math.Abs(float64(currActiveSessions-maxActiveSessions))) + 1

	c, err := ro.store.RevokeLastNSessions(ctx, userID, n)
	if err != nil {
		return liberr.WithOp("revokeOld.apply", err)
	}

	if c != int64(n) {
		return liberr.WithOp(
			"revokeOld.apply",
			errors.New("failed to revoke all n old sessions"),
		)
	}

	return nil
}
