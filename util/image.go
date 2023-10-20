package util

import (
	"context"
	"fmt"
	"time"

	"github.com/digitalocean/godo"
)

const (
	// availableFailure is the amount of times we can fail before deciding
	// the check for available is a total failure. This can help account
	// for servers randomly not answering.
	availableFailure = 3
)

// WaitForAvailable waits for an image to become available
func WaitForAvailable(ctx context.Context, client *godo.Client, monitorURI string) error {
	if len(monitorURI) == 0 {
		return fmt.Errorf("create had no monitor URI")
	}

	failCount := 0
	actionCh := make(chan *godo.Action)
	errCh := make(chan error)

	go func() {
		for {
			action, _, err := client.ImageActions.GetByURI(ctx, monitorURI)
			if err != nil {
				errCh <- err
				return
			}
			actionCh <- action
			time.Sleep(5 * time.Second)
		}
	}()

	for {
		select {
		case action := <-actionCh:
			switch action.Status {
			case godo.ActionInProgress:
				// Continue waiting
			case godo.ActionCompleted:
				return nil
			default:
				return fmt.Errorf("unknown status: [%s]", action.Status)
			}

		case err := <-errCh:
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			if failCount <= availableFailure {
				failCount++
			} else {
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
