package session

import (
	"context"
	"errors"
	"github.com/nsnikhil/erx"
	"math"
)

//TODO: REFACTOR THE ENTIRE FILE
type Strategy interface {
	Apply(ctx context.Context, userID string, currActiveSessions, maxActiveSessions int) error
}

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
		return erx.WithArgs(
			erx.Operation("RevokeOld.Apply"),
			errors.New("current active sessions is less than max active sessions allowed"),
		)
	}

	n := int(math.Abs(float64(currActiveSessions-maxActiveSessions))) + 1

	c, err := ro.store.RevokeLastNSessions(ctx, userID, n)
	if err != nil {
		return erx.WithArgs(erx.Operation("RevokeOld.Apply"), err)
	}

	if c != int64(n) {
		return erx.WithArgs(
			erx.Operation("RevokeOld.Apply"),
			errors.New("failed to revoke all n old sessions"),
		)
	}

	return nil
}
