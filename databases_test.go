package godo

import (
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var privateNetworkUUID = "880b7f98-f062-404d-b33c-458d545696f6"

var db = Database{
	ID:          "da4e0206-d019-41d7-b51f-deadbeefbb8f",
	Name:        "dbtest",
	EngineSlug:  "pg",
	VersionSlug: "11",
	Connection: &DatabaseConnection{
		URI:      "postgres://doadmin:zt91mum075ofzyww@dbtest-do-user-3342561-0.db.ondigitalocean.com:25060/defaultdb?sslmode=require",
		Database: "defaultdb",
		Host:     "dbtest-do-user-3342561-0.db.ondigitalocean.com",
		Port:     25060,
		User:     "doadmin",
		Password: "zt91mum075ofzyww",
		SSL:      true,
	},
	StandbyConnection: &DatabaseConnection{
		URI:      "postgres://doadmin:zt91mum075ofzyww@replica-dbtest-do-user-3342561-0.db.ondigitalocean.com:25060/defaultdb?sslmode=require",
		Database: "defaultdb",
		Host:     "replica-dbtest-do-user-3342561-0.db.ondigitalocean.com",
		Port:     25060,
		User:     "doadmin",
		Password: "zt91mum075ofzyww",
		SSL:      true,
	},
	PrivateConnection: &DatabaseConnection{
		URI:      "postgres://doadmin:zt91mum075ofzyww@private-dbtest-do-user-3342561-0.db.ondigitalocean.com:25060/defaultdb?sslmode=require",
		Database: "defaultdb",
		Host:     "dbtest-do-user-3342561-0.db.ondigitalocean.com",
		Port:     25060,
		User:     "doadmin",
		Password: "zt91mum075ofzyww",
		SSL:      true,
	},
	StandbyPrivateConnection: &DatabaseConnection{
		URI:      "postgres://doadmin:zt91mum075ofzyww@private-replica-dbtest-do-user-3342561-0.db.ondigitalocean.com:25060/defaultdb?sslmode=require",
		Database: "defaultdb",
		Host:     "private-replica-dbtest-do-user-3342561-0.db.ondigitalocean.com",
		Port:     25060,
		User:     "doadmin",
		Password: "zt91mum075ofzyww",
		SSL:      true,
	},
	Users: []DatabaseUser{
		{
			Name:     "doadmin",
			Role:     "primary",
			Password: "zt91mum075ofzyww",
		},
	},
	DBNames: []string{
		"defaultdb",
	},
	NumNodes:   3,
	RegionSlug: "sfo2",
	Status:     "online",
	CreatedAt:  time.Date(2019, 2, 26, 6, 12, 39, 0, time.UTC),
	MaintenanceWindow: &DatabaseMaintenanceWindow{
		Day:         "monday",
		Hour:        "13:51:14",
		Pending:     false,
		Description: nil,
	},
	SizeSlug:           "db-s-2vcpu-4gb",
	PrivateNetworkUUID: "da4e0206-d019-41d7-b51f-deadbeefbb8f",
	Tags:               []string{"production", "staging"},
	ProjectID:          "6d0f9073-0a24-4f1b-9065-7dc5c8bad3e2",
	StorageSizeMib:     61440,
	MetricsEndpoints: []*ServiceAddress{
		{
			Host: "dbtest-do-user-3342561-0.db.ondigitalocean.com",
			Port: 9273,
		},
		{
			Host: "replica-dbtest-do-user-3342561-0.db.ondigitalocean.com",
			Port: 9273,
		},
	},
}

var dbJSON = `
{
	"id": "da4e0206-d019-41d7-b51f-deadbeefbb8f",
	"name": "dbtest",
	"engine": "pg",
	"version": "11",
	"connection": {
		"uri": "postgres://doadmin:zt91mum075ofzyww@dbtest-do-user-3342561-0.db.ondigitalocean.com:25060/defaultdb?sslmode=require",
		"database": "defaultdb",
		"host": "dbtest-do-user-3342561-0.db.ondigitalocean.com",
		"port": 25060,
		"user": "doadmin",
		"password": "zt91mum075ofzyww",
		"ssl": true
	},
	"private_connection": {
		"uri": "postgres://doadmin:zt91mum075ofzyww@private-dbtest-do-user-3342561-0.db.ondigitalocean.com:25060/defaultdb?sslmode=require",
		"database": "defaultdb",
		"host": "dbtest-do-user-3342561-0.db.ondigitalocean.com",
		"port": 25060,
		"user": "doadmin",
		"password": "zt91mum075ofzyww",
		"ssl": true
	},
	"standby_connection": {
		"uri": "postgres://doadmin:zt91mum075ofzyww@replica-dbtest-do-user-3342561-0.db.ondigitalocean.com:25060/defaultdb?sslmode=require",
		"database": "defaultdb",
		"host": "replica-dbtest-do-user-3342561-0.db.ondigitalocean.com",
		"port": 25060,
		"user": "doadmin",
		"password": "zt91mum075ofzyww",
		"ssl": true
	},
	"standby_private_connection": {
		"uri": "postgres://doadmin:zt91mum075ofzyww@private-replica-dbtest-do-user-3342561-0.db.ondigitalocean.com:25060/defaultdb?sslmode=require",
		"database": "defaultdb",
		"host": "private-replica-dbtest-do-user-3342561-0.db.ondigitalocean.com",
		"port": 25060,
		"user": "doadmin",
		"password": "zt91mum075ofzyww",
		"ssl": true
	},
	"users": [
		{
			"name": "doadmin",
			"role": "primary",
			"password": "zt91mum075ofzyww"
		}
	],
	"db_names": [
		"defaultdb"
	],
	"num_nodes": 3,
	"region": "sfo2",
	"status": "online",
	"created_at": "2019-02-26T06:12:39Z",
	"maintenance_window": {
		"day": "monday",
		"hour": "13:51:14",
		"pending": false,
		"description": null
	},
	"size": "db-s-2vcpu-4gb",
	"private_network_uuid": "da4e0206-d019-41d7-b51f-deadbeefbb8f",
	"tags": ["production", "staging"],
	"project_id": "6d0f9073-0a24-4f1b-9065-7dc5c8bad3e2",
	"storage_size_mib": 61440,
	"metrics_endpoints": [
		{
			"host": "dbtest-do-user-3342561-0.db.ondigitalocean.com",
			"port": 9273
		},
		{
			"host": "replica-dbtest-do-user-3342561-0.db.ondigitalocean.com",
			"port": 9273
		}
	]
}
`

var dbsJSON = fmt.Sprintf(`
{
  "databases": [
	%s
  ]
}
`, dbJSON)

func TestDatabases_List(t *testing.T) {
	setup()
	defer teardown()

	dbSvc := client.Databases

	want := []Database{db}

	mux.HandleFunc("/v2/databases", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, dbsJSON)
	})

	got, _, err := dbSvc.List(ctx, nil)
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestDatabases_Get(t *testing.T) {
	setup()
	defer teardown()

	dbID := "da4e0206-d019-41d7-b51f-deadbeefbb8f"

	body := fmt.Sprintf(`
{
  "database": %s
}
`, dbJSON)

	path := fmt.Sprintf("/v2/databases/%s", dbID)

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, body)
	})

	got, _, err := client.Databases.Get(ctx, dbID)
	require.NoError(t, err)
	require.Equal(t, &db, got)
}

func TestDatabases_GetCA(t *testing.T) {
	setup()
	defer teardown()

	dbID := "da4e0206-d019-41d7-b51f-deadbeefbb8f"

	body := `
{
  "ca": {
    "certificate": "ZmFrZQpjYQpjZXJ0"
  }
}
`

	path := fmt.Sprintf("/v2/databases/%s/ca", dbID)

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, body)
	})

	got, _, err := client.Databases.GetCA(ctx, dbID)
	require.NoError(t, err)
	require.Equal(t, &DatabaseCA{Certificate: []byte("fake\nca\ncert")}, got)
}

