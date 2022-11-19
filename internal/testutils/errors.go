package testutils

import (
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// AssertExpectedErrorsHttpReponse - helper method to assert the response errors that happens
// when making API calls
func AssertExpectedErrorsHttpReponse(t *testing.T, expectedStatusCode, actualCode int, expectedError, err error) {
	require.Error(t, err)
	echoError, isOK := err.(*echo.HTTPError)
	if isOK {
		assert.Equalf(t, expectedStatusCode, echoError.Code, "Expected code to be %d, but was %d", expectedStatusCode, echoError.Code)
		expectedEchoError, isOK := expectedError.(*echo.HTTPError)
		if isOK {
			if expectedEchoError.Message != nil {
				assert.Equalf(t, expectedEchoError, echoError, "Expected expectedEchoError to be %s, but was %s", expectedEchoError, echoError)
			}
		}
	} else {
		// assert.Equalf(t, expectedStatusCode, actualCode, "Expected status code to be %d, but was %d")
		assert.Equalf(t, expectedError, err, "Expected err to be %s, but was %s", expectedError, err)
	}
}
