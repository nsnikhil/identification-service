package session

import (
	"context"
	"errors"
	"identification-service/pkg/liberr"
	"math"
)

//TODO: REFACTOR THE ENTIRE FILE
type Strategy interface {
	Apply(ctx context.Context, userID string, currActiveSessions, maxActiveSessions int) error
}

//const (
//	revokeOldName = "revoke_old"
//)

//var strategies = map[string]func(store Store) Strategy{
//	revokeOldName: func(store Store) Strategy { return &RevokeOld{store: store} },
//}
//
//func strategyFromName(name string, store Store) (Strategy, error) {
//	strategy, ok := strategies[name]
//	if !ok {
//		return nil, liberr.WithArgs(
//			liberr.Operation("SessionStrategy.FromName"),
//			liberr.InvalidArgsError,
//			fmt.Errorf("invalid session Strategy %s", name),
//		)
//	}
//
//	return strategy(store), nil
//}

type RevokeOld struct {
	store Store
}

func NewRevokeOldStrategy(store Store) *RevokeOld {
	return &RevokeOld{
		store: store,
	}
}

func (ro *RevokeOld) Apply(ctx context.Context, userID string, currActiveSessions, maxActiveSessions int) error {
	if currActiveSessions < maxActiveSessions {
		return liberr.WithOp(
			"RevokeOld.Apply",
			errors.New("current active sessions is less than max active sessions allowed"),
		)
	}

	n := int(math.Abs(float64(currActiveSessions-maxActiveSessions))) + 1

	c, err := ro.store.RevokeLastNSessions(ctx, userID, n)
	if err != nil {
		return liberr.WithOp("RevokeOld.Apply", err)
	}

	if c != int64(n) {
		return liberr.WithOp(
			"RevokeOld.Apply",
			errors.New("failed to revoke all n old sessions"),
		)
	}

	return nil
}
