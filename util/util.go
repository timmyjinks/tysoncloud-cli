package util

import "regexp"

var addrRegex = regexp.MustCompile("[0-9]*.[0-9]*.[0-9]*.[0-9]*")

func IntPtr(i int32) *int32      { return &i }
func StringPtr(s string) *string { return &s }

func IsAddr(addr string) bool {
	return addrRegex.MatchString(addr)
}
