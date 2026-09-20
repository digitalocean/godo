package webhook

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func init() {
	// add a fake signature scheme to test functionality related to multiple schemes
	registerSignatureScheme(FakeSignatureScheme)
}

var (
	// find the easter egg 🥚
	testPayload = []byte("it is wednesday my dudes 🕷️")
	testSecret  = "du-TY1GUFGk"
)

func TestSignatureSchemeV1(t *testing.T) {
	ts := time.Date(2000, 1, 1, 10, 0, 0, 0, time.UTC)
	sig := NewSignature(SignatureSchemeV1, ts, testPayload, testSecret)
	require.Equal(t, "v1=b70100cf2943bec15996e3ae9392d0dcaf21f285fa81969108185d47b292dfa2", sig.String())
	require.Equal(t, SignatureSchemeV1.Sign(ts, testPayload, testSecret), sig.Value)

	sig = NewSignature(SignatureSchemeV1, ts, testPayload, "other-secret")
	require.Equal(t, "v1=817555f45dd54e36c87ad1a349083e5d2e706cae2eae7f4077379f5444f5b985", sig.String())

	sig = NewSignature(SignatureSchemeV1, ts.Add(time.Minute), testPayload, "other-secret")
	require.Equal(t, "v1=eb5eb314fd727bcbae0713640b66540c64dc7e7cfa18f715deee29ed5db59347", sig.String())
}

func TestSignature(t *testing.T) {
	ts := time.Date(2000, 1, 1, 10, 0, 0, 0, time.UTC)
	var sig Signature
	t.Run("new", func(t *testing.T) {
		sig = NewSignature(SignatureSchemeV1, ts, testPayload, testSecret)
		require.Equal(t, Signature{
			Scheme: SignatureSchemeV1,
			Value:  SignatureSchemeV1.Sign(ts, testPayload, testSecret),
		}, sig)
		require.Equal(t, "v1=b70100cf2943bec15996e3ae9392d0dcaf21f285fa81969108185d47b292dfa2", sig.String())
	})

	t.Run("verify", func(t *testing.T) {
		err := sig.Verify(testPayload, testSecret, ts)
		require.NoError(t, err)

		err = sig.Verify(testPayload, "other-secret", ts)
		require.Equal(t, ErrNoVerifiedSignature, err)

		err = sig.Verify(testPayload, testSecret, ts.Add(time.Hour))
		require.Equal(t, ErrNoVerifiedSignature, err)

		err = sig.Verify([]byte("other-payload"), testSecret, ts)
		require.Equal(t, ErrNoVerifiedSignature, err)

		err = Signature{}.Verify(nil, "", ts)
		require.EqualError(t, err, "invalid signature scheme")
	})

	t.Run("parse", func(t *testing.T) {
		parsed, err := ParseSignature(sig.String())
		require.NoError(t, err)
		require.Equal(t, Signature{
			Scheme: sig.Scheme,
			Value:  sig.Value,
		}, parsed)

		require.True(t, sig.Equal(parsed))
		differentSig := NewSignature(sig.Scheme, ts, testPayload, "other-secret")
		require.False(t, sig.Equal(differentSig))

		for val, err := range map[string]string{
			"🐌":   "invalid signature format",
			"🐌=":  "invalid signature format",
			"=a":  "invalid signature format",
			"v🐌=": "signature scheme version must be an integer",
			"v=a": "signature scheme version must be an integer",
			"v0=": "invalid signature scheme version 0",
		} {
			_, got := ParseSignature(val)
			require.EqualError(t, got, err, val)
		}
	})
}

