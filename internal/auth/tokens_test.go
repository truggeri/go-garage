package auth

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildTokenManager(t *testing.T) {
	t.Run("creates manager with valid secret", func(t *testing.T) {
		mgr, err := BuildTokenManager("test-secret-key-12345", StandardTokenDurations())
		require.NoError(t, err)
		assert.NotNil(t, mgr)
	})

	t.Run("returns error for empty secret", func(t *testing.T) {
		mgr, err := BuildTokenManager("", StandardTokenDurations())
		assert.Error(t, err)
		assert.Nil(t, mgr)
	})
}

func TestTokenManager_GenerateTokenBundle(t *testing.T) {
	mgr, err := BuildTokenManager("test-secret-key-12345", StandardTokenDurations())
	require.NoError(t, err)

	payload := TokenPayload{
		AccountID:   "user-abc-123",
		AccountName: "johndoe",
	}

	t.Run("generates valid token bundle", func(t *testing.T) {
		bundle, err := mgr.GenerateTokenBundle(payload)
		require.NoError(t, err)

		assert.NotEmpty(t, bundle.AccessToken)
		assert.NotEmpty(t, bundle.RefreshToken)
		assert.NotEqual(t, bundle.AccessToken, bundle.RefreshToken)
		assert.True(t, bundle.AccessExpiresAt.After(time.Now()))
		assert.True(t, bundle.RefreshExpiresAt.After(time.Now()))
		assert.True(t, bundle.RefreshExpiresAt.After(bundle.AccessExpiresAt))
	})
}

func TestTokenManager_ValidateToken(t *testing.T) {
	mgr, err := BuildTokenManager("test-secret-key-12345", StandardTokenDurations())
	require.NoError(t, err)

	payload := TokenPayload{
		AccountID:   "user-abc-123",
		AccountName: "johndoe",
	}

	t.Run("validates access token successfully", func(t *testing.T) {
		bundle, err := mgr.GenerateTokenBundle(payload)
		require.NoError(t, err)

		verified, err := mgr.ValidateToken(bundle.AccessToken)
		require.NoError(t, err)

		assert.Equal(t, payload.AccountID, verified.AccountID)
		assert.Equal(t, payload.AccountName, verified.AccountName)
		assert.Equal(t, AccessTokenKind, verified.TokenKind)
	})

	t.Run("validates refresh token successfully", func(t *testing.T) {
		bundle, err := mgr.GenerateTokenBundle(payload)
		require.NoError(t, err)

		verified, err := mgr.ValidateToken(bundle.RefreshToken)
		require.NoError(t, err)

		assert.Equal(t, payload.AccountID, verified.AccountID)
		assert.Equal(t, payload.AccountName, verified.AccountName)
		assert.Equal(t, RefreshTokenKind, verified.TokenKind)
	})

	t.Run("rejects invalid token", func(t *testing.T) {
		_, err := mgr.ValidateToken("invalid.token.string")
		assert.Error(t, err)
	})

	t.Run("rejects token signed with different key", func(t *testing.T) {
		differentMgr, _ := BuildTokenManager("different-secret-key", StandardTokenDurations())
		bundle, _ := differentMgr.GenerateTokenBundle(payload)

		_, err := mgr.ValidateToken(bundle.AccessToken)
		assert.Error(t, err)
	})

	t.Run("rejects expired token", func(t *testing.T) {
		shortDurations := TokenDurations{
			AccessValidity:  -1 * time.Hour,
			RefreshValidity: -1 * time.Hour,
		}
		shortMgr, _ := BuildTokenManager("test-secret-key-12345", shortDurations)
		bundle, _ := shortMgr.GenerateTokenBundle(payload)

		_, err := mgr.ValidateToken(bundle.AccessToken)
		assert.Error(t, err)
	})
}

func TestTokenManager_RefreshAccessToken(t *testing.T) {
	mgr, err := BuildTokenManager("test-secret-key-12345", StandardTokenDurations())
	require.NoError(t, err)

	payload := TokenPayload{
		AccountID:   "user-abc-123",
		AccountName: "johndoe",
	}

	t.Run("generates new bundle from refresh token", func(t *testing.T) {
		originalBundle, err := mgr.GenerateTokenBundle(payload)
		require.NoError(t, err)

		// Wait briefly to ensure different iat timestamp
		time.Sleep(1100 * time.Millisecond)

		newBundle, err := mgr.RefreshAccessToken(originalBundle.RefreshToken)
		require.NoError(t, err)

		assert.NotEmpty(t, newBundle.AccessToken)
		assert.NotEqual(t, originalBundle.AccessToken, newBundle.AccessToken)

		verified, err := mgr.ValidateToken(newBundle.AccessToken)
		require.NoError(t, err)
		assert.Equal(t, payload.AccountID, verified.AccountID)
	})

	t.Run("rejects access token for refresh", func(t *testing.T) {
		bundle, err := mgr.GenerateTokenBundle(payload)
		require.NoError(t, err)

		_, err = mgr.RefreshAccessToken(bundle.AccessToken)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not a refresh token")
	})

	t.Run("rejects invalid token", func(t *testing.T) {
		_, err := mgr.RefreshAccessToken("invalid.token.here")
		assert.Error(t, err)
	})
}

func TestStandardTokenDurations(t *testing.T) {
	durations := StandardTokenDurations()

	assert.Equal(t, 15*time.Minute, durations.AccessValidity)
	assert.Equal(t, 7*24*time.Hour, durations.RefreshValidity)
}
