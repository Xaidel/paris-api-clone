package usecases

import (
	entities "github.com/gyud-adb/paris-api/internal/domain/entities"
	ports "github.com/gyud-adb/paris-api/internal/ports/outbound"
)

func newUserResult(user *entities.User) ports.UserResult {
	middleName := user.Profile().MiddleName()

	return ports.UserResult{
		ID:         user.ID().String(),
		Username:   user.Username(),
		FirstName:  user.Profile().FirstName(),
		MiddleName: middleName,
		LastName:   user.Profile().LastName(),
		GroupID:    user.Profile().GroupID().String(),
		CreatedAt:  user.CreatedAt(),
		UpdatedAt:  user.UpdatedAt(),
	}
}
