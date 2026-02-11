package handlers

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRespondWithProblem(t *testing.T) {
	t.Run("writes correct status and body", func(t *testing.T) {
		rec := httptest.NewRecorder()

		respondWithProblem(rec, 400, "VALIDATION_ERROR", "Invalid input")

		assert.Equal(t, 400, rec.Code)
		assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

		var resp map[string]interface{}
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
		assert.False(t, resp["success"].(bool))
		errObj := resp["error"].(map[string]interface{})
		assert.Equal(t, "VALIDATION_ERROR", errObj["code"])
		assert.Equal(t, "Invalid input", errObj["message"])
	})
}

func TestRespondWithValidationProblems(t *testing.T) {
	t.Run("writes validation error with details", func(t *testing.T) {
		rec := httptest.NewRecorder()
		details := []FieldError{
			{Field: "vin", Message: "required"},
			{Field: "make", Message: "required"},
		}

		respondWithValidationProblems(rec, "Missing fields", details)

		assert.Equal(t, 400, rec.Code)
		var resp map[string]interface{}
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
		errObj := resp["error"].(map[string]interface{})
		assert.Equal(t, "VALIDATION_ERROR", errObj["code"])
		assert.Equal(t, "Missing fields", errObj["message"])
		detailsList := errObj["details"].([]interface{})
		assert.Len(t, detailsList, 2)
	})
}

func TestRespondWithPayload(t *testing.T) {
	t.Run("writes success response", func(t *testing.T) {
		rec := httptest.NewRecorder()
		payload := map[string]interface{}{
			"success": true,
			"data":    "test data",
		}

		respondWithPayload(rec, 200, payload)

		assert.Equal(t, 200, rec.Code)
		assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

		var resp map[string]interface{}
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
		assert.True(t, resp["success"].(bool))
		assert.Equal(t, "test data", resp["data"])
	})

	t.Run("writes created response with 201", func(t *testing.T) {
		rec := httptest.NewRecorder()
		payload := map[string]interface{}{"success": true}

		respondWithPayload(rec, 201, payload)

		assert.Equal(t, 201, rec.Code)
	})
}
