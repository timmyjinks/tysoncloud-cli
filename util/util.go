package util

import "regexp"

var addrRegex = regexp.MustCompile(`^(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}$`)

func IntPtr(i int32) *int32      { return &i }
func StringPtr(s string) *string { return &s }

func IsAddr(addr string) bool {
	return addrRegex.MatchString(addr)
}