func TestSignaturePackage(t *testing.T) {
	ts := time.Date(2000, 1, 1, 10, 0, 0, 0, time.UTC)
	secrets := []string{testSecret, "some-secret"}
	var sp SignaturePackage
	t.Run("new", func(t *testing.T) {
		sp = NewSignaturePackage(ts, testPayload, secrets)
		require.Equal(
			t,
			`t=946720800,`+
				`v1=b70100cf2943bec15996e3ae9392d0dcaf21f285fa81969108185d47b292dfa2,v1=b3218d58417e81cf347b439091b9ede800b2e1555f90fee81ac94f67c249da26,`+
				`v1337=946720800:du-TY1GUFGk:(32),v1337=946720800:some-secret:(32)`,
			sp.String(),
		)
	})

	t.Run("verify", func(t *testing.T) {
		nowWithinTolerance := func() time.Time {
			return ts.Add(DefaultTolerance)
		}

		// verified: happy path
		err := sp.Verify(testPayload, testSecret, VerificationOpts{Now: nowWithinTolerance})
		require.NoError(t, err)

		// false: expired signature
		err = sp.Verify(testPayload, testSecret, VerificationOpts{
			Now: func() time.Time {
				return ts.Add(DefaultTolerance + time.Second)
			},
		})
		require.Equal(t, ErrExpiredSignature, err)

		// false: expired signature w/ custom tolerance
		err = sp.Verify(testPayload, testSecret, VerificationOpts{
			Tolerance: 3 * time.Second,
			Now: func() time.Time {
				return ts.Add(5 * time.Second)
			},
		})
		require.Equal(t, ErrExpiredSignature, err)

		// verified: expired signature w/ ignore tolerance
		err = sp.Verify(testPayload, testSecret, VerificationOpts{
			IgnoreTolerance: true,
			Now: func() time.Time {
				return ts.Add(DefaultTolerance + time.Second)
			},
		})
		require.NoError(t, err)

		// false: signature signed by unknown secret
		err = sp.Verify(testPayload, "other-secret", VerificationOpts{Now: nowWithinTolerance})
		require.Equal(t, ErrNoVerifiedSignature, err)

		// false: signature does not match payload
		err = sp.Verify([]byte("other-payload"), testSecret, VerificationOpts{Now: nowWithinTolerance})
		require.Equal(t, ErrNoVerifiedSignature, err)

		// false: signature signed by untrusted scheme
		err = sp.Verify(testPayload, testSecret, VerificationOpts{
			Now:              nowWithinTolerance,
			UntrustedSchemes: AllSignatureSchemes,
		})
		require.Equal(t, ErrNoVerifiedSignature, err)

		// verified: only one of the schemes is untrusted
		err = sp.Verify(testPayload, testSecret, VerificationOpts{
			Now:              nowWithinTolerance,
			UntrustedSchemes: []SignatureScheme{FakeSignatureScheme},
		})
		require.NoError(t, err)
	})

	t.Run("parse", func(t *testing.T) {
		// sanity checks to double check the correctness of the hardcoded test string above
		require.Equal(t, ts.Unix(), int64(946720800))

		v1s1, err := ParseSignature("v1=b70100cf2943bec15996e3ae9392d0dcaf21f285fa81969108185d47b292dfa2")
		require.NoError(t, err)
		require.Equal(t, NewSignature(SignatureSchemeV1, ts, testPayload, secrets[0]), v1s1)
		v1s2, err := ParseSignature("v1=b3218d58417e81cf347b439091b9ede800b2e1555f90fee81ac94f67c249da26")
		require.NoError(t, err)
		require.Equal(t, NewSignature(SignatureSchemeV1, ts, testPayload, secrets[1]), v1s2)

		v1337s1, err := ParseSignature("v1337=946720800:du-TY1GUFGk:(32)")
		require.NoError(t, err)
		require.Equal(t, NewSignature(FakeSignatureScheme, ts, testPayload, secrets[0]), v1337s1)
		v1337s2, err := ParseSignature("v1337=946720800:some-secret:(32)")
		require.NoError(t, err)
		require.Equal(t, NewSignature(FakeSignatureScheme, ts, testPayload, secrets[1]), v1337s2)

		parsed, err := ParseSignaturePackage(sp.String())
		require.NoError(t, err)
		require.Equal(t, sp, parsed)
		require.Equal(t, sp.String(), parsed.String())

		// error cases
		for val, err := range map[string]string{
			"🐌":      "invalid signature package",
			"v999=🐌": "invalid signature scheme version 999",
			"v1=b70100cf2943bec15996e3ae9392d0dcaf21f285fa81969108185d47b292dfa2": "missing timestamp",
			"t=🐌": "timestamp must be an integer",
			"t=123,v1=b70100cf2943bec15996e3ae9392d0dcaf21f285fa81969108185d47b292dfa2,t=341": "timestamp cannot be specified multiple times",
		} {
			_, got := ParseSignaturePackage(val)
			require.EqualError(t, got, err, val)
		}
	})
}