func TestDatabases_Create(t *testing.T) {
	tests := []struct {
		title         string
		createRequest *DatabaseCreateRequest
		want          *Database
		body          string
	}{
		{
			title: "create",
			createRequest: &DatabaseCreateRequest{
				Name:       "backend-test",
				EngineSlug: "pg",
				Version:    "10",
				Region:     "nyc3",
				SizeSlug:   "db-s-2vcpu-4gb",
				NumNodes:   2,
				Tags:       []string{"production", "staging"},
				ProjectID:  "05d84f74-db8c-4de5-ae72-2fd4823fb1c8",
			},
			want: &Database{
				ID:          "8d91899c-0739-4a1a-acc5-deadbeefbb8f",
				Name:        "backend-test",
				EngineSlug:  "pg",
				VersionSlug: "10",
				Connection: &DatabaseConnection{
					URI:      "postgres://doadmin:zt91mum075ofzyww@dbtest-do-user-3342561-0.db.ondigitalocean.com:25060/defaultdb?sslmode=require",
					Database: "defaultdb",
					Host:     "dbtest-do-user-3342561-0.db.ondigitalocean.com",
					Port:     25060,
					User:     "doadmin",
					Password: "zt91mum075ofzyww",
					SSL:      true,
				},
				StandbyConnection: &DatabaseConnection{
					URI:      "postgres://doadmin:zt91mum075ofzyww@replica-dbtest-do-user-3342561-0.db.ondigitalocean.com:25060/defaultdb?sslmode=require",
					Database: "defaultdb",
					Host:     "replica-dbtest-do-user-3342561-0.db.ondigitalocean.com",
					Port:     25060,
					User:     "doadmin",
					Password: "zt91mum075ofzyww",
					SSL:      true,
				},
				PrivateConnection: &DatabaseConnection{
					URI:      "postgres://doadmin:zt91mum075ofzyww@private-dbtest-do-user-3342561-0.db.ondigitalocean.com:25060/defaultdb?sslmode=require",
					Database: "defaultdb",
					Host:     "private-dbtest-do-user-3342561-0.db.ondigitalocean.com",
					Port:     25060,
					User:     "doadmin",
					Password: "zt91mum075ofzyww",
					SSL:      true,
				},
				StandbyPrivateConnection: &DatabaseConnection{
					URI:      "postgres://doadmin:zt91mum075ofzyww@private-replica-dbtest-do-user-3342561-0.db.ondigitalocean.com:25060/defaultdb?sslmode=require",
					Database: "defaultdb",
					Host:     "private-replica-dbtest-do-user-3342561-0.db.ondigitalocean.com",
					Port:     25060,
					User:     "doadmin",
					Password: "zt91mum075ofzyww",
					SSL:      true,
				},
				Users:             nil,
				DBNames:           nil,
				NumNodes:          2,
				RegionSlug:        "nyc3",
				Status:            "creating",
				CreatedAt:         time.Date(2019, 2, 26, 6, 12, 39, 0, time.UTC),
				MaintenanceWindow: nil,
				SizeSlug:          "db-s-2vcpu-4gb",
				Tags:              []string{"production", "staging"},
				ProjectID:         "05d84f74-db8c-4de5-ae72-2fd4823fb1c8",
				StorageSizeMib:    61440,
			},
			body: `
{
	"database": {
		"id": "8d91899c-0739-4a1a-acc5-deadbeefbb8f",
		"name": "backend-test",
		"engine": "pg",
		"version": "10",
		"connection": {
			"uri": "postgres://doadmin:zt91mum075ofzyww@dbtest-do-user-3342561-0.db.ondigitalocean.com:25060/defaultdb?sslmode=require",
			"database": "defaultdb",
			"host": "dbtest-do-user-3342561-0.db.ondigitalocean.com",
			"port": 25060,
			"user": "doadmin",
			"password": "zt91mum075ofzyww",
			"ssl": true
		},
		"private_connection": {
			"uri": "postgres://doadmin:zt91mum075ofzyww@private-dbtest-do-user-3342561-0.db.ondigitalocean.com:25060/defaultdb?sslmode=require",
			"database": "defaultdb",
			"host": "private-dbtest-do-user-3342561-0.db.ondigitalocean.com",
			"port": 25060,
			"user": "doadmin",
			"password": "zt91mum075ofzyww",
			"ssl": true
		},
		"standby_connection": {
			"uri": "postgres://doadmin:zt91mum075ofzyww@replica-dbtest-do-user-3342561-0.db.ondigitalocean.com:25060/defaultdb?sslmode=require",
			"database": "defaultdb",
			"host": "replica-dbtest-do-user-3342561-0.db.ondigitalocean.com",
			"port": 25060,
			"user": "doadmin",
			"password": "zt91mum075ofzyww",
			"ssl": true
		},
		"standby_private_connection": {
			"uri": "postgres://doadmin:zt91mum075ofzyww@private-replica-dbtest-do-user-3342561-0.db.ondigitalocean.com:25060/defaultdb?sslmode=require",
			"database": "defaultdb",
			"host": "private-replica-dbtest-do-user-3342561-0.db.ondigitalocean.com",
			"port": 25060,
			"user": "doadmin",
			"password": "zt91mum075ofzyww",
			"ssl": true
		},
		"users": null,
		"db_names": null,
		"num_nodes": 2,
		"region": "nyc3",
		"status": "creating",
		"created_at": "2019-02-26T06:12:39Z",
		"maintenance_window": null,
		"size": "db-s-2vcpu-4gb",
		"tags": ["production", "staging"],
		"project_id": "05d84f74-db8c-4de5-ae72-2fd4823fb1c8",
		"storage_size_mib": 61440
	}
}`,
		},
		{
			title: "create from backup",
			createRequest: &DatabaseCreateRequest{
				Name:       "backend-restored",
				EngineSlug: "pg",
				Version:    "10",
				Region:     "nyc3",
				SizeSlug:   "db-s-2vcpu-4gb",
				NumNodes:   2,
				Tags:       []string{"production", "staging"},
				BackupRestore: &DatabaseBackupRestore{
					DatabaseName:    "backend-orig",
					BackupCreatedAt: "2019-01-31T19:25:22Z",
				},
			},
			want: &Database{
				ID:          "8d91899c-0739-4a1a-acc5-deadbeefbb8f",
				Name:        "backend-test",
				EngineSlug:  "pg",
				VersionSlug: "10",
				Connection: &DatabaseConnection{
					URI:      "postgres://doadmin:zt91mum075ofzyww@dbtest-do-user-3342561-0.db.ondigitalocean.com:25060/defaultdb?sslmode=require",
					Database: "defaultdb",
					Host:     "dbtest-do-user-3342561-0.db.ondigitalocean.com",
					Port:     25060,
					User:     "doadmin",
					Password: "zt91mum075ofzyww",
					SSL:      true,
				},
				PrivateConnection: &DatabaseConnection{
					URI:      "postgres://doadmin:zt91mum075ofzyww@private-dbtest-do-user-3342561-0.db.ondigitalocean.com:25060/defaultdb?sslmode=require",
					Database: "defaultdb",
					Host:     "dbtest-do-user-3342561-0.db.ondigitalocean.com",
					Port:     25060,
					User:     "doadmin",
					Password: "zt91mum075ofzyww",
					SSL:      true,
				},
				StandbyConnection: &DatabaseConnection{
					URI:      "postgres://doadmin:zt91mum075ofzyww@replica-dbtest-do-user-3342561-0.db.ondigitalocean.com:25060/defaultdb?sslmode=require",
					Database: "defaultdb",
					Host:     "replica-dbtest-do-user-3342561-0.db.ondigitalocean.com",
					Port:     25060,
					User:     "doadmin",
					Password: "zt91mum075ofzyww",
					SSL:      true,
				},
				StandbyPrivateConnection: &DatabaseConnection{
					URI:      "postgres://doadmin:zt91mum075ofzyww@private-replica-dbtest-do-user-3342561-0.db.ondigitalocean.com:25060/defaultdb?sslmode=require",
					Database: "defaultdb",
					Host:     "private-replica-dbtest-do-user-3342561-0.db.ondigitalocean.com",
					Port:     25060,
					User:     "doadmin",
					Password: "zt91mum075ofzyww",
					SSL:      true,
				},
				Users:             nil,
				DBNames:           nil,
				NumNodes:          2,
				RegionSlug:        "nyc3",
				Status:            "creating",
				CreatedAt:         time.Date(2019, 2, 26, 6, 12, 39, 0, time.UTC),
				MaintenanceWindow: nil,
				SizeSlug:          "db-s-2vcpu-4gb",
				Tags:              []string{"production", "staging"},
				StorageSizeMib:    61440,
			},
			body: `
{
	"database": {
		"id": "8d91899c-0739-4a1a-acc5-deadbeefbb8f",
		"name": "backend-test",
		"engine": "pg",
		"version": "10",
		"connection": {
			"uri": "postgres://doadmin:zt91mum075ofzyww@dbtest-do-user-3342561-0.db.ondigitalocean.com:25060/defaultdb?sslmode=require",
			"database": "defaultdb",
			"host": "dbtest-do-user-3342561-0.db.ondigitalocean.com",
			"port": 25060,
			"user": "doadmin",
			"password": "zt91mum075ofzyww",
			"ssl": true
		},
		"private_connection": {
			"uri": "postgres://doadmin:zt91mum075ofzyww@private-dbtest-do-user-3342561-0.db.ondigitalocean.com:25060/defaultdb?sslmode=require",
			"database": "defaultdb",
			"host": "dbtest-do-user-3342561-0.db.ondigitalocean.com",
			"port": 25060,
			"user": "doadmin",
			"password": "zt91mum075ofzyww",
			"ssl": true
		},
		"standby_connection": {
			"uri": "postgres://doadmin:zt91mum075ofzyww@replica-dbtest-do-user-3342561-0.db.ondigitalocean.com:25060/defaultdb?sslmode=require",
			"database": "defaultdb",
			"host": "replica-dbtest-do-user-3342561-0.db.ondigitalocean.com",
			"port": 25060,
			"user": "doadmin",
			"password": "zt91mum075ofzyww",
			"ssl": true
		},
		"standby_private_connection": {
			"uri": "postgres://doadmin:zt91mum075ofzyww@private-replica-dbtest-do-user-3342561-0.db.ondigitalocean.com:25060/defaultdb?sslmode=require",
			"database": "defaultdb",
			"host": "private-replica-dbtest-do-user-3342561-0.db.ondigitalocean.com",
			"port": 25060,
			"user": "doadmin",
			"password": "zt91mum075ofzyww",
			"ssl": true
		},
		"users": null,
		"db_names": null,
		"num_nodes": 2,
		"region": "nyc3",
		"status": "creating",
		"created_at": "2019-02-26T06:12:39Z",
		"maintenance_window": null,
		"size": "db-s-2vcpu-4gb",
		"tags": ["production", "staging"],
		"storage_size_mib": 61440
	}
}`,
		},
		{
			title: "create with additional storage",
			createRequest: &DatabaseCreateRequest{
				Name:           "additional-storage-test",
				EngineSlug:     "pg",
				Version:        "15",
				Region:         "nyc3",
				SizeSlug:       "db-s-2vcpu-4gb",
				NumNodes:       2,
				Tags:           []string{"production", "staging"},
				ProjectID:      "05d84f74-db8c-4de5-ae72-2fd4823fb1c8",
				StorageSizeMib: 81920,
			},
			want: &Database{
				ID:          "8d91899c-0739-4a1a-acc5-deadbeefbb8f",
				Name:        "backend-test",
				EngineSlug:  "pg",
				VersionSlug: "10",
				Connection: &DatabaseConnection{
					URI:      "postgres://doadmin:zt91mum075ofzyww@dbtest-do-user-3342561-0.db.ondigitalocean.com:25060/defaultdb?sslmode=require",
					Database: "defaultdb",
					Host:     "dbtest-do-user-3342561-0.db.ondigitalocean.com",
					Port:     25060,
					User:     "doadmin",
					Password: "zt91mum075ofzyww",
					SSL:      true,
				},
				PrivateConnection: &DatabaseConnection{
					URI:      "postgres://doadmin:zt91mum075ofzyww@private-dbtest-do-user-3342561-0.db.ondigitalocean.com:25060/defaultdb?sslmode=require",
					Database: "defaultdb",
					Host:     "dbtest-do-user-3342561-0.db.ondigitalocean.com",
					Port:     25060,
					User:     "doadmin",
					Password: "zt91mum075ofzyww",
					SSL:      true,
				},
				StandbyConnection: &DatabaseConnection{
					URI:      "postgres://doadmin:zt91mum075ofzyww@replica-dbtest-do-user-3342561-0.db.ondigitalocean.com:25060/defaultdb?sslmode=require",
					Database: "defaultdb",
					Host:     "replica-dbtest-do-user-3342561-0.db.ondigitalocean.com",
					Port:     25060,
					User:     "doadmin",
					Password: "zt91mum075ofzyww",
					SSL:      true,
				},
				StandbyPrivateConnection: &DatabaseConnection{
					URI:      "postgres://doadmin:zt91mum075ofzyww@private-replica-dbtest-do-user-3342561-0.db.ondigitalocean.com:25060/defaultdb?sslmode=require",
					Database: "defaultdb",
					Host:     "private-replica-dbtest-do-user-3342561-0.db.ondigitalocean.com",
					Port:     25060,
					User:     "doadmin",
					Password: "zt91mum075ofzyww",
					SSL:      true,
				},
				Users:             nil,
				DBNames:           nil,
				NumNodes:          2,
				RegionSlug:        "nyc3",
				Status:            "creating",
				CreatedAt:         time.Date(2019, 2, 26, 6, 12, 39, 0, time.UTC),
				MaintenanceWindow: nil,
				SizeSlug:          "db-s-2vcpu-4gb",
				Tags:              []string{"production", "staging"},
				ProjectID:         "05d84f74-db8c-4de5-ae72-2fd4823fb1c8",
				StorageSizeMib:    81920,
			},
			body: `
{
	"database": {
		"id": "8d91899c-0739-4a1a-acc5-deadbeefbb8f",
		"name": "backend-test",
		"engine": "pg",
		"version": "10",
		"connection": {
			"uri": "postgres://doadmin:zt91mum075ofzyww@dbtest-do-user-3342561-0.db.ondigitalocean.com:25060/defaultdb?sslmode=require",
			"database": "defaultdb",
			"host": "dbtest-do-user-3342561-0.db.ondigitalocean.com",
			"port": 25060,
			"user": "doadmin",
			"password": "zt91mum075ofzyww",
			"ssl": true
		},
		"private_connection": {
			"uri": "postgres://doadmin:zt91mum075ofzyww@private-dbtest-do-user-3342561-0.db.ondigitalocean.com:25060/defaultdb?sslmode=require",
			"database": "defaultdb",
			"host": "dbtest-do-user-3342561-0.db.ondigitalocean.com",
			"port": 25060,
			"user": "doadmin",
			"password": "zt91mum075ofzyww",
			"ssl": true
		},
		"standby_connection": {
			"uri": "postgres://doadmin:zt91mum075ofzyww@replica-dbtest-do-user-3342561-0.db.ondigitalocean.com:25060/defaultdb?sslmode=require",
			"database": "defaultdb",
			"host": "replica-dbtest-do-user-3342561-0.db.ondigitalocean.com",
			"port": 25060,
			"user": "doadmin",
			"password": "zt91mum075ofzyww",
			"ssl": true
		},
		"standby_private_connection": {
			"uri": "postgres://doadmin:zt91mum075ofzyww@private-replica-dbtest-do-user-3342561-0.db.ondigitalocean.com:25060/defaultdb?sslmode=require",
			"database": "defaultdb",
			"host": "private-replica-dbtest-do-user-3342561-0.db.ondigitalocean.com",
			"port": 25060,
			"user": "doadmin",
			"password": "zt91mum075ofzyww",
			"ssl": true
		},
		"users": null,
		"db_names": null,
		"num_nodes": 2,
		"region": "nyc3",
		"status": "creating",
		"created_at": "2019-02-26T06:12:39Z",
		"maintenance_window": null,
		"size": "db-s-2vcpu-4gb",
		"tags": ["production", "staging"],
		"project_id": "05d84f74-db8c-4de5-ae72-2fd4823fb1c8",
		"storage_size_mib": 81920
	}
}`,
		},
		{
			title: "create with firewall rules",
			createRequest: &DatabaseCreateRequest{
				Name:       "firewall-rules-test",
				EngineSlug: "pg",
				Version:    "15",
				Region:     "nyc3",
				SizeSlug:   "db-s-2vcpu-4gb",
				NumNodes:   2,
				Tags:       []string{"production", "staging"},
				ProjectID:  "05d84f74-db8c-4de5-ae72-2fd4823fb1c8",
				Rules: []*DatabaseCreateFirewallRule{
					{
						UUID:  "bc47473b-603e-49a8-b36a-810c2703f1d0",
						Type:  "ip_addr",
						Value: "172.16.1.1",
					},
					{
						UUID:  "17d460b2-5879-4466-ac09-6c90c9a6d7e0",
						Type:  "tag",
						Value: "production",
					},
				},
			},
			want: &Database{
				ID:          "8d91899c-0739-4a1a-acc5-deadbeefbb8f",
				Name:        "firewall-rules-test",
				EngineSlug:  "pg",
				VersionSlug: "10",
				Connection: &DatabaseConnection{
					URI:      "postgres://doadmin:zt91mum075ofzyww@dbtest-do-user-3342561-0.db.ondigitalocean.com:25060/defaultdb?sslmode=require",
					Database: "defaultdb",
					Host:     "dbtest-do-user-3342561-0.db.ondigitalocean.com",
					Port:     25060,
					User:     "doadmin",
					Password: "zt91mum075ofzyww",
					SSL:      true,
				},
				PrivateConnection: &DatabaseConnection{
					URI:      "postgres://doadmin:zt91mum075ofzyww@private-dbtest-do-user-3342561-0.db.ondigitalocean.com:25060/defaultdb?sslmode=require",
					Database: "defaultdb",
					Host:     "dbtest-do-user-3342561-0.db.ondigitalocean.com",
					Port:     25060,
					User:     "doadmin",
					Password: "zt91mum075ofzyww",
					SSL:      true,
				},
				StandbyConnection: &DatabaseConnection{
					URI:      "postgres://doadmin:zt91mum075ofzyww@replica-dbtest-do-user-3342561-0.db.ondigitalocean.com:25060/defaultdb?sslmode=require",
					Database: "defaultdb",
					Host:     "replica-dbtest-do-user-3342561-0.db.ondigitalocean.com",
					Port:     25060,
					User:     "doadmin",
					Password: "zt91mum075ofzyww",
					SSL:      true,
				},
				StandbyPrivateConnection: &DatabaseConnection{
					URI:      "postgres://doadmin:zt91mum075ofzyww@private-replica-dbtest-do-user-3342561-0.db.ondigitalocean.com:25060/defaultdb?sslmode=require",
					Database: "defaultdb",
					Host:     "private-replica-dbtest-do-user-3342561-0.db.ondigitalocean.com",
					Port:     25060,
					User:     "doadmin",
					Password: "zt91mum075ofzyww",
					SSL:      true,
				},
				Users:             nil,
				DBNames:           nil,
				NumNodes:          2,
				RegionSlug:        "nyc3",
				Status:            "creating",
				CreatedAt:         time.Date(2019, 2, 26, 6, 12, 39, 0, time.UTC),
				MaintenanceWindow: nil,
				SizeSlug:          "db-s-2vcpu-4gb",
				Tags:              []string{"production", "staging"},
				ProjectID:         "05d84f74-db8c-4de5-ae72-2fd4823fb1c8",
			},
			body: `
{
	"database": {
		"id": "8d91899c-0739-4a1a-acc5-deadbeefbb8f",
		"name": "firewall-rules-test",
		"engine": "pg",
		"version": "10",
		"connection": {
			"uri": "postgres://doadmin:zt91mum075ofzyww@dbtest-do-user-3342561-0.db.ondigitalocean.com:25060/defaultdb?sslmode=require",
			"database": "defaultdb",
			"host": "dbtest-do-user-3342561-0.db.ondigitalocean.com",
			"port": 25060,
			"user": "doadmin",
			"password": "zt91mum075ofzyww",
			"ssl": true
		},
		"private_connection": {
			"uri": "postgres://doadmin:zt91mum075ofzyww@private-dbtest-do-user-3342561-0.db.ondigitalocean.com:25060/defaultdb?sslmode=require",
			"database": "defaultdb",
			"host": "dbtest-do-user-3342561-0.db.ondigitalocean.com",
			"port": 25060,
			"user": "doadmin",
			"password": "zt91mum075ofzyww",
			"ssl": true
		},
		"standby_connection": {
			"uri": "postgres://doadmin:zt91mum075ofzyww@replica-dbtest-do-user-3342561-0.db.ondigitalocean.com:25060/defaultdb?sslmode=require",
			"database": "defaultdb",
			"host": "replica-dbtest-do-user-3342561-0.db.ondigitalocean.com",
			"port": 25060,
			"user": "doadmin",
			"password": "zt91mum075ofzyww",
			"ssl": true
		},
		"standby_private_connection": {
			"uri": "postgres://doadmin:zt91mum075ofzyww@private-replica-dbtest-do-user-3342561-0.db.ondigitalocean.com:25060/defaultdb?sslmode=require",
			"database": "defaultdb",
			"host": "private-replica-dbtest-do-user-3342561-0.db.ondigitalocean.com",
			"port": 25060,
			"user": "doadmin",
			"password": "zt91mum075ofzyww",
			"ssl": true
		},
		"users": null,
		"db_names": null,
		"num_nodes": 2,
		"region": "nyc3",
		"status": "creating",
		"created_at": "2019-02-26T06:12:39Z",
		"maintenance_window": null,
		"size": "db-s-2vcpu-4gb",
		"tags": ["production", "staging"],
		"project_id": "05d84f74-db8c-4de5-ae72-2fd4823fb1c8"
	}
}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			setup()
			defer teardown()

			mux.HandleFunc("/v2/databases", func(w http.ResponseWriter, r *http.Request) {
				v := new(DatabaseCreateRequest)
				err := json.NewDecoder(r.Body).Decode(v)
				if err != nil {
					t.Fatal(err)
				}

				testMethod(t, r, http.MethodPost)
				require.Equal(t, v, tt.createRequest)
				fmt.Fprint(w, tt.body)
			})

			got, _, err := client.Databases.Create(ctx, tt.createRequest)
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestDatabases_Delete(t *testing.T) {
	setup()
	defer teardown()

	dbID := "deadbeef-dead-4aa5-beef-deadbeef347d"

	path := fmt.Sprintf("/v2/databases/%s", dbID)

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
	})

	_, err := client.Databases.Delete(ctx, "deadbeef-dead-4aa5-beef-deadbeef347d")
	require.NoError(t, err)
}

func TestDatabases_Resize(t *testing.T) {
	setup()
	defer teardown()

	resizeRequest := &DatabaseResizeRequest{
		SizeSlug:       "db-s-16vcpu-64gb",
		NumNodes:       3,
		StorageSizeMib: 921600,
	}

	dbID := "deadbeef-dead-4aa5-beef-deadbeef347d"
	path := fmt.Sprintf("/v2/databases/%s/resize", dbID)

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
	})

	_, err := client.Databases.Resize(ctx, dbID, resizeRequest)
	require.NoError(t, err)
}

func TestDatabases_Migrate(t *testing.T) {
	setup()
	defer teardown()

	migrateRequest := &DatabaseMigrateRequest{
		Region: "lon1",
	}

	dbID := "deadbeef-dead-4aa5-beef-deadbeef347d"

	path := fmt.Sprintf("/v2/databases/%s/migrate", dbID)

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
	})

	_, err := client.Databases.Migrate(ctx, "deadbeef-dead-4aa5-beef-deadbeef347d", migrateRequest)
	require.NoError(t, err)
}

func TestDatabases_Migrate_WithPrivateNet(t *testing.T) {
	setup()
	defer teardown()

	migrateRequest := &DatabaseMigrateRequest{
		Region:             "lon1",
		PrivateNetworkUUID: privateNetworkUUID,
	}

	dbID := "deadbeef-dead-4aa5-beef-deadbeef347d"

	path := fmt.Sprintf("/v2/databases/%s/migrate", dbID)

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
	})

	_, err := client.Databases.Migrate(ctx, "deadbeef-dead-4aa5-beef-deadbeef347d", migrateRequest)
	require.NoError(t, err)
}

func TestDatabases_UpdateMaintenance(t *testing.T) {
	setup()
	defer teardown()

	maintenanceRequest := &DatabaseUpdateMaintenanceRequest{
		Day:  "thursday",
		Hour: "16:00",
	}

	dbID := "deadbeef-dead-4aa5-beef-deadbeef347d"

	path := fmt.Sprintf("/v2/databases/%s/maintenance", dbID)

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
	})

	_, err := client.Databases.UpdateMaintenance(ctx, "deadbeef-dead-4aa5-beef-deadbeef347d", maintenanceRequest)
	require.NoError(t, err)
}

func TestDatabases_InstallUpdate(t *testing.T) {
	setup()
	defer teardown()

	dbID := "deadbeef-dead-4aa5-beef-deadbeef347d"

	path := fmt.Sprintf("/v2/databases/%s/install_update", dbID)

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
	})

	_, err := client.Databases.InstallUpdate(ctx, "deadbeef-dead-4aa5-beef-deadbeef347d")
	require.NoError(t, err)
}

func TestDatabases_ListBackups(t *testing.T) {
	setup()
	defer teardown()

	want := []DatabaseBackup{
		{
			CreatedAt:     time.Date(2019, 1, 11, 18, 42, 27, 0, time.UTC),
			SizeGigabytes: 0.03357696,
		},
		{
			CreatedAt:     time.Date(2019, 1, 12, 18, 42, 29, 0, time.UTC),
			SizeGigabytes: 0.03364864,
		},
	}

	body := `
{
  "backups": [
    {
      "created_at": "2019-01-11T18:42:27Z",
      "size_gigabytes": 0.03357696
    },
    {
      "created_at": "2019-01-12T18:42:29Z",
      "size_gigabytes": 0.03364864
    }
  ]
}
`

	dbID := "deadbeef-dead-4aa5-beef-deadbeef347d"

	path := fmt.Sprintf("/v2/databases/%s/backups", dbID)

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, body)
	})

	got, _, err := client.Databases.ListBackups(ctx, "deadbeef-dead-4aa5-beef-deadbeef347d", nil)
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestDatabases_GetUser(t *testing.T) {
	setup()
	defer teardown()

	want := &DatabaseUser{
		Name:     "name",
		Role:     "foo",
		Password: "pass",
	}

	body := `
{
  "user": {
    "name": "name",
    "role": "foo",
    "password": "pass"
  }
}
`

	dbID := "deadbeef-dead-4aa5-beef-deadbeef347d"

	path := fmt.Sprintf("/v2/databases/%s/users/name", dbID)

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, body)
	})

	got, _, err := client.Databases.GetUser(ctx, dbID, "name")
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestDatabases_ListUsers(t *testing.T) {
	setup()
	defer teardown()

	dbID := "deadbeef-dead-4aa5-beef-deadbeef347d"

	want := []DatabaseUser{
		{
			Name:     "name",
			Role:     "foo",
			Password: "pass",
		},
		{
			Name:     "bar",
			Role:     "foo",
			Password: "pass",
		},
	}

	body := `
{
  "users": [{
    "name": "name",
    "role": "foo",
    "password": "pass"
  },
  {
    "name": "bar",
    "role": "foo",
    "password": "pass"
  }]
}
`
	path := fmt.Sprintf("/v2/databases/%s/users", dbID)

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, body)
	})

	got, _, err := client.Databases.ListUsers(ctx, dbID, nil)
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestDatabases_CreateUser(t *testing.T) {
	setup()
	defer teardown()

	dbID := "deadbeef-dead-4aa5-beef-deadbeef347d"

	want := &DatabaseUser{
		Name:     "name",
		Role:     "foo",
		Password: "pass",
	}

	body := `
{
  "user": {
    "name": "name",
    "role": "foo",
    "password": "pass"
  }
}
`
	path := fmt.Sprintf("/v2/databases/%s/users", dbID)

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, body)
	})

	got, _, err := client.Databases.CreateUser(ctx, dbID, &DatabaseCreateUserRequest{Name: "user"})
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestDatabases_UpdateUser(t *testing.T) {
	setup()
	defer teardown()

	dbID := "deadbeef-dead-4aa5-beef-deadbeef347d"
	userID := "test-user"

	want := &DatabaseUser{
		Name: userID,
		Settings: &DatabaseUserSettings{
			ACL: []*KafkaACL{
				{
					Topic:      "events",
					Permission: "produce_consume",
				},
			},
		},
	}

	body := `
{
  "user": {
	"name": "test-user",
	"settings": {
		"acl": [
			{
				"permission": "produce_consume",
				"topic": "events"
			}
		]
	}
  }
}
`
	path := fmt.Sprintf("/v2/databases/%s/users/%s", dbID, userID)

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		fmt.Fprint(w, body)
	})

	got, _, err := client.Databases.UpdateUser(ctx, dbID, userID, &DatabaseUpdateUserRequest{
		Settings: &DatabaseUserSettings{
			ACL: []*KafkaACL{
				{
					Topic:      "events",
					Permission: "produce_consume",
				},
			},
		},
	})
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestDatabases_UpdateUser_OpenSearchACL(t *testing.T) {
	setup()
	defer teardown()

	dbID := "deadbeef-dead-4aa5-beef-deadbeef347d"
	userID := "test-user"

	want := &DatabaseUser{
		Name: userID,
		Settings: &DatabaseUserSettings{
			OpenSearchACL: []*OpenSearchACL{
				{
					Index:      "sample-index",
					Permission: "read",
				},
			},
		},
	}

	body := `
{
  "user": {
	"name": "test-user",
	"settings": {
		"opensearch_acl": [
			{
				"permission": "read",
				"index": "sample-index"
			}
		]
	}
  }
}
`
	path := fmt.Sprintf("/v2/databases/%s/users/%s", dbID, userID)

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		fmt.Fprint(w, body)
	})

	got, _, err := client.Databases.UpdateUser(ctx, dbID, userID, &DatabaseUpdateUserRequest{
		Settings: &DatabaseUserSettings{
			OpenSearchACL: []*OpenSearchACL{
				{
					Index:      "sample-index",
					Permission: "read",
				},
			},
		},
	})
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestDatabases_DeleteUser(t *testing.T) {
	setup()
	defer teardown()

	dbID := "deadbeef-dead-4aa5-beef-deadbeef347d"

	path := fmt.Sprintf("/v2/databases/%s/users/user", dbID)

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
	})

	_, err := client.Databases.DeleteUser(ctx, dbID, "user")
	require.NoError(t, err)
}

func TestDatabases_ResetUserAuth(t *testing.T) {
	setup()
	defer teardown()
	dbID := "deadbeef-dead-4aa5-beef-deadbeef347d"
	path := fmt.Sprintf("/v2/databases/%s/users/user/reset_auth", dbID)

	body := `
{
  "user": {
     "name": "name",
     "role": "foo",
     "password": "pass",
     "mysql_settings": {
       "auth_plugin": "caching_sha2_password"
     }
  }
}
`
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, body)
	})

	want := &DatabaseUser{
		Name:          "name",
		Role:          "foo",
		Password:      "pass",
		MySQLSettings: &DatabaseMySQLUserSettings{AuthPlugin: SQLAuthPluginCachingSHA2},
	}

	got, _, err := client.Databases.ResetUserAuth(ctx, dbID, "user", &DatabaseResetUserAuthRequest{
		MySQLSettings: &DatabaseMySQLUserSettings{
			AuthPlugin: SQLAuthPluginCachingSHA2,
		}})

	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestDatabases_ResetUserAuthKafka(t *testing.T) {
	setup()
	defer teardown()
	dbID := "deadbeef-dead-4aa5-beef-deadbeef347d"
	path := fmt.Sprintf("/v2/databases/%s/users/user/reset_auth", dbID)

	body := `{
	  "user": {
	     "name": "name",
	     "role": "foo",
	     "password": "otherpass"
	  }
	}`

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, body)
	})

	want := &DatabaseUser{
		Name:     "name",
		Role:     "foo",
		Password: "otherpass",
	}

	got, _, err := client.Databases.ResetUserAuth(ctx, dbID, "user", &DatabaseResetUserAuthRequest{
		Settings: &DatabaseUserSettings{},
	})

	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestDatabases_ListDBs(t *testing.T) {
	setup()
	defer teardown()

	dbID := "deadbeef-dead-4aa5-beef-deadbeef347d"

	want := []DatabaseDB{
		{Name: "foo"},
		{Name: "bar"},
	}

	body := `
{
  "dbs": [{
    "name": "foo"
  },
  {
    "name": "bar"
  }]
}
`
	path := fmt.Sprintf("/v2/databases/%s/dbs", dbID)

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, body)
	})

	got, _, err := client.Databases.ListDBs(ctx, dbID, nil)
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestDatabases_CreateDB(t *testing.T) {
	setup()
	defer teardown()

	dbID := "deadbeef-dead-4aa5-beef-deadbeef347d"

	want := &DatabaseDB{
		Name: "foo",
	}

	body := `
{
  "db": {
    "name": "foo"
  }
}
`
	path := fmt.Sprintf("/v2/databases/%s/dbs", dbID)

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, body)
	})

	got, _, err := client.Databases.CreateDB(ctx, dbID, &DatabaseCreateDBRequest{Name: "foo"})
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestDatabases_GetDB(t *testing.T) {
	setup()
	defer teardown()

	dbID := "deadbeef-dead-4aa5-beef-deadbeef347d"

	want := &DatabaseDB{
		Name: "foo",
	}

	body := `
{
  "db": {
    "name": "foo"
  }
}
`
	path := fmt.Sprintf("/v2/databases/%s/dbs/foo", dbID)

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, body)
	})

	got, _, err := client.Databases.GetDB(ctx, dbID, "foo")
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestDatabases_DeleteDB(t *testing.T) {
	setup()
	defer teardown()

	dbID := "deadbeef-dead-4aa5-beef-deadbeef347d"

	body := `
{
  "db": {
    "name": "foo"
  }
}
`
	path := fmt.Sprintf("/v2/databases/%s/dbs/foo", dbID)

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		fmt.Fprint(w, body)
	})

	_, err := client.Databases.DeleteDB(ctx, dbID, "foo")
	require.NoError(t, err)
}

func TestDatabases_ListPools(t *testing.T) {
	setup()
	defer teardown()

	dbID := "deadbeef-dead-4aa5-beef-deadbeef347d"

	want := []DatabasePool{
		{
			Name:     "pool",
			User:     "user",
			Size:     10,
			Mode:     "transaction",
			Database: "db",
			Connection: &DatabaseConnection{
				URI:      "postgresql://user:pass@host.com/db",
				Host:     "host.com",
				Port:     1234,
				User:     "user",
				Password: "pass",
				SSL:      true,
				Database: "db",
			},
			PrivateConnection: &DatabaseConnection{
				URI:      "postgresql://user:pass@private-host.com/db",
				Host:     "private-host.com",
				Port:     1234,
				User:     "user",
				Password: "pass",
				SSL:      true,
				Database: "db",
			},
			StandbyConnection: &DatabaseConnection{
				URI:      "postgresql://user:pass@replica-host.com/db",
				Host:     "replica-host.com",
				Port:     1234,
				User:     "user",
				Password: "pass",
				SSL:      true,
				Database: "db",
			},
			StandbyPrivateConnection: &DatabaseConnection{
				URI:      "postgresql://user:pass@replica-private-host.com/db",
				Host:     "replica-private-host.com",
				Port:     1234,
				User:     "user",
				Password: "pass",
				SSL:      true,
				Database: "db",
			},
		},
	}

	body := `
{
  "pools": [{
    "name": "pool",
    "user": "user",
    "size": 10,
    "mode": "transaction",
    "db": "db",
    "connection": {
      "uri": "postgresql://user:pass@host.com/db",
      "host": "host.com",
      "port": 1234,
      "user": "user",
      "password": "pass",
      "database": "db",
      "ssl": true
    },
    "private_connection": {
      "uri": "postgresql://user:pass@private-host.com/db",
      "host": "private-host.com",
      "port": 1234,
      "user": "user",
      "password": "pass",
      "database": "db",
      "ssl": true
    },
    "standby_connection": {
      "uri": "postgresql://user:pass@replica-host.com/db",
      "host": "replica-host.com",
      "port": 1234,
      "user": "user",
      "password": "pass",
      "database": "db",
      "ssl": true
    },
    "standby_private_connection": {
      "uri": "postgresql://user:pass@replica-private-host.com/db",
      "host": "replica-private-host.com",
      "port": 1234,
      "user": "user",
      "password": "pass",
      "database": "db",
      "ssl": true
    }
  }]
}
`
	path := fmt.Sprintf("/v2/databases/%s/pools", dbID)

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, body)
	})

	got, _, err := client.Databases.ListPools(ctx, dbID, nil)
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestDatabases_CreatePool(t *testing.T) {
	setup()
	defer teardown()

	dbID := "deadbeef-dead-4aa5-beef-deadbeef347d"

	want := &DatabasePool{
		Name:     "pool",
		User:     "user",
		Size:     10,
		Mode:     "transaction",
		Database: "db",
		Connection: &DatabaseConnection{
			URI:      "postgresql://user:pass@host.com/db",
			Host:     "host.com",
			Port:     1234,
			User:     "user",
			Password: "pass",
			SSL:      true,
			Database: "db",
		},
		PrivateConnection: &DatabaseConnection{
			URI:      "postgresql://user:pass@private-host.com/db",
			Host:     "private-host.com",
			Port:     1234,
			User:     "user",
			Password: "pass",
			SSL:      true,
			Database: "db",
		},
		StandbyConnection: &DatabaseConnection{
			URI:      "postgresql://user:pass@replica-host.com/db",
			Host:     "replica-host.com",
			Port:     1234,
			User:     "user",
			Password: "pass",
			SSL:      true,
			Database: "db",
		},
		StandbyPrivateConnection: &DatabaseConnection{
			URI:      "postgresql://user:pass@replica-private-host.com/db",
			Host:     "replica-private-host.com",
			Port:     1234,
			User:     "user",
			Password: "pass",
			SSL:      true,
			Database: "db",
		},
	}

	body := `
{
  "pool": {
    "name": "pool",
    "user": "user",
    "size": 10,
    "mode": "transaction",
    "db": "db",
    "connection": {
      "uri": "postgresql://user:pass@host.com/db",
      "host": "host.com",
      "port": 1234,
      "user": "user",
      "password": "pass",
      "database": "db",
      "ssl": true
    },
    "private_connection": {
      "uri": "postgresql://user:pass@private-host.com/db",
      "host": "private-host.com",
      "port": 1234,
      "user": "user",
      "password": "pass",
      "database": "db",
      "ssl": true
    },
    "standby_connection": {
      "uri": "postgresql://user:pass@replica-host.com/db",
      "host": "replica-host.com",
      "port": 1234,
      "user": "user",
      "password": "pass",
      "database": "db",
      "ssl": true
    },
    "standby_private_connection": {
      "uri": "postgresql://user:pass@replica-private-host.com/db",
      "host": "replica-private-host.com",
      "port": 1234,
      "user": "user",
      "password": "pass",
      "database": "db",
      "ssl": true
    }
  }
}
`
	path := fmt.Sprintf("/v2/databases/%s/pools", dbID)

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, body)
	})

	got, _, err := client.Databases.CreatePool(ctx, dbID, &DatabaseCreatePoolRequest{
		Name:     "pool",
		Database: "db",
		Size:     10,
		User:     "foo",
		Mode:     "transaction",
	})
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestDatabases_GetPool(t *testing.T) {
	setup()
	defer teardown()

	dbID := "deadbeef-dead-4aa5-beef-deadbeef347d"

	want := &DatabasePool{
		Name:     "pool",
		User:     "user",
		Size:     10,
		Mode:     "transaction",
		Database: "db",
		Connection: &DatabaseConnection{
			URI:      "postgresql://user:pass@host.com/db",
			Host:     "host.com",
			Port:     1234,
			User:     "user",
			Password: "pass",
			SSL:      true,
			Database: "db",
		},
		PrivateConnection: &DatabaseConnection{
			URI:      "postgresql://user:pass@private-host.com/db",
			Host:     "private-host.com",
			Port:     1234,
			User:     "user",
			Password: "pass",
			SSL:      true,
			Database: "db",
		},
		StandbyConnection: &DatabaseConnection{
			URI:      "postgresql://user:pass@replica-host.com/db",
			Host:     "replica-host.com",
			Port:     1234,
			User:     "user",
			Password: "pass",
			SSL:      true,
			Database: "db",
		},
		StandbyPrivateConnection: &DatabaseConnection{
			URI:      "postgresql://user:pass@replica-private-host.com/db",
			Host:     "replica-private-host.com",
			Port:     1234,
			User:     "user",
			Password: "pass",
			SSL:      true,
			Database: "db",
		},
	}

	body := `
{
  "pool": {
    "name": "pool",
    "user": "user",
    "size": 10,
    "mode": "transaction",
    "db": "db",
    "connection": {
      "uri": "postgresql://user:pass@host.com/db",
      "host": "host.com",
      "port": 1234,
      "user": "user",
      "password": "pass",
      "database": "db",
      "ssl": true
    },
    "private_connection": {
      "uri": "postgresql://user:pass@private-host.com/db",
      "host": "private-host.com",
      "port": 1234,
      "user": "user",
      "password": "pass",
      "database": "db",
      "ssl": true
    },
    "standby_connection": {
      "uri": "postgresql://user:pass@replica-host.com/db",
      "host": "replica-host.com",
      "port": 1234,
      "user": "user",
      "password": "pass",
      "database": "db",
      "ssl": true
    },
    "standby_private_connection": {
      "uri": "postgresql://user:pass@replica-private-host.com/db",
      "host": "replica-private-host.com",
      "port": 1234,
      "user": "user",
      "password": "pass",
      "database": "db",
      "ssl": true
    }
  }
}
`
	path := fmt.Sprintf("/v2/databases/%s/pools/pool", dbID)

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, body)
	})

	got, _, err := client.Databases.GetPool(ctx, dbID, "pool")
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestDatabases_DeletePool(t *testing.T) {
	setup()
	defer teardown()

	dbID := "deadbeef-dead-4aa5-beef-deadbeef347d"

	path := fmt.Sprintf("/v2/databases/%s/pools/pool", dbID)

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
	})

	_, err := client.Databases.DeletePool(ctx, dbID, "pool")
	require.NoError(t, err)
}

func TestDatabases_UpdatePool(t *testing.T) {
	setup()
	defer teardown()

	dbID := "deadbeef-dead-4aa5-beef-deadbeef347d"

	path := fmt.Sprintf("/v2/databases/%s/pools/pool", dbID)

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
	})

	_, err := client.Databases.UpdatePool(ctx, dbID, "pool", &DatabaseUpdatePoolRequest{
		User:     "user",
		Size:     12,
		Database: "db",
		Mode:     "transaction",
	})
	require.NoError(t, err)
}

func TestDatabases_GetReplica(t *testing.T) {
	setup()
	defer teardown()

	dbID := "deadbeef-dead-4aa5-beef-deadbeef347d"

	createdAt := time.Date(2019, 01, 01, 0, 0, 0, 0, time.UTC)

	want := &DatabaseReplica{
		ID:        "326f188b-5dd1-45fc-9584-62ad553107cd",
		Name:      "pool",
		Region:    "nyc1",
		Status:    "online",
		CreatedAt: createdAt,
		Connection: &DatabaseConnection{
			URI:      "postgresql://user:pass@host.com/db",
			Host:     "host.com",
			Port:     1234,
			User:     "user",
			Password: "pass",
			SSL:      true,
			Database: "db",
		},
		PrivateConnection: &DatabaseConnection{
			URI:      "postgresql://user:pass@private-host.com/db",
			Host:     "private-host.com",
			Port:     1234,
			User:     "user",
			Password: "pass",
			SSL:      true,
			Database: "db",
		},
		PrivateNetworkUUID: "deadbeef-dead-4aa5-beef-deadbeef347d",
		Tags:               []string{"production", "staging"},
		StorageSizeMib:     51200,
		Size:               "db-s-1vcpu-1gb",
	}

	body := `
{
  "replica": {
    "id": "326f188b-5dd1-45fc-9584-62ad553107cd",
    "name": "pool",
    "region": "nyc1",
    "status": "online",
    "created_at": "` + createdAt.Format(time.RFC3339) + `",
    "connection": {
      "uri": "postgresql://user:pass@host.com/db",
      "host": "host.com",
      "port": 1234,
      "user": "user",
      "password": "pass",
      "database": "db",
      "ssl": true
    },
    "private_connection": {
      "uri": "postgresql://user:pass@private-host.com/db",
      "host": "private-host.com",
      "port": 1234,
      "user": "user",
      "password": "pass",
      "database": "db",
      "ssl": true
    },
    "private_network_uuid": "deadbeef-dead-4aa5-beef-deadbeef347d",
	"tags": ["production", "staging"],
	"storage_size_mib": 51200,
	"size": "db-s-1vcpu-1gb"
  }
}
`
	path := fmt.Sprintf("/v2/databases/%s/replicas/replica", dbID)

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, body)
	})

	got, _, err := client.Databases.GetReplica(ctx, dbID, "replica")
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestDatabases_ListReplicas(t *testing.T) {
	setup()
	defer teardown()

	dbID := "deadbeef-dead-4aa5-beef-deadbeef347d"

	createdAt := time.Date(2019, 01, 01, 0, 0, 0, 0, time.UTC)

	want := []DatabaseReplica{
		{
			Name:      "pool",
			Region:    "nyc1",
			Status:    "online",
			CreatedAt: createdAt,
			Connection: &DatabaseConnection{
				URI:      "postgresql://user:pass@host.com/db",
				Host:     "host.com",
				Port:     1234,
				User:     "user",
				Password: "pass",
				SSL:      true,
				Database: "db",
			},
			PrivateConnection: &DatabaseConnection{
				URI:      "postgresql://user:pass@private-host.com/db",
				Host:     "private-host.com",
				Port:     1234,
				User:     "user",
				Password: "pass",
				SSL:      true,
				Database: "db",
			},
			PrivateNetworkUUID: "deadbeef-dead-4aa5-beef-deadbeef347d",
			Tags:               []string{"production", "staging"},
			StorageSizeMib:     51200,
			Size:               "db-s-1vcpu-1gb",
		},
	}

	body := `
{
  "replicas": [{
    "name": "pool",
    "region": "nyc1",
    "status": "online",
    "created_at": "` + createdAt.Format(time.RFC3339) + `",
    "connection": {
      "uri": "postgresql://user:pass@host.com/db",
      "host": "host.com",
      "port": 1234,
      "user": "user",
      "password": "pass",
      "database": "db",
      "ssl": true
    },
    "private_connection": {
      "uri": "postgresql://user:pass@private-host.com/db",
      "host": "private-host.com",
      "port": 1234,
      "user": "user",
      "password": "pass",
      "database": "db",
      "ssl": true
    },
    "private_network_uuid": "deadbeef-dead-4aa5-beef-deadbeef347d",
	"tags": ["production", "staging"],
	"storage_size_mib": 51200,
	"size": "db-s-1vcpu-1gb"
  }]
}
`
	path := fmt.Sprintf("/v2/databases/%s/replicas", dbID)

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, body)
	})

	got, _, err := client.Databases.ListReplicas(ctx, dbID, nil)
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestDatabases_CreateReplica(t *testing.T) {
	setup()
	defer teardown()

	dbID := "deadbeef-dead-4aa5-beef-deadbeef347d"

	createdAt := time.Date(2019, 01, 01, 0, 0, 0, 0, time.UTC)

	want := &DatabaseReplica{
		Name:      "pool",
		Region:    "nyc1",
		Status:    "online",
		CreatedAt: createdAt,
		Connection: &DatabaseConnection{
			URI:      "postgresql://user:pass@host.com/db",
			Host:     "host.com",
			Port:     1234,
			User:     "user",
			Password: "pass",
			SSL:      true,
			Database: "db",
		},
		PrivateConnection: &DatabaseConnection{
			URI:      "postgresql://user:pass@private-host.com/db",
			Host:     "private-host.com",
			Port:     1234,
			User:     "user",
			Password: "pass",
			SSL:      true,
			Database: "db",
		},
		PrivateNetworkUUID: "deadbeef-dead-4aa5-beef-deadbeef347d",
		Tags:               []string{"production", "staging"},
		StorageSizeMib:     51200,
		Size:               "db-s-2vcpu-4gb",
	}

	body := `
{
  "replica": {
    "name": "pool",
    "region": "nyc1",
    "status": "online",
    "created_at": "` + createdAt.Format(time.RFC3339) + `",
    "connection": {
      "uri": "postgresql://user:pass@host.com/db",
      "host": "host.com",
      "port": 1234,
      "user": "user",
      "password": "pass",
      "database": "db",
      "ssl": true
    },
    "private_connection": {
      "uri": "postgresql://user:pass@private-host.com/db",
      "host": "private-host.com",
      "port": 1234,
      "user": "user",
      "password": "pass",
      "database": "db",
      "ssl": true
    },
    "private_network_uuid": "deadbeef-dead-4aa5-beef-deadbeef347d",
	"tags": ["production", "staging"],
	"storage_size_mib": 51200,
	"size": "db-s-2vcpu-4gb"
  }
}
`
	path := fmt.Sprintf("/v2/databases/%s/replicas", dbID)

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, body)
	})

	got, _, err := client.Databases.CreateReplica(ctx, dbID, &DatabaseCreateReplicaRequest{
		Name:               "replica",
		Region:             "nyc1",
		Size:               "db-s-2vcpu-4gb",
		PrivateNetworkUUID: privateNetworkUUID,
		Tags:               []string{"production", "staging"},
		StorageSizeMib:     uint64(51200),
	})
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestDatabases_PromoteReplicaToPrimary(t *testing.T) {
	setup()
	defer teardown()

	dbID := "deadbeef-dead-4aa5-beef-deadbeef347d"

	path := fmt.Sprintf("/v2/databases/%s/replicas/replica/promote", dbID)

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
	})

	_, err := client.Databases.PromoteReplicaToPrimary(ctx, dbID, "replica")
	require.NoError(t, err)
}

func TestDatabases_DeleteReplica(t *testing.T) {
	setup()
	defer teardown()

	dbID := "deadbeef-dead-4aa5-beef-deadbeef347d"

	path := fmt.Sprintf("/v2/databases/%s/replicas/replica", dbID)

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
	})

	_, err := client.Databases.DeleteReplica(ctx, dbID, "replica")
	require.NoError(t, err)
}

func TestDatabases_SetEvictionPolicy(t *testing.T) {
	setup()
	defer teardown()

	dbID := "deadbeef-dead-4aa5-beef-deadbeef347d"

	path := fmt.Sprintf("/v2/databases/%s/eviction_policy", dbID)

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
	})

	_, err := client.Databases.SetEvictionPolicy(ctx, dbID, "allkeys_lru")
	require.NoError(t, err)
}

func TestDatabases_GetEvictionPolicy(t *testing.T) {
	setup()
	defer teardown()

	dbID := "deadbeef-dead-4aa5-beef-deadbeef347d"

	want := "allkeys_lru"

	body := `{ "eviction_policy": "allkeys_lru" }`

	path := fmt.Sprintf("/v2/databases/%s/eviction_policy", dbID)

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, body)
	})

	got, _, err := client.Databases.GetEvictionPolicy(ctx, dbID)
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestDatabases_SetSQLMode(t *testing.T) {
	setup()
	defer teardown()

	dbID := "deadbeef-dead-4aa5-beef-deadbeef347d"

	path := fmt.Sprintf("/v2/databases/%s/sql_mode", dbID)

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		w.Write([]byte(`{ "sql_mode": "ONLY_FULL_GROUP_BY" }`))
	})

	_, err := client.Databases.SetSQLMode(ctx, dbID, "ONLY_FULL_GROUP_BY")
	require.NoError(t, err)
}

func TestDatabases_SetSQLMode_Multiple(t *testing.T) {
	setup()
	defer teardown()

	dbID := "deadbeef-dead-4aa5-beef-deadbeef347d"

	path := fmt.Sprintf("/v2/databases/%s/sql_mode", dbID)

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		w.Write([]byte(`{ "sql_mode": "ANSI, ANSI_QUOTES" }`))
	})

	_, err := client.Databases.SetSQLMode(ctx, dbID, SQLModeANSI, SQLModeANSIQuotes)
	require.NoError(t, err)
}

func TestDatabases_GetSQLMode(t *testing.T) {
	setup()
	defer teardown()

	dbID := "deadbeef-dead-4aa5-beef-deadbeef347d"

	want := "ONLY_FULL_GROUP_BY"

	body := `{ "sql_mode": "ONLY_FULL_GROUP_BY" }`

	path := fmt.Sprintf("/v2/databases/%s/sql_mode", dbID)

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, body)
	})

	got, _, err := client.Databases.GetSQLMode(ctx, dbID)
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestDatabases_GetFirewallRules(t *testing.T) {
	setup()
	defer teardown()

	dbID := "deadbeef-dead-4aa5-beef-deadbeef347d"

	path := fmt.Sprintf("/v2/databases/%s/firewall", dbID)

	want := []DatabaseFirewallRule{
		{
			Type:        "ip_addr",
			Value:       "192.168.1.1",
			UUID:        "deadbeef-dead-4aa5-beef-deadbeef347d",
			ClusterUUID: "deadbeef-dead-4aa5-beef-deadbeef347d",
		},
	}

	body := ` {"rules": [{
		"type": "ip_addr",
		"value": "192.168.1.1",
		"uuid": "deadbeef-dead-4aa5-beef-deadbeef347d",
		"cluster_uuid": "deadbeef-dead-4aa5-beef-deadbeef347d"
	}]} `

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, body)
	})

	got, _, err := client.Databases.GetFirewallRules(ctx, dbID)
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestDatabases_UpdateFirewallRules(t *testing.T) {
	setup()
	defer teardown()

	dbID := "deadbeef-dead-4aa5-beef-deadbeef347d"

	path := fmt.Sprintf("/v2/databases/%s/firewall", dbID)

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
	})

	_, err := client.Databases.UpdateFirewallRules(ctx, dbID, &DatabaseUpdateFirewallRulesRequest{
		Rules: []*DatabaseFirewallRule{
			{
				Type:  "ip_addr",
				Value: "192.168.1.1",
				UUID:  "deadbeef-dead-4aa5-beef-deadbeef347d",
			},
		},
	})
	require.NoError(t, err)
}

func TestDatabases_GetDatabaseOptions(t *testing.T) {
	setup()
	defer teardown()

	path := "/v2/databases/options"

	body := ` {
		"options": {
			"mongodb": {
				"regions": [
					"ams3",
					"blr1"
				],
				"versions": [
					"4.4",
					"5.0"
				],
				"layouts": [
					{
						"num_nodes": 1,
						"sizes": [
							"db-s-1vcpu-1gb",
							"db-s-1vcpu-2gb"
						]
					},
					{
						"num_nodes": 3,
						"sizes": [
							"so1_5-4vcpu-32gb",
							"so1_5-32vcpu-256gb"
						]
					}
				]
			},
			"mysql": {
				"regions": [
					"ams3",
					"sgp1",
					"tor1"
				],
				"versions": [
					"8"
				],
				"layouts": [
					{
						"num_nodes": 1,
						"sizes": [
							"db-s-1vcpu-1gb",
							"db-s-1vcpu-2gb"
						]
					},
					{
						"num_nodes": 2,
						"sizes": [
							"db-s-1vcpu-2gb",
							"so1_5-32vcpu-256gb"
						]
					},
					{
						"num_nodes": 3,
						"sizes": [
							"db-s-1vcpu-2gb",
							"so1_5-32vcpu-256gb"
						]
					}
				]
			},
			"pg": {
				"regions": [
					"ams3",
					"blr1"
				],
				"versions": [
					"13",
					"14"
				],
				"layouts": [
					{
						"num_nodes": 1,
						"sizes": [
							"db-s-1vcpu-1gb",
							"db-s-1vcpu-2gb"
						]
					},
					{
						"num_nodes": 2,
						"sizes": [
							"db-s-1vcpu-2gb",
							"db-s-2vcpu-4gb"
						]
					},
					{
						"num_nodes": 3,
						"sizes": [
							"db-s-1vcpu-2gb",
							"db-s-2vcpu-4gb"
						]
					}
				]
			},
			"kafka": {
				"regions": [
					"ams3",
					"tor1"
				],
				"versions": [
					"3.3"
				],
				"layouts": [
					{
						"num_nodes": 3,
						"sizes": [
							"gd-2vcpu-8gb",
	            			"gd-4vcpu-16gb"
						]
					}
				]
			},
			"opensearch": {
				"regions": [
					"ams3",
					"tor1"
				],
				"versions": [
					"1",
					"2"
				],
				"layouts": [
					{
						"num_nodes": 1,
						"sizes": [
							"db-s-2vcpu-4gb",
							"db-s-4vcpu-8gb"
						]
					},
					{
						"num_nodes": 3,
						"sizes": [
							"db-s-2vcpu-4gb",
							"m3-2vcpu-16gb",
							"db-s-4vcpu-8gb",
							"m3-4vcpu-32gb"
						]
					},
					{
						"num_nodes": 6,
						"sizes": [
							"m3-2vcpu-16gb",
							"m3-4vcpu-32gb"
						]
					},
					{
						"num_nodes": 9,
						"sizes": [
							"m3-2vcpu-16gb",
							"m3-4vcpu-32gb"
						]
					},
					{
						"num_nodes": 15,
						"sizes": [
							"m3-2vcpu-16gb",
							"m3-4vcpu-32gb"
						]
					}
				]
			},
			"redis": {
				"regions": [
					"ams3",
					"tor1"
				],
				"versions": [
					"6"
				],
				"layouts": [
					{
						"num_nodes": 1,
						"sizes": [
							"m-32vcpu-256gb"
						]
					},
					{
						"num_nodes": 2,
						"sizes": [
							"db-s-1vcpu-2gb",
							"db-s-2vcpu-4gb",
							"m-32vcpu-256gb"
						]
					}
				]
			},
			"valkey": {
				"regions": [
					"ams3",
					"tor1"
				],
				"versions": [
					"8"
				],
				"layouts": [
					{
						"num_nodes": 1,
						"sizes": [
							"m-32vcpu-256gb"
						]
					},
					{
						"num_nodes": 2,
						"sizes": [
							"db-s-1vcpu-2gb",
							"db-s-2vcpu-4gb",
							"m-32vcpu-256gb"
						]
					}
				]
			}
		}
	} `

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, body)
	})

	options, _, err := client.Databases.ListOptions(ctx)
	require.NoError(t, err)
	require.NotNil(t, options)
	require.NotNil(t, options.MongoDBOptions)
	require.NotNil(t, options.PostgresSQLOptions)
	require.NotNil(t, options.RedisOptions)
	require.NotNil(t, options.ValkeyOptions)
	require.NotNil(t, options.MySQLOptions)
	require.NotNil(t, options.KafkaOptions)
	require.NotNil(t, options.OpensearchOptions)
	require.Greater(t, len(options.MongoDBOptions.Regions), 0)
	require.Greater(t, len(options.PostgresSQLOptions.Regions), 0)
	require.Greater(t, len(options.RedisOptions.Regions), 0)
	require.Greater(t, len(options.ValkeyOptions.Regions), 0)
	require.Greater(t, len(options.MySQLOptions.Regions), 0)
	require.Greater(t, len(options.KafkaOptions.Regions), 0)
	require.Greater(t, len(options.OpensearchOptions.Regions), 0)
	require.Greater(t, len(options.MongoDBOptions.Versions), 0)
	require.Greater(t, len(options.PostgresSQLOptions.Versions), 0)
	require.Greater(t, len(options.RedisOptions.Versions), 0)
	require.Greater(t, len(options.ValkeyOptions.Versions), 0)
	require.Greater(t, len(options.MySQLOptions.Versions), 0)
	require.Greater(t, len(options.KafkaOptions.Versions), 0)
	require.Greater(t, len(options.OpensearchOptions.Versions), 0)
	require.Greater(t, len(options.MongoDBOptions.Layouts), 0)
	require.Greater(t, len(options.PostgresSQLOptions.Layouts), 0)
	require.Greater(t, len(options.RedisOptions.Layouts), 0)
	require.Greater(t, len(options.ValkeyOptions.Layouts), 0)
	require.Greater(t, len(options.MySQLOptions.Layouts), 0)
	require.Greater(t, len(options.KafkaOptions.Layouts), 0)
	require.Greater(t, len(options.OpensearchOptions.Layouts), 0)
}

func TestDatabases_CreateDatabaseUserWithMySQLSettings(t *testing.T) {
	setup()
	defer teardown()

	dbID := "deadbeef-dead-4aa5-beef-deadbeef347d"

	path := fmt.Sprintf("/v2/databases/%s/users", dbID)

	responseJSON := []byte(fmt.Sprintf(`{
		"user": {
			"name": "foo",
			"mysql_settings": {
				"auth_plugin": "%s"
			}
		}
	}`, SQLAuthPluginNative))
	expectedUser := &DatabaseUser{
		Name: "foo",
		MySQLSettings: &DatabaseMySQLUserSettings{
			AuthPlugin: SQLAuthPluginNative,
		},
	}

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		w.WriteHeader(http.StatusOK)
		w.Write(responseJSON)
	})

	user, _, err := client.Databases.CreateUser(ctx, dbID, &DatabaseCreateUserRequest{
		Name:          expectedUser.Name,
		MySQLSettings: expectedUser.MySQLSettings,
	})
	require.NoError(t, err)
	require.Equal(t, expectedUser, user)
}

func TestDatabases_ListDatabaseUsersWithMySQLSettings(t *testing.T) {
	setup()
	defer teardown()

	dbID := "deadbeef-dead-4aa5-beef-deadbeef347d"

	path := fmt.Sprintf("/v2/databases/%s/users", dbID)

	responseJSON := []byte(fmt.Sprintf(`{
		"users": [
			{
				"name": "foo",
				"mysql_settings": {
					"auth_plugin": "%s"
				}
			}
		]
	}`, SQLAuthPluginNative))
	expectedUsers := []DatabaseUser{
		{
			Name: "foo",
			MySQLSettings: &DatabaseMySQLUserSettings{
				AuthPlugin: SQLAuthPluginNative,
			},
		},
	}

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.WriteHeader(http.StatusOK)
		w.Write(responseJSON)
	})

	users, _, err := client.Databases.ListUsers(ctx, dbID, &ListOptions{})
	require.NoError(t, err)
	require.Equal(t, expectedUsers, users)
}

func TestDatabases_GetDatabaseUserWithMySQLSettings(t *testing.T) {
	setup()
	defer teardown()

	dbID := "deadbeef-dead-4aa5-beef-deadbeef347d"
	userID := "d290a0a0-27da-42bd-a4b2-bcecf43b8832"

	path := fmt.Sprintf("/v2/databases/%s/users/%s", dbID, userID)

	responseJSON := []byte(fmt.Sprintf(`{
		"user": {
			"name": "foo",
			"mysql_settings": {
				"auth_plugin": "%s"
			}
		}
	}`, SQLAuthPluginNative))
	expectedUser := &DatabaseUser{
		Name: "foo",
		MySQLSettings: &DatabaseMySQLUserSettings{
			AuthPlugin: SQLAuthPluginNative,
		},
	}

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.WriteHeader(http.StatusOK)
		w.Write(responseJSON)
	})

	user, _, err := client.Databases.GetUser(ctx, dbID, userID)
	require.NoError(t, err)
	require.Equal(t, expectedUser, user)
}

func TestDatabases_CreateDatabaseUserWithKafkaSettings(t *testing.T) {
	setup()
	defer teardown()

	dbID := "deadbeef-dead-4aa5-beef-deadbeef347d"
	path := fmt.Sprintf("/v2/databases/%s/users", dbID)

	writeTopicACL := []*KafkaACL{
		{
			ID:         "1",
			Topic:      "bar",
			Permission: "write",
		},
	}

	acljson, err := json.Marshal(writeTopicACL)
	if err != nil {
		t.Fatal(err)
	}

	responseJSON := []byte(fmt.Sprintf(`{
		"user": {
			"name": "foo",
			"settings": {
				"acl": %s
			}
		}
	}`, string(acljson)))

	expectedUser := &DatabaseUser{
		Name: "foo",
		Settings: &DatabaseUserSettings{
			ACL: writeTopicACL,
		},
	}

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		w.WriteHeader(http.StatusOK)
		w.Write(responseJSON)
	})

	user, _, err := client.Databases.CreateUser(ctx, dbID, &DatabaseCreateUserRequest{
		Name:     expectedUser.Name,
		Settings: &DatabaseUserSettings{ACL: expectedUser.Settings.ACL},
	})
	require.NoError(t, err)
	require.Equal(t, expectedUser, user)
}

func TestDatabases_CreateDatabaseUserWithOpenSearchSettings(t *testing.T) {
	setup()
	defer teardown()

	dbID := "deadbeef-dead-4aa5-beef-deadbeef347d"
	path := fmt.Sprintf("/v2/databases/%s/users", dbID)

	writeIndexACL := []*OpenSearchACL{
		{
			Index:      "bar",
			Permission: "write",
		},
	}

	acljson, err := json.Marshal(writeIndexACL)
	if err != nil {
		t.Fatal(err)
	}

	responseJSON := []byte(fmt.Sprintf(`{
		"user": {
			"name": "foo",
			"settings": {
				"opensearch_acl": %s
			}
		}
	}`, string(acljson)))

	expectedUser := &DatabaseUser{
		Name: "foo",
		Settings: &DatabaseUserSettings{
			OpenSearchACL: writeIndexACL,
		},
	}

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		w.WriteHeader(http.StatusOK)
		w.Write(responseJSON)
	})

	user, _, err := client.Databases.CreateUser(ctx, dbID, &DatabaseCreateUserRequest{
		Name:     expectedUser.Name,
		Settings: &DatabaseUserSettings{OpenSearchACL: expectedUser.Settings.OpenSearchACL},
	})
	require.NoError(t, err)
	require.Equal(t, expectedUser, user)
}

func TestDatabases_ListDatabaseUsersWithKafkaSettings(t *testing.T) {
	setup()
	defer teardown()

	dbID := "deadbeef-dead-4aa5-beef-deadbeef347d"

	path := fmt.Sprintf("/v2/databases/%s/users", dbID)

	writeTopicACL := []*KafkaACL{
		{
			ID:         "1",
			Topic:      "bar",
			Permission: "write",
		},
	}

	acljson, err := json.Marshal(writeTopicACL)
	if err != nil {
		t.Fatal(err)
	}

	responseJSON := []byte(fmt.Sprintf(`{
		"users": [
			{
				"name": "foo",
				"settings": {
					"acl": %s
				}
			}
		]
	}`, string(acljson)))

	expectedUsers := []DatabaseUser{
		{
			Name: "foo",
			Settings: &DatabaseUserSettings{
				ACL: writeTopicACL,
			},
		},
	}

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.WriteHeader(http.StatusOK)
		w.Write(responseJSON)
	})

	users, _, err := client.Databases.ListUsers(ctx, dbID, &ListOptions{})
	require.NoError(t, err)
	require.Equal(t, expectedUsers, users)
}

func TestDatabases_ListDatabaseUsersWithOpenSearchSettings(t *testing.T) {
	setup()
	defer teardown()

	dbID := "deadbeef-dead-4aa5-beef-deadbeef347d"

	path := fmt.Sprintf("/v2/databases/%s/users", dbID)

	writeIndexACL := []*OpenSearchACL{
		{
			Index:      "bar",
			Permission: "write",
		},
	}

	acljson, err := json.Marshal(writeIndexACL)
	if err != nil {
		t.Fatal(err)
	}

	responseJSON := []byte(fmt.Sprintf(`{
		"users": [
			{
				"name": "foo",
				"settings": {
					"opensearch_acl": %s
				}
			}
		]
	}`, string(acljson)))

	expectedUsers := []DatabaseUser{
		{
			Name: "foo",
			Settings: &DatabaseUserSettings{
				OpenSearchACL: writeIndexACL,
			},
		},
	}

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.WriteHeader(http.StatusOK)
		w.Write(responseJSON)
	})

	users, _, err := client.Databases.ListUsers(ctx, dbID, &ListOptions{})
	require.NoError(t, err)
	require.Equal(t, expectedUsers, users)
}

func TestDatabases_GetDatabaseUserWithOpenSearchSettings(t *testing.T) {
	setup()
	defer teardown()

	dbID := "deadbeef-dead-4aa5-beef-deadbeef347d"
	userID := "d290a0a0-27da-42bd-a4b2-bcecf43b8832"

	path := fmt.Sprintf("/v2/databases/%s/users/%s", dbID, userID)

	writeIndexACL := []*OpenSearchACL{
		{
			Index:      "bar",
			Permission: "write",
		},
	}

	acljson, err := json.Marshal(writeIndexACL)
	if err != nil {
		t.Fatal(err)
	}

	responseJSON := []byte(fmt.Sprintf(`{
		"user": {
			"name": "foo",
			"settings": {
				"opensearch_acl": %s
			}
		}
	}`, string(acljson)))

	expectedUser := &DatabaseUser{
		Name: "foo",
		Settings: &DatabaseUserSettings{
			OpenSearchACL: writeIndexACL,
		},
	}

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.WriteHeader(http.StatusOK)
		w.Write(responseJSON)
	})

	user, _, err := client.Databases.GetUser(ctx, dbID, userID)
	require.NoError(t, err)
	require.Equal(t, expectedUser, user)
}

func TestDatabases_GetDatabaseUserWithKafkaSettings(t *testing.T) {
	setup()
	defer teardown()

	dbID := "deadbeef-dead-4aa5-beef-deadbeef347d"
	userID := "d290a0a0-27da-42bd-a4b2-bcecf43b8832"

	path := fmt.Sprintf("/v2/databases/%s/users/%s", dbID, userID)

	writeTopicACL := []*KafkaACL{
		{
			ID:         "1",
			Topic:      "bar",
			Permission: "write",
		},
	}

	acljson, err := json.Marshal(writeTopicACL)
	if err != nil {
		t.Fatal(err)
	}

	responseJSON := []byte(fmt.Sprintf(`{
		"user": {
			"name": "foo",
			"settings": {
				"acl": %s
			}
		}
	}`, string(acljson)))

	expectedUser := &DatabaseUser{
		Name: "foo",
		Settings: &DatabaseUserSettings{
			ACL: writeTopicACL,
		},
	}

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.WriteHeader(http.StatusOK)
		w.Write(responseJSON)
	})

	user, _, err := client.Databases.GetUser(ctx, dbID, userID)
	require.NoError(t, err)
	require.Equal(t, expectedUser, user)
}

func TestDatabases_GetConfigPostgres(t *testing.T) {
	setup()
	defer teardown()

	var (
		dbSvc = client.Databases
		dbID  = "da4e0206-d019-41d7-b51f-deadbeefbb8f"
		path  = fmt.Sprintf("/v2/databases/%s/config", dbID)

		postgresConfigJSON = `{
  "config": {
    "autovacuum_naptime": 60,
    "autovacuum_vacuum_threshold": 50,
    "autovacuum_analyze_threshold": 50,
    "autovacuum_vacuum_scale_factor": 0.2,
    "autovacuum_analyze_scale_factor": 0.2,
    "autovacuum_vacuum_cost_delay": 20,
    "autovacuum_vacuum_cost_limit": -1,
    "bgwriter_flush_after": 512,
    "bgwriter_lru_maxpages": 100,
    "bgwriter_lru_multiplier": 2,
    "idle_in_transaction_session_timeout": 0,
    "jit": true,
    "log_autovacuum_min_duration": -1,
    "log_min_duration_statement": -1,
    "max_prepared_transactions": 0,
    "max_parallel_workers": 8,
    "max_parallel_workers_per_gather": 2,
    "temp_file_limit": -1,
    "wal_sender_timeout": 60000,
    "backup_hour": 18,
    "backup_minute": 26
  }
}`

		postgresConfig = PostgreSQLConfig{
			AutovacuumNaptime:               PtrTo(60),
			AutovacuumVacuumThreshold:       PtrTo(50),
			AutovacuumAnalyzeThreshold:      PtrTo(50),
			AutovacuumVacuumScaleFactor:     PtrTo(float32(0.2)),
			AutovacuumAnalyzeScaleFactor:    PtrTo(float32(0.2)),
			AutovacuumVacuumCostDelay:       PtrTo(20),
			AutovacuumVacuumCostLimit:       PtrTo(-1),
			BGWriterFlushAfter:              PtrTo(512),
			BGWriterLRUMaxpages:             PtrTo(100),
			BGWriterLRUMultiplier:           PtrTo(float32(2)),
			IdleInTransactionSessionTimeout: PtrTo(0),
			JIT:                             PtrTo(true),
			LogAutovacuumMinDuration:        PtrTo(-1),
			LogMinDurationStatement:         PtrTo(-1),
			MaxPreparedTransactions:         PtrTo(0),
			MaxParallelWorkers:              PtrTo(8),
			MaxParallelWorkersPerGather:     PtrTo(2),
			TempFileLimit:                   PtrTo(-1),
			WalSenderTimeout:                PtrTo(60000),
			BackupHour:                      PtrTo(18),
			BackupMinute:                    PtrTo(26),
		}
	)

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, postgresConfigJSON)
	})

	got, _, err := dbSvc.GetPostgreSQLConfig(ctx, dbID)
	require.NoError(t, err)
	require.Equal(t, &postgresConfig, got)
}

func TestDatabases_UpdateConfigPostgres(t *testing.T) {
	setup()
	defer teardown()

	var (
		dbID           = "deadbeef-dead-4aa5-beef-deadbeef347d"
		path           = fmt.Sprintf("/v2/databases/%s/config", dbID)
		postgresConfig = &PostgreSQLConfig{
			AutovacuumNaptime:          PtrTo(75),
			AutovacuumVacuumThreshold:  PtrTo(45),
			AutovacuumAnalyzeThreshold: PtrTo(45),
			MaxPreparedTransactions:    PtrTo(0),
		}
	)

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPatch)

		var b databasePostgreSQLConfigRoot
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&b)
		require.NoError(t, err)

		assert.Equal(t, b.Config, postgresConfig)
		assert.Equal(t, 0, *b.Config.MaxPreparedTransactions, "pointers to zero value should be sent")
		assert.Nil(t, b.Config.MaxParallelWorkers, "excluded value should not be sent")

		w.WriteHeader(http.StatusNoContent)
	})

	_, err := client.Databases.UpdatePostgreSQLConfig(ctx, dbID, postgresConfig)
	require.NoError(t, err)
}

func TestDatabases_GetConfigRedis(t *testing.T) {
	setup()
	defer teardown()

	var (
		dbSvc = client.Databases
		dbID  = "da4e0206-d019-41d7-b51f-deadbeefbb8f"
		path  = fmt.Sprintf("/v2/databases/%s/config", dbID)

		redisConfigJSON = `{
  "config": {
    "redis_maxmemory_policy": "allkeys-lru",
    "redis_lfu_log_factor": 10,
    "redis_lfu_decay_time": 1,
    "redis_ssl": true,
    "redis_timeout": 300,
    "redis_notify_keyspace_events": "",
    "redis_persistence": "off",
    "redis_acl_channels_default": "allchannels"
  }
}`

		redisConfig = RedisConfig{
			RedisMaxmemoryPolicy:      PtrTo("allkeys-lru"),
			RedisLFULogFactor:         PtrTo(10),
			RedisLFUDecayTime:         PtrTo(1),
			RedisSSL:                  PtrTo(true),
			RedisTimeout:              PtrTo(300),
			RedisNotifyKeyspaceEvents: PtrTo(""),
			RedisPersistence:          PtrTo("off"),
			RedisACLChannelsDefault:   PtrTo("allchannels"),
		}
	)

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, redisConfigJSON)
	})

	got, _, err := dbSvc.GetRedisConfig(ctx, dbID)
	require.NoError(t, err)
	require.Equal(t, &redisConfig, got)
}

func TestDatabases_UpdateConfigRedis(t *testing.T) {
	setup()
	defer teardown()

	var (
		dbID        = "deadbeef-dead-4aa5-beef-deadbeef347d"
		path        = fmt.Sprintf("/v2/databases/%s/config", dbID)
		redisConfig = &RedisConfig{
			RedisMaxmemoryPolicy:      PtrTo("allkeys-lru"),
			RedisLFULogFactor:         PtrTo(10),
			RedisNotifyKeyspaceEvents: PtrTo(""),
		}
	)

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPatch)

		var b databaseRedisConfigRoot
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&b)
		require.NoError(t, err)

		assert.Equal(t, b.Config, redisConfig)
		assert.Equal(t, "", *b.Config.RedisNotifyKeyspaceEvents, "pointers to zero value should be sent")
		assert.Nil(t, b.Config.RedisPersistence, "excluded value should not be sent")

		w.WriteHeader(http.StatusNoContent)
	})

	_, err := client.Databases.UpdateRedisConfig(ctx, dbID, redisConfig)
	require.NoError(t, err)
}

func TestDatabases_UpdateConfigRedisNormalizeEvictionPolicy(t *testing.T) {
	type test struct {
		input string
		want  string
	}

	tests := []test{
		{input: EvictionPolicyAllKeysLRU, want: "allkeys-lru"},
		{input: EvictionPolicyAllKeysRandom, want: "allkeys-random"},
		{input: EvictionPolicyVolatileLRU, want: "volatile-lru"},
		{input: EvictionPolicyVolatileRandom, want: "volatile-random"},
		{input: EvictionPolicyVolatileTTL, want: "volatile-ttl"},
		{input: "allkeys-lru", want: "allkeys-lru"},
		{input: "allkeys-random", want: "allkeys-random"},
		{input: "volatile-lru", want: "volatile-lru"},
		{input: "volatile-random", want: "volatile-random"},
		{input: "volatile-ttl", want: "volatile-ttl"},
		{input: "some_unknown_value", want: "some_unknown_value"},
	}

	for _, tt := range tests {
		setup()
		defer teardown()

		var (
			dbID        = "deadbeef-dead-4aa5-beef-deadbeef347d"
			path        = fmt.Sprintf("/v2/databases/%s/config", dbID)
			redisConfig = &RedisConfig{
				RedisMaxmemoryPolicy: PtrTo(tt.input),
			}
		)

		mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
			testMethod(t, r, http.MethodPatch)

			var b databaseRedisConfigRoot
			decoder := json.NewDecoder(r.Body)
			err := decoder.Decode(&b)
			require.NoError(t, err)
			assert.Equal(t, tt.want, *b.Config.RedisMaxmemoryPolicy)

			w.WriteHeader(http.StatusNoContent)
		})

		_, err := client.Databases.UpdateRedisConfig(ctx, dbID, redisConfig)
		require.NoError(t, err)
	}
}

func TestDatabases_GetConfigValkey(t *testing.T) {
	setup()
	defer teardown()

	var (
		dbSvc = client.Databases
		dbID  = "da4e0206-d019-41d7-b51f-deadbeefbb8f"
		path  = fmt.Sprintf("/v2/databases/%s/config", dbID)

		valkeyConfigJSON = `{
  "config": {
	"valkey_maxmemory_policy": "allkeys-lru",
	"valkey_lfu_log_factor": 10,
	"valkey_lfu_decay_time": 1,
	"valkey_ssl": true,
	"valkey_timeout": 300,
	"valkey_notify_keyspace_events": "",
	"valkey_persistence": "off",
	"valkey_acl_channels_default": "allchannels",
	"valkey_number_of_databases": 16
  }
}`

		valkeyConfig = ValkeyConfig{
			ValkeyMaxmemoryPolicy:      PtrTo("allkeys-lru"),
			ValkeyLFULogFactor:         PtrTo(10),
			ValkeyLFUDecayTime:         PtrTo(1),
			ValkeySSL:                  PtrTo(true),
			ValkeyTimeout:              PtrTo(300),
			ValkeyNotifyKeyspaceEvents: PtrTo(""),
			ValkeyPersistence:          PtrTo("off"),
			ValkeyACLChannelsDefault:   PtrTo("allchannels"),
			ValkeyNumberOfDatabases:    PtrTo(16),
		}
	)

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, valkeyConfigJSON)
	})

	got, _, err := dbSvc.GetValkeyConfig(ctx, dbID)
	require.NoError(t, err)
	require.Equal(t, &valkeyConfig, got)
}

func TestDatabases_UpdateConfigValkey(t *testing.T) {
	setup()
	defer teardown()

	var (
		dbID         = "deadbeef-dead-4aa5-beef-deadbeef347d"
		path         = fmt.Sprintf("/v2/databases/%s/config", dbID)
		valkeyConfig = &ValkeyConfig{
			ValkeyMaxmemoryPolicy:      PtrTo("allkeys-lru"),
			ValkeyLFULogFactor:         PtrTo(10),
			ValkeyNotifyKeyspaceEvents: PtrTo(""),
		}
	)

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPatch)

		var b databaseValkeyConfigRoot
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&b)
		require.NoError(t, err)

		assert.Equal(t, b.Config, valkeyConfig)
		assert.Equal(t, "", *b.Config.ValkeyNotifyKeyspaceEvents, "pointers to zero value should be sent")
		assert.Nil(t, b.Config.ValkeyPersistence, "excluded value should not be sent")

		w.WriteHeader(http.StatusNoContent)
	})

	_, err := client.Databases.UpdateValkeyConfig(ctx, dbID, valkeyConfig)
	require.NoError(t, err)
}

func TestDatabases_GetConfigMySQL(t *testing.T) {
	setup()
	defer teardown()

	var (
		dbSvc = client.Databases
		dbID  = "da4e0206-d019-41d7-b51f-deadbeefbb8f"
		path  = fmt.Sprintf("/v2/databases/%s/config", dbID)

		mySQLConfigJSON = `{
  "config": {
    "sql_mode": "ANSI,ERROR_FOR_DIVISION_BY_ZERO,NO_ENGINE_SUBSTITUTION,NO_ZERO_DATE,NO_ZERO_IN_DATE,STRICT_ALL_TABLES",
    "sql_require_primary_key": true,
    "innodb_ft_min_token_size": 3,
    "innodb_ft_server_stopword_table": "",
    "innodb_print_all_deadlocks": false,
    "innodb_rollback_on_timeout": false,
    "slow_query_log": false,
    "long_query_time": 10,
    "backup_hour": 21,
    "backup_minute": 59
  }
}`

		mySQLConfig = MySQLConfig{
			SQLMode:                     PtrTo("ANSI,ERROR_FOR_DIVISION_BY_ZERO,NO_ENGINE_SUBSTITUTION,NO_ZERO_DATE,NO_ZERO_IN_DATE,STRICT_ALL_TABLES"),
			SQLRequirePrimaryKey:        PtrTo(true),
			InnodbFtMinTokenSize:        PtrTo(3),
			InnodbFtServerStopwordTable: PtrTo(""),
			InnodbPrintAllDeadlocks:     PtrTo(false),
			InnodbRollbackOnTimeout:     PtrTo(false),
			SlowQueryLog:                PtrTo(false),
			LongQueryTime:               PtrTo(float32(10)),
			BackupHour:                  PtrTo(21),
			BackupMinute:                PtrTo(59),
		}
	)

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, mySQLConfigJSON)
	})

	got, _, err := dbSvc.GetMySQLConfig(ctx, dbID)
	require.NoError(t, err)
	require.Equal(t, &mySQLConfig, got)
}

func TestDatabases_UpdateConfigMySQL(t *testing.T) {
	setup()
	defer teardown()

	var (
		dbID        = "deadbeef-dead-4aa5-beef-deadbeef347d"
		path        = fmt.Sprintf("/v2/databases/%s/config", dbID)
		mySQLConfig = &MySQLConfig{
			SQLRequirePrimaryKey:        PtrTo(true),
			InnodbFtMinTokenSize:        PtrTo(3),
			InnodbFtServerStopwordTable: PtrTo(""),
		}
	)

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPatch)

		var b databaseMySQLConfigRoot
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&b)
		require.NoError(t, err)

		assert.Equal(t, b.Config, mySQLConfig)
		assert.Equal(t, "", *b.Config.InnodbFtServerStopwordTable, "pointers to zero value should be sent")
		assert.Nil(t, b.Config.InnodbPrintAllDeadlocks, "excluded value should not be sent")

		w.WriteHeader(http.StatusNoContent)
	})

	_, err := client.Databases.UpdateMySQLConfig(ctx, dbID, mySQLConfig)
	require.NoError(t, err)
}

func TestDatabases_GetConfigMongoDB(t *testing.T) {
	setup()
	defer teardown()

	var (
		dbSvc = client.Databases
		dbID  = "da4e0206-d019-41d7-b51f-deadbeefbb8f"
		path  = fmt.Sprintf("/v2/databases/%s/config", dbID)

		mongoDBConfigJSON = `{
  "config": {
    "default_read_concern": "LOCAL",
    "default_write_concern": "majority",
    "transaction_lifetime_limit_seconds": 60,
    "slow_op_threshold_ms": 100,
    "verbosity": 0
  }
}`

		mongoDBConfig = MongoDBConfig{
			DefaultReadConcern:              PtrTo("LOCAL"),
			DefaultWriteConcern:             PtrTo("majority"),
			TransactionLifetimeLimitSeconds: PtrTo(60),
			SlowOpThresholdMs:               PtrTo(100),
			Verbosity:                       PtrTo(0),
		}
	)

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, mongoDBConfigJSON)
	})

	got, _, err := dbSvc.GetMongoDBConfig(ctx, dbID)
	require.NoError(t, err)
	require.Equal(t, &mongoDBConfig, got)
}

func TestDatabases_UpdateConfigMongoDB(t *testing.T) {
	setup()
	defer teardown()

	var (
		dbID          = "deadbeef-dead-4aa5-beef-deadbeef347d"
		path          = fmt.Sprintf("/v2/databases/%s/config", dbID)
		mongoDBConfig = &MongoDBConfig{
			DefaultReadConcern:  PtrTo("AVAILABLE"),
			DefaultWriteConcern: PtrTo(""),
			SlowOpThresholdMs:   PtrTo(0),
			Verbosity:           PtrTo(5),
		}
	)

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPatch)

		var b databaseMongoDBConfigRoot
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&b)
		require.NoError(t, err)

		assert.Equal(t, b.Config, mongoDBConfig)
		assert.Equal(t, "", *b.Config.DefaultWriteConcern, "pointers to zero value should be sent")
		assert.Nil(t, b.Config.TransactionLifetimeLimitSeconds, "excluded value should not be sent")

		w.WriteHeader(http.StatusNoContent)
	})

	_, err := client.Databases.UpdateMongoDBConfig(ctx, dbID, mongoDBConfig)
	require.NoError(t, err)
}

func TestDatabases_GetConfigKafka(t *testing.T) {
	setup()
	defer teardown()

	var (
		dbSvc = client.Databases
		dbID  = "da4e0206-d019-41d7-b51f-deadbeefbb8f"
		path  = fmt.Sprintf("/v2/databases/%s/config", dbID)

		kafkaConfigJSON = `{
  "config": {
    "group_initial_rebalance_delay_ms": 3000,
    "group_min_session_timeout_ms": 6000,
    "group_max_session_timeout_ms": 1800000,
    "message_max_bytes": 1048588,
    "log_cleaner_delete_retention_ms": 86400000,
    "log_cleaner_min_compaction_lag_ms": 0,
    "log_flush_interval_ms": 60000,
    "log_index_interval_bytes": 4096,
    "log_message_downconversion_enable": true,
    "log_message_timestamp_difference_max_ms": 120000,
    "log_preallocate": false,
    "log_retention_bytes": -1,
    "log_retention_hours": 168,
    "log_retention_ms": 604800000,
    "log_roll_jitter_ms": 0,
    "log_segment_delete_delay_ms": 60000,
    "auto_create_topics_enable": true
  }
}`

		kafkaConfig = KafkaConfig{
			GroupInitialRebalanceDelayMs:       PtrTo(3000),
			GroupMinSessionTimeoutMs:           PtrTo(6000),
			GroupMaxSessionTimeoutMs:           PtrTo(1800000),
			MessageMaxBytes:                    PtrTo(1048588),
			LogCleanerDeleteRetentionMs:        PtrTo(int64(86400000)),
			LogCleanerMinCompactionLagMs:       PtrTo(uint64(0)),
			LogFlushIntervalMs:                 PtrTo(uint64(60000)),
			LogIndexIntervalBytes:              PtrTo(4096),
			LogMessageDownconversionEnable:     PtrTo(true),
			LogMessageTimestampDifferenceMaxMs: PtrTo(uint64(120000)),
			LogPreallocate:                     PtrTo(false),
			LogRetentionBytes:                  big.NewInt(int64(-1)),
			LogRetentionHours:                  PtrTo(168),
			LogRetentionMs:                     big.NewInt(int64(604800000)),
			LogRollJitterMs:                    PtrTo(uint64(0)),
			LogSegmentDeleteDelayMs:            PtrTo(60000),
			AutoCreateTopicsEnable:             PtrTo(true),
		}
	)

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, kafkaConfigJSON)
	})

	got, _, err := dbSvc.GetKafkaConfig(ctx, dbID)
	require.NoError(t, err)
	require.Equal(t, &kafkaConfig, got)
}

func TestDatabases_UpdateConfigKafka(t *testing.T) {
	setup()
	defer teardown()

	var (
		dbID        = "deadbeef-dead-4aa5-beef-deadbeef347d"
		path        = fmt.Sprintf("/v2/databases/%s/config", dbID)
		kafkaConfig = &KafkaConfig{
			GroupInitialRebalanceDelayMs: PtrTo(3000),
			GroupMinSessionTimeoutMs:     PtrTo(6000),
			GroupMaxSessionTimeoutMs:     PtrTo(1800000),
			MessageMaxBytes:              PtrTo(1048588),
			LogCleanerDeleteRetentionMs:  PtrTo(int64(86400000)),
			LogCleanerMinCompactionLagMs: PtrTo(uint64(0)),
		}
	)

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPatch)

		var b databaseKafkaConfigRoot
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&b)
		require.NoError(t, err)

		assert.Equal(t, b.Config, kafkaConfig)
		assert.Equal(t, uint64(0), *b.Config.LogCleanerMinCompactionLagMs, "pointers to zero value should be sent")
		assert.Nil(t, b.Config.LogFlushIntervalMs, "excluded value should not be sent")

		w.WriteHeader(http.StatusNoContent)
	})

	_, err := client.Databases.UpdateKafkaConfig(ctx, dbID, kafkaConfig)
	require.NoError(t, err)
}

func TestDatabases_GetConfigOpensearch(t *testing.T) {
	setup()
	defer teardown()

	var (
		dbSvc = client.Databases
		dbID  = "da4e0206-d019-41d7-b51f-deadbeefbb8f"
		path  = fmt.Sprintf("/v2/databases/%s/config", dbID)

		opensearchConfigJSON = `{
  "config": {
    "ism_enabled": true,
    "ism_history_enabled": true,
    "ism_history_max_age_hours": 24,
    "ism_history_max_docs": 2500000,
    "ism_history_rollover_check_period_hours": 8,
    "ism_history_rollover_retention_period_days": 30,
    "http_max_content_length_bytes": 100000000,
    "http_max_header_size_bytes": 8192,
    "http_max_initial_line_length_bytes": 4096,
    "indices_query_bool_max_clause_count": 1024,
    "search_max_buckets": 10000,
    "indices_fielddata_cache_size_percentage": 0,
    "indices_memory_index_buffer_size_percentage": 10,
    "indices_memory_min_index_buffer_size_mb": 48,
    "indices_memory_max_index_buffer_size_mb": 0,
    "indices_queries_cache_size_percentage": 10,
    "indices_recovery_max_mb_per_sec": 40,
    "indices_recovery_max_concurrent_file_chunks": 2,
    "action_auto_create_index_enabled": true,
    "action_destructive_requires_name": false,
    "plugins_alerting_filter_by_backend_roles_enabled": false,
    "enable_security_audit": false,
    "thread_pool_search_size": 0,
    "thread_pool_search_throttled_size": 0,
    "thread_pool_search_throttled_queue_size": 0,
    "thread_pool_search_queue_size": 0,
    "thread_pool_get_size": 0,
    "thread_pool_get_queue_size": 0,
    "thread_pool_analyze_size": 0,
    "thread_pool_analyze_queue_size": 0,
    "thread_pool_write_size": 0,
    "thread_pool_write_queue_size": 0,
    "thread_pool_force_merge_size": 0,
    "override_main_response_version": false,
    "script_max_compilations_rate": "use-context",
    "cluster_max_shards_per_node": 0,
    "cluster_routing_allocation_node_concurrent_recoveries": 2,
    "plugins_alerting_filter_by_backend_roles_enabled": true
  }
}`

		opensearchConfig = OpensearchConfig{
			HttpMaxContentLengthBytes:                        PtrTo(100000000),
			HttpMaxHeaderSizeBytes:                           PtrTo(8192),
			HttpMaxInitialLineLengthBytes:                    PtrTo(4096),
			IndicesQueryBoolMaxClauseCount:                   PtrTo(1024),
			IndicesFielddataCacheSizePercentage:              PtrTo(0),
			IndicesMemoryIndexBufferSizePercentage:           PtrTo(10),
			IndicesMemoryMinIndexBufferSizeMb:                PtrTo(48),
			IndicesMemoryMaxIndexBufferSizeMb:                PtrTo(0),
			IndicesQueriesCacheSizePercentage:                PtrTo(10),
			IndicesRecoveryMaxMbPerSec:                       PtrTo(40),
			IndicesRecoveryMaxConcurrentFileChunks:           PtrTo(2),
			ThreadPoolSearchSize:                             PtrTo(0),
			ThreadPoolSearchThrottledSize:                    PtrTo(0),
			ThreadPoolGetSize:                                PtrTo(0),
			ThreadPoolAnalyzeSize:                            PtrTo(0),
			ThreadPoolWriteSize:                              PtrTo(0),
			ThreadPoolForceMergeSize:                         PtrTo(0),
			ThreadPoolSearchQueueSize:                        PtrTo(0),
			ThreadPoolSearchThrottledQueueSize:               PtrTo(0),
			ThreadPoolGetQueueSize:                           PtrTo(0),
			ThreadPoolAnalyzeQueueSize:                       PtrTo(0),
			ThreadPoolWriteQueueSize:                         PtrTo(0),
			IsmEnabled:                                       PtrTo(true),
			IsmHistoryEnabled:                                PtrTo(true),
			IsmHistoryMaxAgeHours:                            PtrTo(24),
			IsmHistoryMaxDocs:                                PtrTo(int64(2500000)),
			IsmHistoryRolloverCheckPeriodHours:               PtrTo(8),
			IsmHistoryRolloverRetentionPeriodDays:            PtrTo(30),
			SearchMaxBuckets:                                 PtrTo(10000),
			ActionAutoCreateIndexEnabled:                     PtrTo(true),
			EnableSecurityAudit:                              PtrTo(false),
			ActionDestructiveRequiresName:                    PtrTo(false),
			ClusterMaxShardsPerNode:                          PtrTo(0),
			OverrideMainResponseVersion:                      PtrTo(false),
			ScriptMaxCompilationsRate:                        PtrTo("use-context"),
			ClusterRoutingAllocationNodeConcurrentRecoveries: PtrTo(2),
			ReindexRemoteWhitelist:                           nil,
			PluginsAlertingFilterByBackendRolesEnabled:       PtrTo(true),
		}
	)

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, opensearchConfigJSON)
	})

	got, _, err := dbSvc.GetOpensearchConfig(ctx, dbID)
	require.NoError(t, err)
	require.Equal(t, &opensearchConfig, got)
}

func TestDatabases_UpdateConfigOpensearch(t *testing.T) {
	setup()
	defer teardown()

	var (
		dbID             = "deadbeef-dead-4aa5-beef-deadbeef347d"
		path             = fmt.Sprintf("/v2/databases/%s/config", dbID)
		opensearchConfig = &OpensearchConfig{
			HttpMaxContentLengthBytes: PtrTo(1),
			HttpMaxHeaderSizeBytes:    PtrTo(0),
		}
	)

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPatch)

		var b databaseOpensearchConfigRoot
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&b)
		require.NoError(t, err)

		assert.Equal(t, b.Config, opensearchConfig)
		assert.Equal(t, 0, *b.Config.HttpMaxHeaderSizeBytes, "pointers to zero value should be sent")
		assert.Nil(t, b.Config.HttpMaxInitialLineLengthBytes, "excluded value should not be sent")

		w.WriteHeader(http.StatusNoContent)
	})

	_, err := client.Databases.UpdateOpensearchConfig(ctx, dbID, opensearchConfig)
	require.NoError(t, err)
}

func TestDatabases_UpgradeMajorVersion(t *testing.T) {
	setup()
	defer teardown()

	var (
		dbID              = "deadbeef-dead-4aa5-beef-deadbeef347d"
		path              = fmt.Sprintf("/v2/databases/%s/upgrade", dbID)
		upgradeVersionReq = &UpgradeVersionRequest{
			Version: "14",
		}
	)
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		var b UpgradeVersionRequest
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&b)
		require.NoError(t, err)
		assert.Equal(t, b.Version, upgradeVersionReq.Version)
		w.WriteHeader(http.StatusNoContent)
	})
	_, err := client.Databases.UpgradeMajorVersion(ctx, dbID, upgradeVersionReq)
	require.NoError(t, err)
}

func TestDatabases_CreateTopic(t *testing.T) {
	setup()
	defer teardown()

	var (
		dbID              = "deadbeef-dead-4aa5-beef-deadbeef347d"
		numPartitions     = uint32(3)
		replicationFactor = uint32(2)
		retentionMS       = int64(1000 * 60)
	)

	want := &DatabaseTopic{
		Name:              "events",
		ReplicationFactor: &replicationFactor,
		Config: &TopicConfig{
			RetentionMS: &retentionMS,
		},
	}

	body := `{
	  "topic": {
	    "name": "events",
	    "partition_count": 3,
	    "replication_factor": 2,
	    "config": {
	    	"retention_ms": 60000
	    }
	  }
	}`

	path := fmt.Sprintf("/v2/databases/%s/topics", dbID)

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, body)
	})

	topic, _, err := client.Databases.CreateTopic(ctx, dbID, &DatabaseCreateTopicRequest{
		Name:              "events",
		PartitionCount:    &numPartitions,
		ReplicationFactor: &replicationFactor,
		Config: &TopicConfig{
			RetentionMS: &retentionMS,
		},
	})

	require.NoError(t, err)
	require.Equal(t, want, topic)
}

func TestDatabases_UpdateTopic(t *testing.T) {
	setup()
	defer teardown()

	var (
		dbID              = "deadbeef-dead-4aa5-beef-deadbeef347d"
		topicName         = "events"
		numPartitions     = uint32(3)
		replicationFactor = uint32(2)
		retentionMS       = int64(1000 * 60)
	)

	body := `{
	  "topic": {
	    "name": "events",
	    "partition_count": 3,
	    "replication_factor": 2,
	    "config": {
	    	"retention_ms": 60000
	    }
	  }
	}`

	path := fmt.Sprintf("/v2/databases/%s/topics/%s", dbID, topicName)

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		fmt.Fprint(w, body)
	})

	_, err := client.Databases.UpdateTopic(ctx, dbID, topicName, &DatabaseUpdateTopicRequest{
		PartitionCount:    &numPartitions,
		ReplicationFactor: &replicationFactor,
		Config: &TopicConfig{
			RetentionMS: &retentionMS,
		},
	})

	require.NoError(t, err)
}

func TestDatabases_DeleteTopic(t *testing.T) {
	setup()
	defer teardown()

	var (
		dbID      = "deadbeef-dead-4aa5-beef-deadbeef347d"
		topicName = "events"
	)

	path := fmt.Sprintf("/v2/databases/%s/topics/%s", dbID, topicName)

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
	})

	_, err := client.Databases.DeleteTopic(ctx, dbID, topicName)
	require.NoError(t, err)
}

func TestDatabases_GetTopic(t *testing.T) {
	setup()
	defer teardown()

	var (
		dbID              = "deadbeef-dead-4aa5-beef-deadbeef347d"
		topicName         = "events"
		replicationFactor = uint32(2)
		retentionMS       = int64(1000 * 60)
	)

	want := &DatabaseTopic{
		Name: "events",
		Partitions: []*TopicPartition{
			{
				Size:           0,
				Id:             0,
				InSyncReplicas: 2,
				EarliestOffset: 0,
				ConsumerGroups: nil,
			},
			{
				Size:           0,
				Id:             1,
				InSyncReplicas: 2,
				EarliestOffset: 0,
				ConsumerGroups: nil,
			},
			{
				Size:           0,
				Id:             2,
				InSyncReplicas: 2,
				EarliestOffset: 0,
				ConsumerGroups: nil,
			},
		},
		ReplicationFactor: &replicationFactor,
		Config: &TopicConfig{
			RetentionMS: &retentionMS,
		},
	}

	body := `{
		"topic":{
		   "name":"events",
		   "replication_factor":2,
		   "config":{
			  "retention_ms":60000
		   },
		   "partitions":[
			  {
				 "size":0,
				 "id":0,
				 "in_sync_replicas":2,
				 "earliest_offset":0,
				 "consumer_groups":null
			  },
			  {
				 "size":0,
				 "id":1,
				 "in_sync_replicas":2,
				 "earliest_offset":0,
				 "consumer_groups":null
			  },
			  {
				 "size":0,
				 "id":2,
				 "in_sync_replicas":2,
				 "earliest_offset":0,
				 "consumer_groups":null
			  }
		   ]
		}
	 }`

	path := fmt.Sprintf("/v2/databases/%s/topics/%s", dbID, topicName)

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, body)
	})

	got, _, err := client.Databases.GetTopic(ctx, dbID, topicName)
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestDatabases_ListTopics(t *testing.T) {
	setup()
	defer teardown()

	var (
		dbID              = "deadbeef-dead-4aa5-beef-deadbeef347d"
		replicationFactor = uint32(2)
		retentionMS       = int64(1000 * 60)
	)

	want := []DatabaseTopic{
		{
			Name: "events",
			Partitions: []*TopicPartition{
				{
					Size:           0,
					Id:             0,
					InSyncReplicas: 2,
					EarliestOffset: 0,
					ConsumerGroups: nil,
				},
				{
					Size:           0,
					Id:             1,
					InSyncReplicas: 2,
					EarliestOffset: 0,
					ConsumerGroups: nil,
				},
				{
					Size:           0,
					Id:             2,
					InSyncReplicas: 2,
					EarliestOffset: 0,
					ConsumerGroups: nil,
				},
			},
			ReplicationFactor: &replicationFactor,
			Config: &TopicConfig{
				RetentionMS: &retentionMS,
			},
		},
		{
			Name: "events_ii",
			Partitions: []*TopicPartition{
				{
					Size:           0,
					Id:             0,
					InSyncReplicas: 2,
					EarliestOffset: 0,
					ConsumerGroups: nil,
				},
				{
					Size:           0,
					Id:             1,
					InSyncReplicas: 2,
					EarliestOffset: 0,
					ConsumerGroups: nil,
				},
				{
					Size:           0,
					Id:             2,
					InSyncReplicas: 2,
					EarliestOffset: 0,
					ConsumerGroups: nil,
				},
			},
			ReplicationFactor: &replicationFactor,
			Config: &TopicConfig{
				RetentionMS: &retentionMS,
			},
		},
	}

	body := `{
	  "topics": [
	  	{
		    "name": "events",
			"partitions":[
				{
				   "size":0,
				   "id":0,
				   "in_sync_replicas":2,
				   "earliest_offset":0,
				   "consumer_groups":null
				},
				{
				   "size":0,
				   "id":1,
				   "in_sync_replicas":2,
				   "earliest_offset":0,
				   "consumer_groups":null
				},
				{
				   "size":0,
				   "id":2,
				   "in_sync_replicas":2,
				   "earliest_offset":0,
				   "consumer_groups":null
				}
			],
		    "replication_factor": 2,
		    "config": {
		    	"retention_ms": 60000
		    }
		  },
		  {
		    "name": "events_ii",
			"partitions":[
				{
				   "size":0,
				   "id":0,
				   "in_sync_replicas":2,
				   "earliest_offset":0,
				   "consumer_groups":null
				},
				{
				   "size":0,
				   "id":1,
				   "in_sync_replicas":2,
				   "earliest_offset":0,
				   "consumer_groups":null
				},
				{
				   "size":0,
				   "id":2,
				   "in_sync_replicas":2,
				   "earliest_offset":0,
				   "consumer_groups":null
				}
			],
		    "replication_factor": 2,
		    "config": {
		    	"retention_ms": 60000
		    }
		  }
		]		 
	}`

	path := fmt.Sprintf("/v2/databases/%s/topics", dbID)

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, body)
	})

	got, _, err := client.Databases.ListTopics(ctx, dbID, &ListOptions{})
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestDatabases_GetMetricsCredentials(t *testing.T) {
	setup()
	defer teardown()

	want := &DatabaseMetricsCredentials{
		BasicAuthUsername: "username_for_http_basic_auth",
		BasicAuthPassword: "password_for_http_basic_auth",
	}

	body := `{
		"credentials": {
			"basic_auth_username": "username_for_http_basic_auth",
			"basic_auth_password": "password_for_http_basic_auth"
		}
	}`

	mux.HandleFunc("/v2/databases/metrics/credentials", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, body)
	})

	got, _, err := client.Databases.GetMetricsCredentials(ctx)
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestDatabases_UpdateMetricsCredentials(t *testing.T) {
	setup()
	defer teardown()

	updateRequest := &DatabaseUpdateMetricsCredentialsRequest{
		Credentials: &DatabaseMetricsCredentials{
			BasicAuthUsername: "username_for_http_basic_auth",
			BasicAuthPassword: "password_for_http_basic_auth",
		},
	}

	mux.HandleFunc("/v2/databases/metrics/credentials", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
	})

	_, err := client.Databases.UpdateMetricsCredentials(ctx, updateRequest)
	require.NoError(t, err)
}

func TestDatabases_ListDatabaseEvents(t *testing.T) {
	setup()
	defer teardown()

	dbID := "deadbeef-dead-4aa5-beef-deadbeef347d"

	path := fmt.Sprintf("/v2/databases/%s/events", dbID)

	want := []DatabaseEvent{
		{
			ID:          "pe8u2huh",
			ServiceName: "customer-events",
			EventType:   "cluster_create",
			CreateTime:  "2020-10-29T15:57:38Z",
		},
	}

	body := `{
		"events": [
		  {
			"id": "pe8u2huh",
			"cluster_name": "customer-events",
			"event_type": "cluster_create",
			"create_time": "2020-10-29T15:57:38Z"
		  }
		]
	  } `

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, body)
	})

	got, _, err := client.Databases.ListDatabaseEvents(ctx, dbID, &ListOptions{})
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestDatabases_ListIndexes(t *testing.T) {
	setup()
	defer teardown()

	dbID := "deadbeef-dead-4aa5-beef-deadbeef347d"

	path := fmt.Sprintf("/v2/databases/%s/indexes", dbID)

	want := []DatabaseIndex{
		{
			IndexName:        "sample_index",
			NumberofShards:   uint64(1),
			NumberofReplicas: uint64(0),
			CreateTime:       "2020-10-29T15:57:38Z",
			Health:           "green",
			Size:             int64(5314),
			Status:           "open",
			Docs:             int64(64811),
		},
		{
			IndexName:        "sample_index_2",
			NumberofShards:   uint64(1),
			NumberofReplicas: uint64(0),
			CreateTime:       "2020-10-30T15:57:38Z",
			Health:           "red",
			Size:             int64(6105247),
			Status:           "close",
			Docs:             int64(64801),
		},
	}

	body := `{
		"indexes": [
			{
            "create_time": "2020-10-29T15:57:38Z",
            "docs": 64811,
            "health": "green",
            "index_name": "sample_index",
            "number_of_replica": 0,
            "number_of_shards": 1,
            "size": 5314,
            "status": "open"
        },
        {
            "create_time": "2020-10-30T15:57:38Z",
            "docs": 64801,
            "health": "red",
            "index_name": "sample_index_2",
            "number_of_replica": 0,
            "number_of_shards": 1,
            "size": 6105247,
            "status": "close"
        }]
	  } `

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, body)
	})

	got, _, err := client.Databases.ListIndexes(ctx, dbID, &ListOptions{})
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestDatabases_DeleteIndexes(t *testing.T) {
	setup()
	defer teardown()

	dbID := "deadbeef-dead-4aa5-beef-deadbeef347d"
	indexName := "sample_index"

	path := fmt.Sprintf("/v2/databases/%s/indexes/%s", dbID, indexName)

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
	})

	_, err := client.Databases.DeleteIndex(ctx, dbID, indexName)
	require.NoError(t, err)
}

func TestDatabases_CreateLogsink(t *testing.T) {
	setup()
	defer teardown()

	var (
		dbID = "deadbeef-dead-4aa5-beef-deadbeef347d"
	)

	want := &DatabaseLogsink{
		ID:   "deadbeef-dead-4aa5-beef-deadbeef347d",
		Name: "logs-sink",
		Type: "opensearch",
		Config: &DatabaseLogsinkConfig{
			URL:         "https://user:passwd@192.168.0.1:25060",
			IndexPrefix: "opensearch-logs",
		},
	}

	body := `{
        "sink_id":"deadbeef-dead-4aa5-beef-deadbeef347d",
        "sink_name": "logs-sink",
        "sink_type": "opensearch",
        "config": {
          "url": "https://user:passwd@192.168.0.1:25060",
          "index_prefix": "opensearch-logs"
        }
      }`

	path := fmt.Sprintf("/v2/databases/%s/logsink", dbID)

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, body)
	})

	log, _, err := client.Databases.CreateLogsink(ctx, dbID, &DatabaseCreateLogsinkRequest{
		Name: "logs-sink",
		Type: "opensearch",
		Config: &DatabaseLogsinkConfig{
			URL:         "https://user:passwd@192.168.0.1:25060",
			IndexPrefix: "opensearch-logs",
		},
	})

	require.NoError(t, err)

	require.Equal(t, want, log)
}

func TestDatabases_GetLogsink(t *testing.T) {
	setup()
	defer teardown()

	var (
		dbID      = "deadbeef-dead-4aa5-beef-deadbeef347d"
		logsinkID = "50484ec3-19d6-4cd3-b56f-3b0381c289a6"
	)

	want := &DatabaseLogsink{
		ID:   "deadbeef-dead-4aa5-beef-deadbeef347d",
		Name: "logs-sink",
		Type: "opensearch",
		Config: &DatabaseLogsinkConfig{
			URL:         "https://user:passwd@192.168.0.1:25060",
			IndexPrefix: "opensearch-logs",
		},
	}

	body := `{
        "sink_id":"deadbeef-dead-4aa5-beef-deadbeef347d",
        "sink_name": "logs-sink",
        "sink_type": "opensearch",
        "config": {
          "url": "https://user:passwd@192.168.0.1:25060",
          "index_prefix": "opensearch-logs"
        }
      }`

	path := fmt.Sprintf("/v2/databases/%s/logsink/%s", dbID, logsinkID)

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, body)
	})

	got, _, err := client.Databases.GetLogsink(ctx, dbID, logsinkID)
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestDatabases_UpdateLogsink(t *testing.T) {
	setup()
	defer teardown()

	var (
		dbID      = "deadbeef-dead-4aa5-beef-deadbeef347d"
		logsinkID = "50484ec3-19d6-4cd3-b56f-3b0381c289a6"
	)

	body := `{
        "sink_id":"deadbeef-dead-4aa5-beef-deadbeef347d",
        "sink_name": "logs-sink",
        "sink_type": "opensearch",
        "config": {
          "url": "https://user:passwd@192.168.0.1:25060",
          "index_prefix": "opensearch-logs"
        }
      }`

	path := fmt.Sprintf("/v2/databases/%s/logsink/%s", dbID, logsinkID)

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		fmt.Fprint(w, body)
	})

	_, err := client.Databases.UpdateLogsink(ctx, dbID, logsinkID, &DatabaseUpdateLogsinkRequest{
		Config: &DatabaseLogsinkConfig{
			Server: "192.168.0.1",
			Port:   514,
			TLS:    false,
			Format: "rfc3164",
		},
	})

	require.NoError(t, err)
}

func TestDatabases_ListLogsinks(t *testing.T) {
	setup()
	defer teardown()

	var (
		dbID = "deadbeef-dead-4aa5-beef-deadbeef347d"
	)

	want := []DatabaseLogsink{
		{
			ID:   "deadbeef-dead-4aa5-beef-deadbeef347d",
			Name: "logs-sink",
			Type: "opensearch",
			Config: &DatabaseLogsinkConfig{
				URL:         "https://user:passwd@192.168.0.1:25060",
				IndexPrefix: "opensearch-logs",
			},
		},
		{
			ID:   "d6e95157-5f58-48d0-9023-8cfb409d102a",
			Name: "logs-sink-2",
			Type: "opensearch",
			Config: &DatabaseLogsinkConfig{
				URL:         "https://user:passwd@192.168.0.1:25060",
				IndexPrefix: "opensearch-logs",
			},
		}}

	body := `{
		"sinks": [
		  {
			"sink_id": "deadbeef-dead-4aa5-beef-deadbeef347d",
			"sink_name": "logs-sink",
			"sink_type": "opensearch",
			"config": {
			  "url": "https://user:passwd@192.168.0.1:25060",
			  "index_prefix": "opensearch-logs"
			}
		  },
		  {
			"sink_id": "d6e95157-5f58-48d0-9023-8cfb409d102a",
			"sink_name": "logs-sink-2",
			"sink_type": "opensearch",
			"config": {
				"url": "https://user:passwd@192.168.0.1:25060",
				"index_prefix": "opensearch-logs"
			}
		  }
		]
	  }`

	path := fmt.Sprintf("/v2/databases/%s/logsink", dbID)

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, body)
	})

	got, _, err := client.Databases.ListLogsinks(ctx, dbID, &ListOptions{})
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestDatabases_DeleteLogsink(t *testing.T) {
	setup()
	defer teardown()

	var (
		dbID      = "deadbeef-dead-4aa5-beef-deadbeef347d"
		logsinkID = "50484ec3-19d6-4cd3-b56f-3b0381c289a6"
	)

	path := fmt.Sprintf("/v2/databases/%s/logsink/%s", dbID, logsinkID)

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
	})

	_, err := client.Databases.DeleteLogsink(ctx, dbID, logsinkID)
	require.NoError(t, err)
}

func TestDatabases_StartOnlineMigration(t *testing.T) {
	setup()
	defer teardown()

	var (
		dbID = "deadbeef-dead-4aa5-beef-deadbeef347d"
	)

	body := `{
		"source": {
			"host": "source-do-user-6607903-0.b.db.ondigitalocean.com",
			"dbname": "defaultdb",
			"port": 25060,
			"username": "doadmin",
			"password": "paakjnfe10rsrsmf"
		},
		"disable_ssl": false,
		"ignore_dbs": [
			"db0",
			"db1"
		]
		}`

	path := fmt.Sprintf("/v2/databases/%s/online-migration", dbID)

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		fmt.Fprint(w, body)
	})

	_, resp, err := client.Databases.StartOnlineMigration(ctx, dbID, &DatabaseStartOnlineMigrationRequest{
		DisableSSL: false,

		Source: &DatabaseOnlineMigrationConfig{
			Host:         "https://user:passwd@192.168.0.1:25060",
			DatabaseName: "defaultdb",
			Port:         25060,
			Username:     "doadmin",
			Password:     "paakjnfe10rsrsmf",
		},
	})

	require.NoError(t, err)

	require.Equal(t, 200, resp.StatusCode)
}

func TestDatabases_GetOnlineMigrationStatus(t *testing.T) {
	setup()
	defer teardown()

	var (
		dbID = "deadbeef-dead-4aa5-beef-deadbeef347d"
	)

	body := `{
		"source": {
			"host": "source-do-user-6607903-0.b.db.ondigitalocean.com",
			"dbname": "defaultdb",
			"port": 25060,
			"username": "doadmin",
			"password": "paakjnfe10rsrsmf"
		},
		"disable_ssl": false,
		"ignore_dbs": [
			"db0",
			"db1"
		]
		}`
	path := fmt.Sprintf("/v2/databases/%s/online-migration", dbID)

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, body)
	})

	_, resp, err := client.Databases.GetOnlineMigrationStatus(ctx, dbID)
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)
}

func TestDatabases_StopOnlineMigration(t *testing.T) {
	setup()
	defer teardown()

	var (
		dbID        = "deadbeef-dead-4aa5-beef-deadbeef347d"
		migrationID = "50484ec3-19d6-4cd3-b56f-3b0381c289a6"
	)

	path := fmt.Sprintf("/v2/databases/%s/online-migration/%s", dbID, migrationID)

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
	})

	_, err := client.Databases.StopOnlineMigration(ctx, dbID, migrationID)
	require.NoError(t, err)
}

func TestDatabases_CreateDatabaseUserWithMongoUserSettings(t *testing.T) {
	setup()
	defer teardown()

	dbID := "deadbeef-dead-4aa5-beef-deadbeef347d"
	path := fmt.Sprintf("/v2/databases/%s/users", dbID)

	writeMongoSettings := &MongoUserSettings{
		Databases: []string{"bar"},
		Role:      "readWrite",
	}

	acljson, err := json.Marshal(writeMongoSettings)
	if err != nil {
		t.Fatal(err)
	}

	responseJSON := []byte(fmt.Sprintf(`{
		"user": {
			"name": "foo",
			"settings": {
				"mongo_user_settings": %s
			}
		}
	}`, string(acljson)))

	expectedUser := &DatabaseUser{
		Name: "foo",
		Settings: &DatabaseUserSettings{
			MongoUserSettings: writeMongoSettings,
		},
	}

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		w.WriteHeader(http.StatusOK)
		w.Write(responseJSON)
	})

	user, _, err := client.Databases.CreateUser(ctx, dbID, &DatabaseCreateUserRequest{
		Name:     expectedUser.Name,
		Settings: &DatabaseUserSettings{MongoUserSettings: expectedUser.Settings.MongoUserSettings},
	})
	require.NoError(t, err)
	require.Equal(t, expectedUser, user)
}

func TestDatabases_CreateKafkaSchemaRegistry(t *testing.T) {
	setup()
	defer teardown()

	dbID := "deadbeef-dead-4aa5-beef-deadbeef347d"
	path := fmt.Sprintf("/v2/databases/%s/schema-registry", dbID)

	responseJSON, err := json.Marshal(&DatabaseKafkaSchemaRegistrySubject{
		SubjectName: "test-subject",
		SchemaType:  "AVRO",
		Schema:      `{"type":"record","name":"test","fields":[{"name":"field1","type":"string"}]}`,
		SchemaID:    1,
	})
	if err != nil {
		t.Fatal(err)
	}

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		w.WriteHeader(http.StatusOK)
		w.Write(responseJSON)
	})

	kafkaSchemaRegistry, _, err := client.Databases.CreateKafkaSchemaRegistry(ctx, dbID, &DatabaseKafkaSchemaRegistryRequest{
		SubjectName: "test-subject",
		SchemaType:  "AVRO",
		Schema:      `{"type":"record","name":"test","fields":[{"name":"field1","type":"string"}]}`,
	})
	require.NoError(t, err)
	require.Equal(t, &DatabaseKafkaSchemaRegistrySubject{
		SubjectName: "test-subject",
		SchemaType:  "AVRO",
		Schema:      `{"type":"record","name":"test","fields":[{"name":"field1","type":"string"}]}`,
		SchemaID:    1,
	}, kafkaSchemaRegistry)
}

func TestDatabases_ListKafkaSchemaRegistry(t *testing.T) {
	setup()
	defer teardown()

	dbID := "deadbeef-dead-4aa5-beef-deadbeef347d"
	path := fmt.Sprintf("/v2/databases/%s/schema-registry", dbID)

	responseJSON, err := json.Marshal(&ListDatabaseKafkaSchemaRegistrySubjectsRoot{
		Subjects: []DatabaseKafkaSchemaRegistrySubject{
			{
				SubjectName: "test-subject",
				SchemaType:  "AVRO",
				Schema:      `{"type":"record","name":"test","fields":[{"name":"field1","type":"string"}]}`,
				SchemaID:    1,
			},
			{
				SubjectName: "test-subject-2",
				SchemaType:  "AVRO",
				Schema:      `{"type":"record","name":"test","fields":[{"name":"field1","type":"string"}]}`,
				SchemaID:    2,
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.WriteHeader(http.StatusOK)
		w.Write(responseJSON)
	})

	kafkaSchemaRegistrySubjects, _, err := client.Databases.ListKafkaSchemaRegistry(ctx, dbID, &ListOptions{})
	require.NoError(t, err)
	require.Equal(t, 2, len(kafkaSchemaRegistrySubjects))
	require.Equal(t, "test-subject", kafkaSchemaRegistrySubjects[0].SubjectName)
	require.Equal(t, "AVRO", kafkaSchemaRegistrySubjects[0].SchemaType)
	require.Equal(t, `{"type":"record","name":"test","fields":[{"name":"field1","type":"string"}]}`, kafkaSchemaRegistrySubjects[0].Schema)
	require.Equal(t, 1, kafkaSchemaRegistrySubjects[0].SchemaID)
}

func TestDatabases_GetKafkaSchemaRegistry(t *testing.T) {
	setup()
	defer teardown()

	dbID := "deadbeef-dead-4aa5-beef-deadbeef347d"
	subjectName := "test-subject"
	path := fmt.Sprintf("/v2/databases/%s/schema-registry/%s", dbID, subjectName)

	responseJSON, err := json.Marshal(&DatabaseKafkaSchemaRegistrySubject{
		SubjectName: subjectName,
		SchemaType:  "AVRO",
		Schema:      `{"type":"record","name":"test","fields":[{"name":"field1","type":"string"}]}`,
		SchemaID:    1,
	})
	if err != nil {
		t.Fatal(err)
	}

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.WriteHeader(http.StatusOK)
		w.Write(responseJSON)
	})

	kafkaSchemaRegistrySubject, _, err := client.Databases.GetKafkaSchemaRegistry(ctx, dbID, subjectName)
	require.NoError(t, err)
	require.Equal(t, &DatabaseKafkaSchemaRegistrySubject{
		SubjectName: subjectName,
		SchemaType:  "AVRO",
		Schema:      `{"type":"record","name":"test","fields":[{"name":"field1","type":"string"}]}`,
		SchemaID:    1,
	}, kafkaSchemaRegistrySubject)
}

func TestDatabases_DeleteKafkaSchemaRegistry(t *testing.T) {
	setup()
	defer teardown()

	dbID := "deadbeef-dead-4aa5-beef-deadbeef347d"
	subjectName := "test-subject"
	path := fmt.Sprintf("/v2/databases/%s/schema-registry/%s", dbID, subjectName)

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		w.WriteHeader(http.StatusNoContent)
	})

	_, err := client.Databases.DeleteKafkaSchemaRegistry(ctx, dbID, subjectName)
	require.NoError(t, err)
}

func TestDatabases_UpdateKafkaSchemaRegistryConfig(t *testing.T) {
	setup()
	defer teardown()

	dbID := "deadbeef-dead-4aa5-beef-deadbeef347d"
	path := fmt.Sprintf("/v2/databases/%s/schema-registry/config", dbID)

	updateConfig := &DatabaseKafkaSchemaRegistryConfig{
		CompatibilityLevel: "FULL",
	}

	responseJSON, err := json.Marshal(updateConfig)
	if err != nil {
		t.Fatal(err)
	}

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		w.WriteHeader(http.StatusOK)
		w.Write(responseJSON)
	})

	resp, _, err := client.Databases.UpdateKafkaSchemaRegistryConfig(ctx, dbID, updateConfig)
	require.NoError(t, err)
	require.Equal(t, updateConfig, resp)
}

func TestDatabases_GetKafkaSchemaRegistryConfig(t *testing.T) {
	setup()
	defer teardown()

	dbID := "deadbeef-dead-4aa5-beef-deadbeef347d"
	path := fmt.Sprintf("/v2/databases/%s/schema-registry/config", dbID)

	responseJSON, err := json.Marshal(&DatabaseKafkaSchemaRegistryConfig{
		CompatibilityLevel: "FULL",
	})
	if err != nil {
		t.Fatal(err)
	}

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.WriteHeader(http.StatusOK)
		w.Write(responseJSON)
	})

	resp, _, err := client.Databases.GetKafkaSchemaRegistryConfig(ctx, dbID)
	require.NoError(t, err)
	require.Equal(t, &DatabaseKafkaSchemaRegistryConfig{
		CompatibilityLevel: "FULL",
	}, resp)
}

func TestDatabases_UpdateKafkaSchemaRegistrySubjectConfig(t *testing.T) {
	setup()
	defer teardown()

	dbID := "deadbeef-dead-4aa5-beef-deadbeef347d"
	subjectName := "test-subject"
	path := fmt.Sprintf("/v2/databases/%s/schema-registry/config/%s", dbID, subjectName)

	updateConfig := &DatabaseKafkaSchemaRegistrySubjectConfigResponse{
		SubjectName:        subjectName,
		CompatibilityLevel: "FULL",
	}

	responseJSON, err := json.Marshal(updateConfig)
	if err != nil {
		t.Fatal(err)
	}

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		w.WriteHeader(http.StatusOK)
		w.Write(responseJSON)
	})

	resp, _, err := client.Databases.UpdateKafkaSchemaRegistrySubjectConfig(ctx, dbID, subjectName, &DatabaseKafkaSchemaRegistryConfig{
		CompatibilityLevel: "FULL",
	})
	require.NoError(t, err)
	require.Equal(t, subjectName, resp.SubjectName)
	require.Equal(t, updateConfig.CompatibilityLevel, resp.CompatibilityLevel)
}

func TestDatabases_GetKafkaSchemaRegistrySubjectConfig(t *testing.T) {
	setup()
	defer teardown()

	dbID := "deadbeef-dead-4aa5-beef-deadbeef347d"
	subjectName := "test-subject"
	path := fmt.Sprintf("/v2/databases/%s/schema-registry/config/%s", dbID, subjectName)

	responseJSON, err := json.Marshal(&DatabaseKafkaSchemaRegistrySubjectConfigResponse{
		SubjectName:        subjectName,
		CompatibilityLevel: "FULL",
	})
	if err != nil {
		t.Fatal(err)
	}

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.WriteHeader(http.StatusOK)
		w.Write(responseJSON)
	})

	resp, _, err := client.Databases.GetKafkaSchemaRegistrySubjectConfig(ctx, dbID, subjectName)
	require.NoError(t, err)
	require.Equal(t, subjectName, resp.SubjectName)
	require.Equal(t, "FULL", resp.CompatibilityLevel)
}
