package util

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/timmyjinks/tysoncloud-cli/store"
)

var addrRegex = regexp.MustCompile(`^(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}$`)

func IntPtr(i int32) *int32      { return &i }
func StringPtr(s string) *string { return &s }

func IsAddr(addr string) bool {
	return addrRegex.MatchString(addr)
}

func ToEnvMap(envs []store.EnvironmentsTable) map[string][]byte {
	envsMap := make(map[string][]byte)

	for _, env := range envs {
		envsMap[env.Key] = []byte(env.Val)
	}

	return envsMap
}

func ToEnvString(envs []store.EnvironmentsTable) string {
	sort.Slice(envs, func(i, j int) bool {
		return envs[i].Key < envs[j].Key
	})
	var parts []string
	for _, e := range envs {
		parts = append(parts, fmt.Sprintf("%s=%s", e.Key, e.Val))
	}
	return strings.Join(parts, ",")
}

func CompareDiff(file1, file2 string) (deletions []string, additions []string) {
	oldSet := toSet(file1)
	newSet := toSet(file2)

	var removed []string
	var added []string

	for id := range oldSet {
		if _, ok := newSet[id]; !ok {
			removed = append(removed, id)
		}
	}

	for id := range newSet {
		if _, ok := oldSet[id]; !ok {
			added = append(added, id)
		}
	}

	return removed, added
}

func toSet(s string) map[string]struct{} {
	set := make(map[string]struct{})
	scanner := bufio.NewScanner(strings.NewReader(s))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			set[line] = struct{}{}
		}
	}
	return set
}

func ReadFile(filename string) (string, error) {
	var lines strings.Builder
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_RDONLY, 0600)
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines.WriteString(scanner.Text() + "\n")
	}

	return lines.String(), scanner.Err()
}

func WriteFile(content, filename string) error {
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err := file.WriteString(content); err != nil {
		return err
	}
	return nil
}

func GetEnv(key, fallback string) string {
	env := os.Getenv(key)
	if env == "" {
		return fallback
	}
	return env
}

func PrintDiff(added, removed []string) {
	for _, id := range removed {
		fmt.Println("-", id)
	}

	for _, id := range added {
		fmt.Println("+", id)
	}
}