var FakeSignatureScheme = &fakeSignatureScheme{}

type fakeSignatureScheme struct{}

func (s *fakeSignatureScheme) Version() int {
	return 1337
}

func (s *fakeSignatureScheme) Sign(t time.Time, payload []byte, secret string) string {
	return strings.ReplaceAll(fmt.Sprintf("%d:%s:(%d)", t.Unix(), secret, len(payload)), ",", "")
}

func TestHTTPRequests(t *testing.T) {
	now := time.Date(2000, 1, 1, 10, 0, 0, 0, time.UTC)
	nowFunc := func() time.Time { return now }

	tcs := []struct {
		name    string
		handler http.HandlerFunc
		request func(t *testing.T, url string) *http.Request
	}{
		{
			name: "happy path",
			request: func(t *testing.T, url string) *http.Request {
				req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(testPayload))
				require.NoError(t, err)
				// sign request
				err = SignHTTPRequest(req, now, []string{testSecret, "other-secret"})
				require.NoError(t, err)
				return req
			},
			handler: func(w http.ResponseWriter, r *http.Request) {
				err := VerifyHTTPRequest(r, testSecret, VerificationOpts{Now: nowFunc})
				require.NoError(t, err)

				err = VerifyHTTPRequest(r, "other-secret", VerificationOpts{Now: nowFunc})
				require.NoError(t, err)

				err = VerifyHTTPRequest(r, testSecret, VerificationOpts{
					Now:              nowFunc,
					UntrustedSchemes: AllSignatureSchemes,
				})
				require.Equal(t, ErrNoVerifiedSignature, err)

				headerValue := r.Header.Get(HTTPHeaderSignature)
				require.Equal(t, "t=946720800,v1=b70100cf2943bec15996e3ae9392d0dcaf21f285fa81969108185d47b292dfa2,v1=817555f45dd54e36c87ad1a349083e5d2e706cae2eae7f4077379f5444f5b985,v1337=946720800:du-TY1GUFGk:(32),v1337=946720800:other-secret:(32)", headerValue)

				// sanity check & ensure that r.Body is still accessible
				body, err := io.ReadAll(r.Body)
				require.NoError(t, err)
				sig := NewSignature(SignatureSchemeV1, time.Unix(946720800, 0), body, testSecret)
				require.Equal(t, Signature{
					Scheme: SignatureSchemeV1,
					Value:  "b70100cf2943bec15996e3ae9392d0dcaf21f285fa81969108185d47b292dfa2",
				}, sig)
			},
		},
		{
			name: "unsigned request",
			request: func(t *testing.T, url string) *http.Request {
				req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(testPayload))
				require.NoError(t, err)
				return req
			},
			handler: func(w http.ResponseWriter, r *http.Request) {
				err := VerifyHTTPRequest(r, testSecret, VerificationOpts{Now: nowFunc})
				require.Equal(t, ErrNotSigned, err)
			},
		},
		{
			name: "header present without signatures",
			request: func(t *testing.T, url string) *http.Request {
				req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(testPayload))
				require.NoError(t, err)
				req.Header.Set(HTTPHeaderSignature, fmt.Sprintf("t=%d", now.Unix()))
				return req
			},
			handler: func(w http.ResponseWriter, r *http.Request) {
				err := VerifyHTTPRequest(r, testSecret, VerificationOpts{Now: nowFunc})
				require.Equal(t, ErrNotSigned, err)
			},
		},
		{
			name: "bad header value",
			request: func(t *testing.T, url string) *http.Request {
				req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(testPayload))
				require.NoError(t, err)
				req.Header.Set(HTTPHeaderSignature, "🧘🏻‍♂️🌍🥖🚗📱🎉✅")
				return req
			},
			handler: func(w http.ResponseWriter, r *http.Request) {
				err := VerifyHTTPRequest(r, testSecret, VerificationOpts{Now: nowFunc})
				require.EqualError(t, err, "parsing signature header: invalid signature package")
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				tc.handler(w, r)
				w.WriteHeader(http.StatusUnavailableForLegalReasons)
			}))

			req := tc.request(t, server.URL)
			res, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer res.Body.Close()
			require.Equal(t, http.StatusUnavailableForLegalReasons, res.StatusCode)
		})
	}
}
