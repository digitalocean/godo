package fakeapi

import "encoding/base64"

func b64(s string) string { return base64.StdEncoding.EncodeToString([]byte(s)) }
