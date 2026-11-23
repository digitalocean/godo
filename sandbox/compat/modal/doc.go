// Package modal is a migration aid for code written against the Modal Go SDK.
// Change the import path and keep calling the same methods; underneath, the
// calls go to the DigitalOcean Sandboxes API.
//
// It mirrors github.com/modal-labs/modal-client/go v0.11.2 (the module that
// replaced github.com/modal-labs/libmodal/modal-go). ClientParams keeps only
// the exported fields that are not gRPC plumbing: the control-plane client, the
// connection, the interceptors and Config are left out.
//
// Where Modal and DigitalOcean agree, the call works:
//
//   - Client.Apps.FromName, Client.Images.FromRegistry and Image.Build,
//     Client.Secrets.FromMap.
//   - Client.Sandboxes.Create and ExperimentalCreate, FromID, List.
//   - Sandbox.Exec with its stdin, stdout and stderr streams, Poll, Wait,
//     Terminate, Detach, WaitUntilReady, ExperimentalPause and
//     ExperimentalResume, and ExperimentalSnapshot.
//   - An Image from a registry becomes a DigitalOcean template built from that
//     OCI image. Its RUN, ENV and WORKDIR layers are replayed in the sandbox
//     after it starts, so a create with layers is slower.
//
// Everything else keeps its signature and returns a
// *unsupported.NotImplementedError (see the compat package), which matches
// errors.Is(err, compat.ErrNotImplemented). A method with no error result
// panics with that error instead. A parameter that has no DigitalOcean
// equivalent, such as Timeout, Volumes or Tags, fails the call when it is set,
// so nothing is silently ignored.
//
// Differences that are not errors: the sandbox does not end when its Command
// exits, Terminate and Wait report exit code 0 when no entrypoint ran, and
// Environment settings have no effect.
//
// Authentication: NewClient reads only DIGITALOCEAN_ACCESS_TOKEN. It never
// reads MODAL_TOKEN_ID, MODAL_TOKEN_SECRET or ~/.modal.toml, so a Modal
// credential is never sent to DigitalOcean. NewClientWithOptions uses
// ClientParams.TokenSecret as the DigitalOcean token when it is set.
package modal
