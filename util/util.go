package util

import (
	"regexp"

	"github.com/timmyjinks/tysoncloud-cli/store"
)

var addrRegex = regexp.MustCompile(`^(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}$`)

func IntPtr(i int32) *int32      { return &i }
func StringPtr(s string) *string { return &s }

func IsAddr(addr string) bool {
	return addrRegex.MatchString(addr)
}

func ToEnvString(envs []store.EnvironmentsTable) map[string][]byte {
	envsMap := make(map[string][]byte)

	for _, env := range envs {
		envsMap[env.Key] = []byte(env.Val)
	}

	return envsMap
}
