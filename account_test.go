package godo

import (
	"fmt"
	"net/http"
	"reflect"
	"testing"
)

func TestAccountGet(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/account", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)

		response := `
		{ "account": {
			"droplet_limit": 25,
			"floating_ip_limit": 25,
			"reserved_ip_limit": 25,
			"volume_limit": 22,
			"email": "sammy@digitalocean.com",
			"uuid": "b6fr89dbf6d9156cace5f3c78dc9851d957381ef",
			"email_verified": true
			}
		}`

		fmt.Fprint(w, response)
	})

	acct, _, err := client.Account.Get(ctx)
	if err != nil {
		t.Errorf("Account.Get returned error: %v", err)
	}

	expected := &Account{DropletLimit: 25, FloatingIPLimit: 25, ReservedIPLimit: 25, Email: "sammy@digitalocean.com",
		UUID: "b6fr89dbf6d9156cace5f3c78dc9851d957381ef", EmailVerified: true, VolumeLimit: 22}
	if !reflect.DeepEqual(acct, expected) {
		t.Errorf("Account.Get returned %+v, expected %+v", acct, expected)
	}
}

func TestAccountString(t *testing.T) {
	acct := &Account{
		DropletLimit:    25,
		FloatingIPLimit: 25,
		ReservedIPLimit: 25,
		VolumeLimit:     22,
		Email:           "sammy@digitalocean.com",
		UUID:            "b6fr89dbf6d9156cace5f3c78dc9851d957381ef",
		EmailVerified:   true,
		Status:          "active",
		StatusMessage:   "message",
	}

	stringified := acct.String()
	expected := `godo.Account{DropletLimit:25, FloatingIPLimit:25, ReservedIPLimit:25, VolumeLimit:22, Email:"sammy@digitalocean.com", UUID:"b6fr89dbf6d9156cace5f3c78dc9851d957381ef", EmailVerified:true, Status:"active", StatusMessage:"message"}`
	if expected != stringified {
		t.Errorf("\n     got %+v\nexpected %+v", stringified, expected)
	}

}

func TestAccountGetWithTeam(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/account", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)

		response := `
		{ "account": {
			"droplet_limit": 25,
			"floating_ip_limit": 25,
			"volume_limit": 22,
			"email": "sammy@digitalocean.com",
			"uuid": "b6fr89dbf6d9156cace5f3c78dc9851d957381ef",
			"email_verified": true,
			"team": {
				"name": "My Team",
				"uuid": "b6fr89dbf6d9156cace5f3c78dc9851d957381ef"
			}
			}
		}`

		fmt.Fprint(w, response)
	})

	acct, _, err := client.Account.Get(ctx)
	if err != nil {
		t.Errorf("Account.Get returned error: %v", err)
	}

	expected := &Account{
		DropletLimit:    25,
		FloatingIPLimit: 25,
		Email:           "sammy@digitalocean.com",
		UUID:            "b6fr89dbf6d9156cace5f3c78dc9851d957381ef",
		EmailVerified:   true,
		VolumeLimit:     22,
		Team: &TeamInfo{
			Name: "My Team",
			UUID: "b6fr89dbf6d9156cace5f3c78dc9851d957381ef",
		},
	}
	if !reflect.DeepEqual(acct, expected) {
		t.Errorf("Account.Get returned %+v, expected %+v", acct, expected)
	}
}

func TestAccountStringWithTeam(t *testing.T) {
	acct := &Account{
		DropletLimit:    25,
		FloatingIPLimit: 25,
		ReservedIPLimit: 25,
		VolumeLimit:     22,
		Email:           "sammy@digitalocean.com",
		UUID:            "b6fr89dbf6d9156cace5f3c78dc9851d957381ef",
		EmailVerified:   true,
		Status:          "active",
		StatusMessage:   "message",
		Team: &TeamInfo{
			Name: "My Team",
			UUID: "b6fr89dbf6d9156cace5f3c78dc9851d957381ef",
		},
	}

	stringified := acct.String()
	expected := `godo.Account{DropletLimit:25, FloatingIPLimit:25, ReservedIPLimit:25, VolumeLimit:22, Email:"sammy@digitalocean.com", UUID:"b6fr89dbf6d9156cace5f3c78dc9851d957381ef", EmailVerified:true, Status:"active", StatusMessage:"message", Team:godo.TeamInfo{Name:"My Team", UUID:"b6fr89dbf6d9156cace5f3c78dc9851d957381ef"}}`
	if expected != stringified {
		t.Errorf("\n     got %+v\nexpected %+v", stringified, expected)
	}

}
