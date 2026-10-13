package mailafrica

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestVerifyWebhookUsesRawBodyHMAC(t *testing.T) {
	client := New("https://api.mailafrica.online", "", "whsec_mail", "", 0)
	raw := []byte(`{"event":"inbound.message_received","message_id":9001}`)
	mac := hmac.New(sha256.New, []byte("whsec_mail")); _, _ = mac.Write(raw)
	signature := hex.EncodeToString(mac.Sum(nil))
	require.NoError(t, client.VerifyWebhook(raw, signature))
	require.ErrorIs(t, client.VerifyWebhook(raw, "bad"), ErrInvalidSignature)
}
