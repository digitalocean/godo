package modal

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"sort"
)

func hashID(prefix string, parts ...string) string {
	h := sha256.New()
	for _, p := range parts {
		h.Write([]byte(p))
		h.Write([]byte{0})
	}
	return prefix + "-" + hex.EncodeToString(h.Sum(nil))[:22]
}

// appService implements AppService. A Modal App is a name; DigitalOcean has no
// Apps, so FromName only builds the handle.
type appService struct {
	notImplementedAppService
	c *Client
}

// FromName mirrors AppService.FromName.
func (s appService) FromName(ctx context.Context, name string, params *AppFromNameParams) (*App, error) {
	env := ""
	if params != nil {
		env = params.Environment
	}
	return &App{AppID: hashID("ap", env, name), Name: name, Environment: env, Sandboxes: &AppSandboxesManager{}}, nil
}

// App mirrors modal.App.
type App struct {
	AppID       string
	Name        string
	Environment string

	// Sandboxes mirrors App.Sandboxes. Its methods are not implemented.
	Sandboxes *AppSandboxesManager
}

// Secret mirrors modal.Secret. Only secrets made with Secrets.FromMap exist.
type Secret struct {
	SecretID string
	Name     string

	env map[string]string
}

// secretService implements SecretService. FromName and Delete need a
// DigitalOcean secrets store and are not implemented.
type secretService struct {
	notImplementedSecretService
}

// FromMap mirrors SecretService.FromMap. The secret is kept in memory and
// injected into sandboxes as environment variables.
func (secretService) FromMap(ctx context.Context, keyValuePairs map[string]string, params *SecretFromMapParams) (*Secret, error) {
	keys := make([]string, 0, len(keyValuePairs))
	for k := range keyValuePairs {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, 2*len(keys))
	for _, k := range keys {
		parts = append(parts, k, keyValuePairs[k])
	}
	env := make(map[string]string, len(keyValuePairs))
	for k, v := range keyValuePairs {
		env[k] = v
	}
	return &Secret{SecretID: hashID("st", parts...), env: env}, nil
}

func secretEnv(secrets []*Secret) (map[string]string, error) {
	out := map[string]string{}
	for _, s := range secrets {
		if s == nil {
			continue
		}
		if s.env == nil {
			return nil, notImplemented("Secret", "a named secrets store API: only secrets from Secrets.FromMap can be used")
		}
		for k, v := range s.env {
			out[k] = v
		}
	}
	return out, nil
}

func joinEnv(layers ...map[string]string) map[string]string {
	out := map[string]string{}
	for _, l := range layers {
		for k, v := range l {
			out[k] = v
		}
	}
	return out
}
