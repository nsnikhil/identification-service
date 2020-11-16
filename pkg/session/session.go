package session

import (
	"fmt"
	"identification-service/pkg/liberr"
	"identification-service/pkg/util"
	"time"
)

type Session struct {
	id string

	userID       string
	refreshToken string

	revoked bool

	createdAt time.Time
	updatedAt time.Time
}

func (s Session) IsExpired(ttl float64) bool {
	return time.Now().Sub(s.createdAt).Minutes() >= ttl
}

type Builder struct {
	id string

	userID       string
	refreshToken string

	revoked bool

	createdAt time.Time
	updatedAt time.Time

	err error
}

func (b *Builder) ID(id string) *Builder {
	if b.err != nil {
		return b
	}

	if !util.IsValidUUID(id) {
		b.err = fmt.Errorf("invalid id %s", id)
		return b
	}

	b.id = id
	return b
}

func (b *Builder) UserID(userID string) *Builder {
	if b.err != nil {
		return b
	}

	if !util.IsValidUUID(userID) {
		b.err = fmt.Errorf("invalid user id %s", userID)
		return b
	}

	b.userID = userID
	return b
}

func (b *Builder) RefreshToken(refreshToken string) *Builder {
	if b.err != nil {
		return b
	}

	if !util.IsValidUUID(refreshToken) {
		b.err = fmt.Errorf("invalid id %s", refreshToken)
		return b
	}

	b.refreshToken = refreshToken
	return b
}

func (b *Builder) Revoked(revoked bool) *Builder {
	if b.err != nil {
		return b
	}

	b.revoked = revoked
	return b
}

func (b *Builder) CreatedAt(createdAt time.Time) *Builder {
	if b.err != nil {
		return b
	}

	b.createdAt = createdAt
	return b
}

func (b *Builder) UpdatedAt(updatedAt time.Time) *Builder {
	if b.err != nil {
		return b
	}

	b.updatedAt = updatedAt
	return b
}

func (b *Builder) Build() (Session, error) {
	if b.err != nil {
		return Session{}, liberr.WithOp("Builder.Build", b.err)
	}

	return Session{
		id:           b.id,
		userID:       b.userID,
		refreshToken: b.refreshToken,
		revoked:      b.revoked,
		createdAt:    b.createdAt,
		updatedAt:    b.updatedAt,
	}, nil
}

func NewSessionBuilder() *Builder {
	return &Builder{}
}
