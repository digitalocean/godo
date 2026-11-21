package webhook

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var (
	// CurrentSignatureScheme is the current latest signature scheme.
	CurrentSignatureScheme = SignatureSchemeV1

	// AllSignatureSchemes is a list of all supported signature schemes.
	AllSignatureSchemes = []SignatureScheme{ /* populated by init() */ }

	// HTTPHeaderSignature is the name of the header that contains signature packages.
	HTTPHeaderSignature = "Do-Signature"
	// HTTPHeaderEventName is the name of the header that contains the event name.
	HTTPHeaderEventName = "Do-Event-Name"

	// DefaultTolerance is the default time tolerance for signature verification (3 minutes).
	DefaultTolerance time.Duration = 3 * 60 * time.Second

	// ErrExpiredSignature indicates that the signature timestamp is outside of the allowed tolerance.
	ErrExpiredSignature = fmt.Errorf("signature has expired")
	// ErrNoVerifiedSignature indicates that no verified signature was found.
	ErrNoVerifiedSignature = fmt.Errorf("no verified signature")
	// ErrNotSigned indicates that the payload is not signed.
	ErrNotSigned = fmt.Errorf("payload not signed")
)

var (
	signatureSchemesByVersion = map[int]SignatureScheme{}
)

func registerSignatureScheme(s SignatureScheme) {
	signatureSchemesByVersion[s.Version()] = s
	AllSignatureSchemes = append(AllSignatureSchemes, s)
}

func init() {
	// Schemes should be ordered by version number descending i.e. from newest to oldest.
	registerSignatureScheme(SignatureSchemeV1)
}

// SignaturePackage contains multiple signatures.
type SignaturePackage struct {
	Timestamp  time.Time
	Signatures []Signature
}

// NewSignaturePackage creates a signature package.
func NewSignaturePackage(t time.Time, payload []byte, secrets []string) SignaturePackage {
	p := SignaturePackage{
		Timestamp: t,
	}

	for _, scheme := range AllSignatureSchemes {
		for _, secret := range secrets {
			p.Signatures = append(p.Signatures, NewSignature(scheme, t, payload, secret))
		}
	}

	return p
}

// String returns the string representation of the signature package.
func (p *SignaturePackage) String() string {
	value := make([]string, 0, len(p.Signatures)+1)

	value = append(value, fmt.Sprintf("t=%d", p.Timestamp.Unix()))
	for _, s := range p.Signatures {
		value = append(value, s.String())
	}

	return strings.Join(value, ",")
}

// ParseSignaturePackage parses a signature package from its string representation.
func ParseSignaturePackage(value string) (SignaturePackage, error) {
	sigPack := SignaturePackage{}

	pairs := strings.Split(value, ",")
	for _, p := range pairs {
		parts := strings.SplitN(p, "=", 2)
		if len(parts) != 2 {
			return SignaturePackage{}, fmt.Errorf("invalid signature package")
		}

		k, v := parts[0], parts[1]
		if k == "t" {
			if !sigPack.Timestamp.IsZero() {
				return SignaturePackage{}, fmt.Errorf("timestamp cannot be specified multiple times")
			}
			ts, err := strconv.ParseInt(v, 10, 64)
			if err != nil {
				return SignaturePackage{}, fmt.Errorf("timestamp must be an integer")
			}
			sigPack.Timestamp = time.Unix(ts, 0).UTC()
		} else {
			sig, err := ParseSignature(p)
			if err != nil {
				return SignaturePackage{}, err
			}
			sigPack.Signatures = append(sigPack.Signatures, sig)
		}
	}

	if sigPack.Timestamp.IsZero() {
		return SignaturePackage{}, fmt.Errorf("missing timestamp")
	}

	return sigPack, nil
}

// VerificationOpts sets options for verifying signature packages.
type VerificationOpts struct {
	// Tolerance configures the maximum allowed signature age. Signatures older than this time window will fail verification.
	// If unset, defaults to DefaultTolerance.
	Tolerance time.Duration
	// IgnoreTolerance skips checking if the signature package timestamp is within the allowed tolerance.
	IgnoreTolerance bool
	// Now is an optional override of time.Now.
	Now func() time.Time
	// UntrustedSchemes is a list of signature schemes that are untrusted.
	UntrustedSchemes []SignatureScheme
}

