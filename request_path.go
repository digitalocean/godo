package godo

import (
	"fmt"
	"net/url"
	"strings"
)

// resolveRequestURL resolves urlStr against base after checking its path with
// checkRequestPath.
func resolveRequestURL(base *url.URL, urlStr string) (*url.URL, error) {
	rel, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}
	if err := checkRequestPath(rel); err != nil {
		return nil, err
	}
	return base.ResolveReference(rel), nil
}

// checkRequestPath rejects request paths that would not reach the route they
// were built for. Resolving against a base URL removes "." and ".." segments,
// and servers may decode and remove percent-encoded forms of them, so a value
// such as "../../v2/account/keys/1" used as an ID would send the request to a
// different resource. Empty segments come from empty IDs and would address
// the parent collection instead.
func checkRequestPath(u *url.URL) error {
	p := strings.TrimPrefix(u.EscapedPath(), "/")
	if p == "" {
		return nil
	}
	for _, seg := range strings.Split(p, "/") {
		if seg == "" {
			return NewArgError("path", fmt.Sprintf("%q contains an empty segment", u.Path))
		}
		decoded, err := url.PathUnescape(seg)
		if err != nil {
			return err
		}
		if strings.IndexFunc(decoded, isControlChar) >= 0 {
			return NewArgError("path", fmt.Sprintf("%q contains a control character", u.Path))
		}
		for _, part := range strings.FieldsFunc(decoded, isPathSeparator) {
			if part == "." || part == ".." {
				return NewArgError("path", fmt.Sprintf("%q contains a %q segment", u.Path, part))
			}
		}
	}
	return nil
}

func isPathSeparator(r rune) bool {
	return r == '/' || r == '\\'
}

func isControlChar(r rune) bool {
	return r < 0x20 || r == 0x7f
}
