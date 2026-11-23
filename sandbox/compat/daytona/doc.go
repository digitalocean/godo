// Package daytona is a migration aid for code written against the Daytona Go
// SDK. Change the import path and keep calling the same methods; underneath,
// the calls go to the DigitalOcean Sandboxes API.
//
// It mirrors github.com/daytonaio/daytona/libs/sdk-go v0.190.0, package
// pkg/daytona, with the sibling packages pkg/types, pkg/options and pkg/errors
// as the sub-packages types, options and errors. The API client packages the
// SDK exposes (api-client-go and toolbox-api-client-go) are mirrored as the
// sub-packages apiclient and toolbox: only the types the SDK surface needs, with
// no accessor methods.
//
// Where Daytona and DigitalOcean agree, the method works:
//
//   - Client.Create, Get, List and ListSeq; Sandbox.Start, Stop, Pause, Delete,
//     RefreshData and the wait methods.
//   - Process.ExecuteCommand: a foreground command with its working directory,
//     environment and timeout.
//   - Snapshots, which are Daytona's reusable environments, are DigitalOcean
//     templates: Snapshot.Create builds a template from an OCI image, and
//     Client.Create with types.SnapshotParams starts a sandbox from it.
//   - Sandbox.ExperimentalCreateSnapshot captures a sandbox as a DigitalOcean
//     snapshot.
//
// Everything else keeps its signature and returns a
// *unsupported.NotImplementedError (see the compat package), which matches
// errors.Is(err, compat.ErrNotImplemented). A method with no error result
// panics with that error instead. A parameter that has no DigitalOcean
// equivalent, such as Labels or AutoStopInterval, fails the call when it is
// set, so nothing is silently ignored.
//
// Authentication: NewClient reads only DIGITALOCEAN_ACCESS_TOKEN. It never
// reads DAYTONA_API_KEY, so a Daytona credential is never sent to
// DigitalOcean. NewClientWithConfig uses DaytonaConfig.APIKey as the
// DigitalOcean token and DaytonaConfig.APIUrl as the API origin when they are
// set.
package daytona
