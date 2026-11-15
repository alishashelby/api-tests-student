package managers

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

// ErrorResponse описывает стандартное тело ошибки API
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

// ReadBody проверяет статус, читает и закрывает тело HTTP-ответа
func ReadBody(t *testing.T, response *http.Response, expectedStatusCodes ...int) []byte {
	t.Helper()
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	require.NoError(t, err, "не удалось прочитать тело HTTP-ответа")
	require.Containsf(
		t,
		expectedStatusCodes,
		response.StatusCode,
		"HTTP status code должен входить в %v, тело ответа: %s",
		expectedStatusCodes,
		string(body),
	)

	return body
}

// Unmarshal преобразует JSON в типизированную модель
func Unmarshal[T any](t *testing.T, body []byte) T {
	t.Helper()

	var result T
	require.NoErrorf(t, json.Unmarshal(body, &result), "не удалось разобрать JSON-ответ: %s", string(body))

	return result
}
