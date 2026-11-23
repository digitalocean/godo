// Package shared has the pieces that the Daytona and Modal compat clients both
// need: size mapping, the DigitalOcean client, and small conversions.
package shared

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/digitalocean/godo"
	"github.com/digitalocean/godo/sandbox"
	"github.com/digitalocean/godo/sandbox/compat/unsupported"
)

// Size is a DigitalOcean sandbox size.
type Size struct {
	Slug      string
	VCPUs     int
	MemoryMiB int
}

// Sizes is the catalog of sandbox sizes, smallest first. The Sandboxes API has
// no size listing, so this list is fixed: it comes from
// `doctl harness-runtime sizes list` (2026-11-05) and can go stale. The API
// rejects a size that is not enabled.
var Sizes = []Size{
	{"mars-1vcpu-1gb", 1, 1024},
	{"mars-2vcpu-2gb", 2, 2048},
	{"mars-2vcpu-4gb", 2, 4096},
	{"mars-4vcpu-8gb", 4, 8192},
	{"mars-16vcpu-32gb", 16, 32768},
}

// SizeFor returns the smallest size with at least the given CPU and memory.
// With no request (both zero) it returns "", which means the default size.
func SizeFor(vcpus float64, memoryMiB int) (string, error) {
	if vcpus <= 0 && memoryMiB <= 0 {
		return "", nil
	}
	for _, s := range Sizes {
		if float64(s.VCPUs) >= vcpus && s.MemoryMiB >= memoryMiB {
			return s.Slug, nil
		}
	}
	big := Sizes[len(Sizes)-1]
	return "", fmt.Errorf("no sandbox size has %g vCPU and %d MiB (the largest is %s)", vcpus, memoryMiB, big.Slug)
}

// SizeInfo looks a size slug up. An unknown slug returns false.
func SizeInfo(slug string) (Size, bool) {
	for _, s := range Sizes {
		if s.Slug == slug {
			return s, true
		}
	}
	return Size{}, false
}

// NewClient returns the DigitalOcean sandbox client. An empty token reads
// DIGITALOCEAN_ACCESS_TOKEN. The compat clients never read a vendor's
// credential variables, so a Daytona or Modal secret is never sent to
// DigitalOcean by accident.
func NewClient(token, apiURL string) (*sandbox.Client, error) {
	return sandbox.NewClient(&sandbox.Config{Token: token, APIURL: apiURL})
}

// TemplateName is a stable template name for an OCI reference, such as
// "daytona-3f2a9c1b7d4e".
func TemplateName(prefix, ref string) string {
	sum := sha256.Sum256([]byte(ref))
	return prefix + "-" + hex.EncodeToString(sum[:])[:12]
}

// ImageRef renders a template's image source back into one OCI reference.
func ImageRef(src *godo.SandboxImageSource) string {
	if src == nil {
		return ""
	}
	ref := src.Repository
	if src.Registry != "" {
		ref = src.Registry + "/" + src.Repository
	}
	switch {
	case src.Digest != "":
		return ref + "@" + src.Digest
	case src.Tag != "":
		return ref + ":" + src.Tag
	}
	return ref
}

// CeilSeconds rounds a duration up to whole seconds, as the commands API needs.
func CeilSeconds(d time.Duration) int64 {
	if d <= 0 {
		return 0
	}
	return int64((d + time.Second - 1) / time.Second)
}

// Option builds the error for a vendor option that has no DigitalOcean API yet.
// method is the vendor method and option the field or flag that was set.
func Option(platform, method, option, needs string) error {
	return unsupported.New(platform, fmt.Sprintf("%s (%s)", method, option), needs)
}

// TrimOutput removes one trailing newline, the way a shell substitution does.
func TrimOutput(s string) string { return strings.TrimRight(s, "\r\n") }

// Pipe is an in-memory byte pipe with an unbounded buffer. A writer never
// blocks on a slow reader, which is how Modal sandboxes behave: output waits on
// the server until you read it.
type Pipe struct {
	mu     sync.Mutex
	cond   *sync.Cond
	buf    []byte
	closed bool // no more writes: readers see io.EOF after the buffer drains
	gone   bool // the reader closed: writes are dropped
}

// NewPipe returns an empty Pipe.
func NewPipe() *Pipe {
	p := &Pipe{}
	p.cond = sync.NewCond(&p.mu)
	return p
}

// Write appends b. It drops the data, and reports success, once the reader closed.
func (p *Pipe) Write(b []byte) (int, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		return 0, io.ErrClosedPipe
	}
	if !p.gone {
		p.buf = append(p.buf, b...)
		p.cond.Broadcast()
	}
	return len(b), nil
}

// CloseWrite marks the end of the data.
func (p *Pipe) CloseWrite() {
	p.mu.Lock()
	p.closed = true
	p.cond.Broadcast()
	p.mu.Unlock()
}

// Read blocks until data arrives or the writer is done.
func (p *Pipe) Read(b []byte) (int, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	for len(p.buf) == 0 {
		if p.closed || p.gone {
			return 0, io.EOF
		}
		p.cond.Wait()
	}
	n := copy(b, p.buf)
	p.buf = p.buf[n:]
	return n, nil
}

// Close is the reader's Close: it discards what is buffered and unblocks reads.
func (p *Pipe) Close() error {
	p.mu.Lock()
	p.gone = true
	p.buf = nil
	p.cond.Broadcast()
	p.mu.Unlock()
	return nil
}
