package godo

import (
	"net/http"
	"net/url"
	"testing"
)

func TestCheckRequestPath(t *testing.T) {
	valid := []string{
		"",
		"/",
		"/v2/databases/db-123",
		"v2/account/keys/aa:bb:cc",
		"/v2/registry/reg/repositories/foo%2Fbar/tags",
		"/v2/databases/db/topics/my.topic",
		"/v2/databases/db/users/..user",
		"https://api.digitalocean.com/v2/droplets?page=2",
	}
	for _, p := range valid {
		u, err := url.Parse(p)
		if err != nil {
			t.Fatalf("url.Parse(%q): %v", p, err)
		}
		if err := checkRequestPath(u); err != nil {
			t.Errorf("checkRequestPath(%q) = %v, want nil", p, err)
		}
	}

	invalid := []string{
		"/v2/databases/../../v2/account/keys/999",
		"/v2/databases/./x",
		"/v2/databases/..",
		"/v2/databases/%2e%2e/x",
		"/v2/databases/%2E%2e",
		"/v2/databases/..%2F..%2Fv2%2Faccount%2Fkeys%2F999",
		"/v2/databases/x%2F..",
		"/v2/databases/..%5Cx",
		"/v2/databases/",
		"/v2/databases//users",
		"/v2/databases/bad%0Aid",
		"https://api.digitalocean.com/v2/../v2/account/keys/1",
	}
	for _, p := range invalid {
		u, err := url.Parse(p)
		if err != nil {
			t.Fatalf("url.Parse(%q): %v", p, err)
		}
		if err := checkRequestPath(u); err == nil {
			t.Errorf("checkRequestPath(%q) = nil, want error", p)
		}
	}
}

func TestNewRequest_rejectsTraversal(t *testing.T) {
	c := NewClient(nil)
	for _, p := range []string{
		"/v2/databases/../../v2/account/keys/999",
		"/v2/databases/..%2F..%2Fv2%2Faccount%2Fkeys%2F999",
	} {
		if _, err := c.NewRequest(ctx, http.MethodDelete, p, nil); err == nil {
			t.Errorf("NewRequest(%q) error = nil, want error", p)
		}
	}
}

// failOnRequest registers a catch-all handler: any request reaching the server
// means the path was not rejected client-side.
func failOnRequest(t *testing.T) {
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("unexpected request %s %s", r.Method, r.URL.EscapedPath())
	})
}

func TestPathTraversal_rejectedBeforeSending(t *testing.T) {
	setup()
	defer teardown()
	failOnRequest(t)

	ids := []string{
		"../../v2/account/keys/999",
		"..",
		".",
		"",
	}
	for _, id := range ids {
		if _, err := client.Databases.Delete(ctx, id); err == nil {
			t.Errorf("Databases.Delete(%q) error = nil, want error", id)
		}
		if _, err := client.Databases.DeleteUser(ctx, "db", id); err == nil {
			t.Errorf("Databases.DeleteUser(%q) error = nil, want error", id)
		}
		if _, err := client.Firewalls.Delete(ctx, id); err == nil {
			t.Errorf("Firewalls.Delete(%q) error = nil, want error", id)
		}
		if _, _, err := client.Projects.Get(ctx, id); err == nil {
			t.Errorf("Projects.Get(%q) error = nil, want error", id)
		}
		if _, err := client.UptimeChecks.Delete(ctx, id); err == nil {
			t.Errorf("UptimeChecks.Delete(%q) error = nil, want error", id)
		}
		if _, err := client.VPCs.Delete(ctx, id); err == nil {
			t.Errorf("VPCs.Delete(%q) error = nil, want error", id)
		}
		if _, err := client.Keys.DeleteByFingerprint(ctx, id); err == nil {
			t.Errorf("Keys.DeleteByFingerprint(%q) error = nil, want error", id)
		}
	}
}

func TestPathSegments_escaped(t *testing.T) {
	setup()
	defer teardown()

	var got string
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		got = r.URL.EscapedPath()
		w.WriteHeader(http.StatusNoContent)
	})

	tests := []struct {
		name string
		do   func() error
		want string
	}{
		{
			name: "database id with encoded traversal",
			do: func() error {
				_, err := client.Databases.Delete(ctx, "..%2F..%2Fv2%2Faccount%2Fkeys%2F999")
				return err
			},
			want: "/v2/databases/..%252F..%252Fv2%252Faccount%252Fkeys%252F999",
		},
		{
			name: "database id with slash",
			do: func() error {
				_, err := client.Databases.Delete(ctx, "db/users/admin")
				return err
			},
			want: "/v2/databases/db%2Fusers%2Fadmin",
		},
		{
			name: "database user with query and fragment",
			do: func() error {
				_, err := client.Databases.DeleteUser(ctx, "db", "user?x=1#y")
				return err
			},
			want: "/v2/databases/db/users/user%3Fx=1%23y",
		},
		{
			name: "firewall id with slash",
			do: func() error {
				_, err := client.Firewalls.Delete(ctx, "fw/rules")
				return err
			},
			want: "/v2/firewalls/fw%2Frules",
		},
		{
			name: "vpc id with slash",
			do: func() error {
				_, err := client.VPCs.Delete(ctx, "vpc/members")
				return err
			},
			want: "/v2/vpcs/vpc%2Fmembers",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got = ""
			if err := tt.do(); err != nil {
				t.Fatalf("request error: %v", err)
			}
			if got != tt.want {
				t.Errorf("escaped path = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestQueryValues_escaped(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/droplets", func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.URL.Query().Get("name"), "web&tag_name=prod"; got != want {
			t.Errorf("name = %q, want %q", got, want)
		}
		if got := r.URL.Query().Get("tag_name"); got != "" {
			t.Errorf("tag_name = %q, want empty", got)
		}
		w.Write([]byte(`{"droplets":[]}`))
	})

	if _, _, err := client.Droplets.ListByName(ctx, "web&tag_name=prod", nil); err != nil {
		t.Fatalf("ListByName: %v", err)
	}
}
