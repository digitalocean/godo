package modal

import "fmt"

// argMax is the largest total size of command arguments Modal accepts.
const argMax = 2 << 20

// ValidateExecArgs mirrors modal.ValidateExecArgs: it checks that the command
// arguments are not larger than ARG_MAX.
func ValidateExecArgs(args []string) error {
	total := 0
	for _, a := range args {
		total += len(a) + 1
	}
	if total > argMax {
		return fmt.Errorf("modal: command arguments total %d bytes, over the limit of %d", total, argMax)
	}
	return nil
}
