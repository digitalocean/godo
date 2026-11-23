// Command gen mirrors the exported API of a vendor Go package as declarations
// plus stub bodies that return *unsupported.NotImplementedError. It parses the
// vendor source with go/parser and never runs it. See generate.sh.
package main
