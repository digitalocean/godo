package godo

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

const databaseBasePath = "/v2/databases"

// DatabasesService is an interface for interfacing with the databases endpoints
// of the DigitalOcean API.
// See: https://developers.digitalocean.com/documentation/v2#databases
type DatabasesService interface {
	List(context.Context, *ListOptions) ([]*Database, *Response, error)
	Get(context.Context, string) (*Database, *Response, error)
	Create(context.Context, *DatabaseCreateRequest) (*Database, *Response, error)
	Delete(context.Context, string) (*Response, error)
}

// DatabasesServiceOp handles communication with the Databases related methods
// of the DigitalOcean API.
type DatabasesServiceOp struct {
	client *Client
}

var _ DatabasesService = &DatabasesServiceOp{}

// Database represents a database cluster
type Database struct {
	ID                string                    `json:"id,omitempty"`
	Name              string                    `json:"name,omitempty"`
	Engine            string                    `json:"engine,omitempty"`
	Version           string                    `json:"version,omitempty"`
	Connection        DatabaseConnection        `json:"connection,omitempty"`
	Users             []DatabaseUser            `json:"users,omitempty"`
	NumNodes          int                       `json:"num_nodes,omitempty"`
	SizeSlug          string                    `json:"size,omitempty"` // TODO(ez): Should this be a Size struct?
	DBNames           []string                  `json:"db_names,omitempty"`
	RegionSlug        string                    `json:"region,omitempty"`
	Status            string                    `json:"status,omitempty"` // TODO(ez): Should this be a struct?
	MaintenanceWindow DatabaseMaintenanceWindow `json:"maintenance_window,omitempty"`
	CreatedAt         time.Time                 `json:"created_at,omitempty"`
}

// DatabaseConnection represents a database connection
type DatabaseConnection struct {
	URI      string `json:"uri,omitempty"`
	Database string `json:"database,omitempty"`
	Host     string `json:"host,omitempty"`
	Port     int    `json:"port,omitempty"`
	User     string `json:"user,omitempty"`
	Password string `json:"password,omitempty"`
	SSL      bool   `json:"ssl,omitempty"`
}

// DatabaseUser represents a user in the database
type DatabaseUser struct {
	Name     string `json:"name,omitempty"`
	Role     string `json:"role,omitempty"`
	Password string `json:"password,omitempty"`
}

// DatabaseMaintenanceWindow represents the maintenance_window of a database
// cluster
type DatabaseMaintenanceWindow struct {
	Day         string   `json:"day,omitempty"` // TODO(ez): Should this be a day type?
	Hour        string   `json:"hour,omitempty"`
	Pending     bool     `json:"pending,omitempty"`
	Description []string `json:"description,omitempty"`
}

// DatabaseCreateRequest represents a request to create a database cluster
type DatabaseCreateRequest struct {
	Name     string `json:"name,omitempty"`
	Engine   string `json:"engine,omitempty"`
	Version  string `json:"version,omitempty"`
	Size     string `json:"size,omitempty"`
	Region   string `json:"region,omitempty"`
	NumNodes int    `json:"num_nodes,omitempty"`
}

type databasesRoot struct {
	Databases []*Database `json:"databases,omitempty"`
}

type databaseRoot struct {
	Database *Database `json:"database,omitempty"`
}

// List returns a list of the Databases visible with the caller's API token
func (svc *DatabasesServiceOp) List(ctx context.Context, opts *ListOptions) ([]*Database, *Response, error) {
	path := databaseBasePath
	path, err := addOptions(path, opts)
	if err != nil {
		return nil, nil, err
	}
	req, err := svc.client.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}
	root := new(databasesRoot)
	resp, err := svc.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}
	return root.Databases, resp, nil
}

// Get retrieves the details of a database cluster
func (svc *DatabasesServiceOp) Get(ctx context.Context, databaseID string) (*Database, *Response, error) {
	path := fmt.Sprintf("%s/%s", databaseBasePath, databaseID)
	req, err := svc.client.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}
	root := new(databaseRoot)
	resp, err := svc.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}
	return root.Database, resp, nil
}

// Create creates a database cluster
func (svc *DatabasesServiceOp) Create(ctx context.Context, create *DatabaseCreateRequest) (*Database, *Response, error) {
	path := databaseBasePath
	req, err := svc.client.NewRequest(ctx, http.MethodPost, path, create)
	if err != nil {
		return nil, nil, err
	}
	root := new(databaseRoot)
	resp, err := svc.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}
	return root.Database, resp, nil
}

// Delete deletes a database cluster. There is no way to recover a cluster once
// it has been destroyed.
func (svc *DatabasesServiceOp) Delete(ctx context.Context, databaseID string) (*Response, error) {
	path := fmt.Sprintf("%s/%s", databaseBasePath, databaseID)
	req, err := svc.client.NewRequest(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return nil, err
	}
	resp, err := svc.client.Do(ctx, req, nil)
	if err != nil {
		return resp, err
	}
	return resp, nil
}
