package daytona

import "github.com/digitalocean/godo/sandbox"

// SandboxIterator mirrors daytona.SandboxIterator.
type SandboxIterator struct {
	client  *Client
	pull    func() (*sandbox.Sandbox, error, bool)
	stop    func()
	keep    func(*sandbox.Sandbox) bool
	current *Sandbox
	err     error
}

// Next mirrors SandboxIterator.Next.
func (it *SandboxIterator) Next() bool {
	if it.err != nil || it.pull == nil {
		return false
	}
	for {
		sb, err, ok := it.pull()
		if !ok {
			it.close()
			return false
		}
		if err != nil {
			it.err = convertError(err)
			it.close()
			return false
		}
		if it.keep != nil && !it.keep(sb) {
			continue
		}
		it.current = newSandbox(it.client, sb)
		return true
	}
}

// Value mirrors SandboxIterator.Value.
func (it *SandboxIterator) Value() *Sandbox { return it.current }

// Err mirrors SandboxIterator.Err.
func (it *SandboxIterator) Err() error { return it.err }

func (it *SandboxIterator) close() {
	if it.stop != nil {
		it.stop()
		it.stop, it.pull = nil, nil
	}
}
