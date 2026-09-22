package service

import "github.com/google/uuid"

func mustParseUUID(s string) uuid.UUID {
	u, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil
	}
	return u
}
