package godo

import (
	"context"
	"fmt"
	"net/http"
)

// NfsActionsService is an interface for interacting with the NFS actions
// endpoints of the DigitalOcean API
// See: https://docs.digitalocean.com/reference/api/api-reference/#tag/NFS-Actions
type NfsActionsService interface {
	Resize(ctx context.Context, nfsShareId string, size uint64, region string) (*Action, *Response, error)
	Snapshot(ctx context.Context, nfsShareId string, nfsSnapshotName string, region string) (*Action, *Response, error)
}

// NfsActionsServiceOp handles communication with the NFS action related
// methods of the DigitalOcean API.
type NfsActionsServiceOp struct {
	client *Client
}

var _ NfsActionsService = &NfsActionsServiceOp{}

// Resize an NFS share
func (s *NfsActionsServiceOp) Resize(ctx context.Context, nfsShareId string, size uint64, region string) (*Action, *Response, error) {
	requestType := "resize"
	request := &ActionRequest{
		"type":   requestType,
		"region": region,
		"size":   size,
	}
	return s.doAction(ctx, nfsShareId, request)
}

// Snapshot an NFS share
func (s *NfsActionsServiceOp) Snapshot(ctx context.Context, nfsShareId, nfsSnapshotName, region string) (*Action, *Response, error) {
	requestType := "snapshot"
	request := &ActionRequest{
		"type":   requestType,
		"name":   nfsSnapshotName,
		"region": region,
	}
	return s.doAction(ctx, nfsShareId, request)
}

func (s *NfsActionsServiceOp) doAction(ctx context.Context, nfsShareId string, request *ActionRequest) (*Action, *Response, error) {
	if request == nil {
		return nil, nil, NewArgError("request", "request can't be nil")
	}

	path := nfsActionPath(nfsShareId)

	req, err := s.client.NewRequest(ctx, http.MethodPost, path, request)
	if err != nil {
		return nil, nil, err
	}

	root := new(actionRoot)
	resp, err := s.client.Do(ctx, req, root)
	if err != nil {
		return nil, resp, err
	}

	return root.Event, resp, err
}

func nfsActionPath(nfsID string) string {
	return fmt.Sprintf("v2/nfs/%v/actions", nfsID)
}
