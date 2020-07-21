package api

import (
	"bytes"
	"context"
	"testing"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/encoding"
	"github.com/keys-pub/keys/tsutil"
	"github.com/stretchr/testify/require"
)

func TestAuth(t *testing.T) {
	alice := keys.NewEdX25519KeyFromSeed(keys.Bytes32(bytes.Repeat([]byte{0x01}, 32)))

	clock := tsutil.NewTestClock()

	tm := clock.Now()
	nonce := keys.Bytes32(bytes.Repeat([]byte{0x01}, 32))
	urs := "https://keys.pub/vault/kex132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqqph077?idx=123"
	auth, err := newAuth("GET", urs, "", tm, nonce, alice)
	require.NoError(t, err)
	require.Equal(t, "kex132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqqph077:K0KnYYnx+VnhpRS0lBJVfwSaYa3zweapGtc87Uh4h1pfv/VeVMaS/YRD/d+Y+U3ANFMkR+OFGRYniWirFK3sBg==", auth.Header())
	require.Equal(t, "https://keys.pub/vault/kex132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqqph077?idx=123&nonce=0El6XFXwsUFD8J2vGxsaboW7rZYnQRBP5d9erwRwd29&ts=1234567890001", auth.URL.String())

	req, err := newRequest(context.TODO(), "GET", urs, nil, "", tm, nonce, alice)
	require.NoError(t, err)
	require.Equal(t, "https://keys.pub/vault/kex132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqqph077?idx=123&nonce=0El6XFXwsUFD8J2vGxsaboW7rZYnQRBP5d9erwRwd29&ts=1234567890001", req.URL.String())
	require.Equal(t, "kex132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqqph077:K0KnYYnx+VnhpRS0lBJVfwSaYa3zweapGtc87Uh4h1pfv/VeVMaS/YRD/d+Y+U3ANFMkR+OFGRYniWirFK3sBg==", req.Header.Get("Authorization"))

	rds := NewRedisTest(tsutil.NewTestClock())
	_, err = CheckAuthorization(context.TODO(),
		"GET",
		"https://keys.pub/vault/kex132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqqph077?idx=123&nonce=0El6XFXwsUFD8J2vGxsaboW7rZYnQRBP5d9erwRwd29&ts=1234567890001",
		"kex132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqqph077",
		"kex132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqqph077:K0KnYYnx+VnhpRS0lBJVfwSaYa3zweapGtc87Uh4h1pfv/VeVMaS/YRD/d+Y+U3ANFMkR+OFGRYniWirFK3sBg==",
		"",
		rds, clock.Now())
	require.NoError(t, err)

	// Change method
	rds = NewRedisTest(tsutil.NewTestClock())
	_, err = CheckAuthorization(context.TODO(),
		"HEAD",
		"https://keys.pub/vault/kex132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqqph077?idx=123&nonce=0El6XFXwsUFD8J2vGxsaboW7rZYnQRBP5d9erwRwd29&ts=1234567890001",
		"kex132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqqph077",
		"kex132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqqph077:K0KnYYnx+VnhpRS0lBJVfwSaYa3zweapGtc87Uh4h1pfv/VeVMaS/YRD/d+Y+U3ANFMkR+OFGRYniWirFK3sBg==",
		"",
		rds, clock.Now())
	require.EqualError(t, err, "verify failed")

	// Re-order url params
	rds = NewRedisTest(tsutil.NewTestClock())
	_, err = CheckAuthorization(context.TODO(),
		"GET",
		"https://keys.pub/vault/kex132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqqph077?nonce=0El6XFXwsUFD8J2vGxsaboW7rZYnQRBP5d9erwRwd29&ts=1234567890001&idx=123",
		"kex132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqqph077",
		"kex132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqqph077:K0KnYYnx+VnhpRS0lBJVfwSaYa3zweapGtc87Uh4h1pfv/VeVMaS/YRD/d+Y+U3ANFMkR+OFGRYniWirFK3sBg==",
		"",
		rds, clock.Now())
	require.EqualError(t, err, "verify failed")

	// Different kid
	rds = NewRedisTest(tsutil.NewTestClock())
	_, err = CheckAuthorization(context.TODO(),
		"GET",
		"https://keys.pub/vault/kex132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqqph077?idx=123&nonce=0El6XFXwsUFD8J2vGxsaboW7rZYnQRBP5d9erwRwd29&ts=1234567890001",
		"kex16jvh9cc6na54xwpjs3ztlxdsj6q3scl65lwxxj72m6cadewm404qts0jw9",
		"kex132yw8ht5p8cetl2jmvknewjawt9xwzdlrk2pyxlnwjyqrdq0dawqqph077:K0KnYYnx+VnhpRS0lBJVfwSaYa3zweapGtc87Uh4h1pfv/VeVMaS/YRD/d+Y+U3ANFMkR+OFGRYniWirFK3sBg==",
		"",
		rds, clock.Now())
	require.EqualError(t, err, "invalid kid")
}

func TestNewRequest(t *testing.T) {
	key := keys.GenerateEdX25519Key()
	clock := tsutil.NewTestClock()
	rds := NewRedisTest(tsutil.NewTestClock())

	// GET
	req, err := NewRequest("GET", "https://keys.pub/test", nil, "", clock.Now(), key)
	require.NoError(t, err)
	check, err := CheckAuthorization(context.TODO(),
		"GET",
		req.URL.String(),
		key.ID(),
		req.Header["Authorization"][0],
		"",
		rds, clock.Now())
	require.NoError(t, err)
	require.Equal(t, key.ID(), check.KID)

	// POST
	body := []byte(`{\"test\": 1}`)
	contentHash := encoding.EncodeBase64(keys.SHA256([]byte(body)))
	req, err = NewRequest("POST", "https://keys.pub/test", bytes.NewReader(body), contentHash, clock.Now(), key)
	require.NoError(t, err)
	check, err = CheckAuthorization(context.TODO(),
		"POST",
		req.URL.String(),
		key.ID(),
		req.Header["Authorization"][0],
		contentHash,
		rds, clock.Now())
	require.NoError(t, err)
	require.Equal(t, key.ID(), check.KID)

	// POST (invalid content hash)
	contentHash = encoding.EncodeBase64(keys.SHA256([]byte("invalid")))
	req, err = NewRequest("POST", "https://keys.pub/test", bytes.NewReader([]byte(body)), contentHash, clock.Now(), key)
	require.NoError(t, err)
	check, err = CheckAuthorization(context.TODO(),
		"POST",
		req.URL.String(),
		key.ID(),
		req.Header["Authorization"][0],
		contentHash,
		rds, clock.Now())
	require.EqualError(t, err, "verify failed")
}