// Verify verifies the given signature package. Verification passes if at least of the signatures in the package is verified.
func (p SignaturePackage) Verify(payload []byte, secret string, opts VerificationOpts) error {
	now := time.Now()
	if opts.Now != nil {
		now = opts.Now()
	}

	if !opts.IgnoreTolerance {
		tolerance := DefaultTolerance
		if opts.Tolerance > 0 {
			tolerance = opts.Tolerance
		}
		if now.Sub(p.Timestamp) > tolerance {
			return ErrExpiredSignature
		}
	}

	if len(p.Signatures) == 0 {
		return ErrNotSigned
	}

	// try to find at least one verified signature
verifySignatures:
	for _, s := range p.Signatures {
		for _, scheme := range opts.UntrustedSchemes {
			if scheme.Version() == s.Scheme.Version() {
				continue verifySignatures
			}
		}
		verified := s.Verify(payload, secret, p.Timestamp)
		if verified == nil {
			return nil
		}
	}

	return ErrNoVerifiedSignature
}

// SignHTTPRequest signs the given HTTP request and sets the signature header.
func SignHTTPRequest(r *http.Request, t time.Time, secrets []string) error {
	body, err := r.GetBody()
	if err != nil {
		return err
	}
	defer body.Close()

	payload, err := io.ReadAll(body)
	if err != nil {
		return err
	}

	sigPack := NewSignaturePackage(t, payload, secrets)
	r.Header.Set(HTTPHeaderSignature, sigPack.String())
	return nil
}

// VerifyHTTPRequest verifies an HTTP request.
func VerifyHTTPRequest(r *http.Request, secret string, opts VerificationOpts) error {
	header := r.Header.Get(HTTPHeaderSignature)
	if header == "" {
		return ErrNotSigned
	}

	sigPack, err := ParseSignaturePackage(header)
	if err != nil {
		return fmt.Errorf("parsing signature header: %w", err)
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return fmt.Errorf("reading request body: %w", err)
	}
	// Replace the body with a new reader after reading from the original
	r.Body = io.NopCloser(bytes.NewBuffer(body))

	return sigPack.Verify(body, secret, opts)
}

// NewSignature creates a new signature.
func NewSignature(scheme SignatureScheme, t time.Time, payload []byte, secret string) Signature {
	return Signature{
		Scheme: scheme,
		Value:  scheme.Sign(t, payload, secret),
	}
}

// Signature describes a signature.
type Signature struct {
	Scheme SignatureScheme
	Value  string
}

// String returns the string representation of a signature.
func (s Signature) String() string {
	return fmt.Sprintf("v%d=%s", s.Scheme.Version(), s.Value)
}

// Equal compares two signatures for equality without leaking timing information.
func (s Signature) Equal(o Signature) bool {
	return subtle.ConstantTimeCompare([]byte(s.Value), []byte(o.Value)) == 1
}

// Verify verifies the given signature. The timestamp that was used to generate this signature must be provided.
func (s Signature) Verify(payload []byte, secret string, t time.Time) error {
	if s.Scheme == nil {
		return fmt.Errorf("invalid signature scheme")
	}

	freshSig := NewSignature(s.Scheme, t, payload, secret)
	if !s.Equal(freshSig) {
		return ErrNoVerifiedSignature
	}

	// the signatures are identical
	return nil
}

// ParseSignature attempts to parse a signature from its string representation.
func ParseSignature(value string) (Signature, error) {
	parts := strings.SplitN(value, "=", 2)
	if len(parts) != 2 {
		return Signature{}, fmt.Errorf("invalid signature format")
	}

	versionStr, value := parts[0], parts[1]
	if !strings.HasPrefix(versionStr, "v") {
		return Signature{}, fmt.Errorf("invalid signature format")
	}
	version, err := strconv.ParseInt(versionStr[1:], 10, 0)
	if err != nil {
		return Signature{}, fmt.Errorf("signature scheme version must be an integer")
	}
	scheme := signatureSchemesByVersion[int(version)]
	if scheme == nil {
		return Signature{}, fmt.Errorf("invalid signature scheme version %d", version)
	}

	return Signature{
		Scheme: scheme,
		Value:  value,
	}, nil
}

// SignatureScheme describes a signature scheme.
type SignatureScheme interface {
	Sign(t time.Time, payload []byte, secret string) string
	Version() int
}

// SignatureSchemeV1 computes an HMAC-SHA256 signature of the timestamp and payload in the following format:
//
//	{unix timestamp}.{payload}
//
// The resulting signature is then hex-encoded.
var SignatureSchemeV1 SignatureScheme = &signatureSchemeV1{}

type signatureSchemeV1 struct{}

// Version returns the scheme version.
func (s *signatureSchemeV1) Version() int {
	return 1
}

// Sign signs a payload.
func (s *signatureSchemeV1) Sign(t time.Time, payload []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(fmt.Sprintf("%d", t.Unix())))
	mac.Write([]byte("."))
	mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}

// EventName returns a namespaced event name.
func EventName(ns string, name string) string {
	return fmt.Sprintf("%s.%s", ns, name)
}
