package model

import "github.com/google/uuid"

func IsValidUUID(u string) bool {
	_, err := uuid.Parse(u)
	return err == nil
}

func GenStringUUID() string {
	return uuid.New().String()
}
