package testutils

import (
	"testing"
	"user-microservice/internal/models"

	"github.com/stretchr/testify/assert"
)

// AssertUserConfig - extra configuration for test assertion
type AssertUserConfig struct {
	EqualPasswords bool
	CreatedAt      *DateCheck
	UpdatedAt      *DateCheck
}

// AssertUserBody - helper test function to assert the body basic data inside tests
func AssertUserBody(t *testing.T, expected models.User, result models.User, cfg AssertUserConfig) {
	assert.NotEmptyf(t, result.ID, "Expected ID not to be nil")
	assert.Equalf(t, expected.FirstName, result.FirstName, "Expected FirstName to be %s, but was %s", expected.FirstName, result.FirstName)
	assert.Equalf(t, expected.LastName, result.LastName, "Expected LastName to be %s, but was %s", expected.LastName, result.LastName)
	assert.Equalf(t, expected.Nickname, result.Nickname, "Expected Nickname to be %s, but was %s", expected.Nickname, result.Nickname)
	if cfg.EqualPasswords {
		assert.Equalf(t, expected.Password, result.Password, "Expected Password to be %s, but was %s", expected.Password, result.Password)
	}
	assert.Equalf(t, expected.Email, result.Email, "Expected Email to be %s, but was %s", expected.Email, result.Email)
	assert.Equalf(t, expected.Country, result.Country, "Expected Country to be %s, but was %s", expected.Country, result.Country)
	assert.NotZerof(t, result.CreatedAt, "Expected CreatedAt not to be zero")
	assert.NotZerof(t, result.UpdatedAt, "Expected CreatedAt not to be zero")

	if cfg.CreatedAt != nil {
		assertDateCheck(t, *cfg.CreatedAt, "CreatedAt", result.CreatedAt)
	}
	if cfg.UpdatedAt != nil {
		assertDateCheck(t, *cfg.UpdatedAt, "UpdatedAt", result.UpdatedAt)
	}
}
