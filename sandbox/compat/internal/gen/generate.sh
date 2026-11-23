#!/usr/bin/env bash
# Regenerates the mirrored declarations of the Daytona and Modal compat packages
# from the vendor sources. Run it from the godo module root:
#
#   sandbox/compat/internal/gen/generate.sh
#
# It reads the vendor modules from the Go module cache (it never runs their
# code). Download them first, from any scratch module:
#
#   go mod download github.com/daytonaio/daytona/libs/sdk-go@v0.190.0 \
#     github.com/daytonaio/daytona/libs/api-client-go@v0.190.0 \
#     github.com/daytonaio/daytona/libs/toolbox-api-client-go@v0.190.0 \
#     github.com/modal-labs/modal-client/go@v0.11.2
#
# Bumping a version here means re-checking the hand-written files next to the
# generated ones (client.go, sandbox.go and so on) against the new API.
set -euo pipefail
cd "$(dirname "$0")/../../../.."
ROOT=$PWD
MC=$(go env GOMODCACHE)
DT_VER=${DT_VER:-v0.190.0}
MODAL_VER=${MODAL_VER:-v0.11.2}
D=$MC/github.com/daytonaio/daytona/libs
M=$MC/github.com/modal-labs/modal-client/go@$MODAL_VER
GEN=$(mktemp -d)/gen
go build -o "$GEN" ./sandbox/compat/internal/gen
RULES=$ROOT/sandbox/compat/internal/gen

# ---- Daytona
S=$D/sdk-go@$DT_VER/pkg
O=$ROOT/sandbox/compat/daytona
P=github.com/digitalocean/godo/sandbox/compat/daytona
"$GEN" -mode copy -src "$S/errors" -pkg errors -vendor github.com/daytonaio/daytona/libs/sdk-go/pkg/errors -version "$DT_VER" -drop ConvertAPIError,ConvertToolboxError -out "$O/errors/errors.go"
"$GEN" -mode copy -src "$S/options" -pkg options -vendor github.com/daytonaio/daytona/libs/sdk-go/pkg/options -version "$DT_VER" -importmap types=$P/types -out "$O/options/options.go"
"$GEN" -mode copy -src "$S/types" -pkg types -vendor github.com/daytonaio/daytona/libs/sdk-go/pkg/types -version "$DT_VER" -importmap apiclient=$P/apiclient,toolbox=$P/toolbox -out "$O/types/types.go"
# NewRecordingService takes the toolbox API client, which is not mirrored.
IMPL="NewRecordingService,NewClient,NewClientWithConfig,WithDownloadProgress,WithUploadProgress,\
Client.Close,Client.Create,Client.Get,Client.List,Client.ListSeq,\
Sandbox.Start,Sandbox.StartWithTimeout,Sandbox.Stop,Sandbox.StopWithTimeout,Sandbox.Pause,Sandbox.PauseWithTimeout,\
Sandbox.Delete,Sandbox.DeleteWithTimeout,Sandbox.RefreshData,Sandbox.WaitForStart,Sandbox.WaitForStop,\
Sandbox.ExperimentalCreateSnapshot,Sandbox.ExperimentalCreateSnapshotWithTimeout,Sandbox.GetUserHomeDir,Sandbox.GetWorkingDir,\
ProcessService.ExecuteCommand,SandboxIterator.Next,SandboxIterator.Value,SandboxIterator.Err,\
SnapshotService.Create,SnapshotService.Get,SnapshotService.List,SnapshotService.Delete"
"$GEN" -mode stubs -src "$S/daytona" -pkg daytona -product Daytona -vendor github.com/daytonaio/daytona/libs/sdk-go/pkg/daytona -version "$DT_VER" \
  -rules "$RULES/daytona.rules" -hwtypes Client,Sandbox,SandboxIterator,ProcessService,SnapshotService \
  -allow downloadStreamConfig,uploadStreamConfig -impl "$IMPL" \
  -importmap apiclient=$P/apiclient,toolbox=$P/toolbox,types=$P/types,options=$P/options,errors=$P/errors \
  -out "$O/zz_generated.go" 2> "$GEN.daytona.report" || true
# The api and toolbox clients: only the types the mirrored packages mention.
cat "$O/types/types.go" "$O/options/options.go" "$O"/*.go > "$GEN.all"
rootsOf() { grep -oE "\b$1\.[A-Za-z_0-9]+" "$GEN.all" | sed "s/^$1\.//" | grep -vE '^[A-Z0-9]+_[A-Z0-9_]+$|^APIClient$' | sort -u | tr '\n' ',' | sed 's/,$//'; }
"$GEN" -mode closure -src "$D/api-client-go@$DT_VER" -pkg apiclient -vendor github.com/daytonaio/daytona/libs/api-client-go -version "$DT_VER" -roots "$(rootsOf apiclient)" -out "$O/apiclient/types.go" 2>> "$GEN.daytona.report"
"$GEN" -mode closure -src "$D/toolbox-api-client-go@$DT_VER" -pkg toolbox -vendor github.com/daytonaio/daytona/libs/toolbox-api-client-go -version "$DT_VER" -roots "$(rootsOf toolbox)" -out "$O/toolbox/types.go" 2>> "$GEN.daytona.report"

# ---- Modal
O=$ROOT/sandbox/compat/modal
IMPL="Client.Close,Client.Version,NewClient,NewClientWithOptions,Image.Build,Image.DockerfileCommands,\
Sandbox.Exec,Sandbox.Poll,Sandbox.Wait,Sandbox.Terminate,Sandbox.Detach,Sandbox.WaitUntilReady,\
Sandbox.ExperimentalPause,Sandbox.ExperimentalResume,Sandbox.ExperimentalSnapshot,ContainerProcess.Wait,\
NewExperimentalOutboundPolicy,ExperimentalOutboundPolicy.WithHeaderReplacement,ValidateExecArgs"
"$GEN" -mode stubs -src "$M" -pkg modal -product Modal -vendor github.com/modal-labs/modal-client/go -version "$MODAL_VER" \
  -rules "$RULES/modal.rules" -hwtypes Client,ClientParams,App,Image,Secret,Sandbox,ContainerProcess,SandboxSnapshot,ExperimentalOutboundPolicy \
  -stubifaces AppService,CloudBucketMountService,ClsService,EnvironmentObjectsService,EnvironmentAppsService,FunctionService,FunctionCallService,ImageService,ProxyService,QueueService,SandboxService,SandboxSnapshotService,SidecarService,SecretService,VolumeService \
  -impl "$IMPL" -verbatim errors.go -verbatimout "$O/errors.go" -out "$O/zz_generated.go" 2> "$GEN.modal.report" || true
gofmt -w "$ROOT/sandbox/compat"
echo "reports: $GEN.daytona.report $GEN.modal.report"
