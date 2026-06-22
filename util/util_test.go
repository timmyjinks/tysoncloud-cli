package util

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/timmyjinks/tysoncloud-cli/store"
)

func TestIsAddr(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"valid ip", "192.168.1.1", true},
		{"valid ip zeros", "0.0.0.0", true},
		{"valid ip max", "255.255.255.255", true},
		{"invalid ip too high", "256.0.0.1", false},
		{"invalid ip letters", "abc.def.ghi.jkl", false},
		{"invalid ip missing octet", "192.168.1", false},
		{"invalid ip extra octet", "192.168.1.1.1", false},
		{"empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, IsAddr(tt.input))
		})
	}
}

func TestToEnvMap(t *testing.T) {
	tests := []struct {
		name     string
		input    []store.EnvironmentsTable
		expected map[string][]byte
	}{
		{
			name: "basic conversion",
			input: []store.EnvironmentsTable{
				{Key: "PORT", Val: "5432"},
				{Key: "DB", Val: "postgres"},
			},
			expected: map[string][]byte{
				"PORT": []byte("5432"),
				"DB":   []byte("postgres"),
			},
		},
		{
			name:     "empty slice",
			input:    []store.EnvironmentsTable{},
			expected: map[string][]byte{},
		},
		{
			name: "empty value",
			input: []store.EnvironmentsTable{
				{Key: "SECRET", Val: ""},
			},
			expected: map[string][]byte{
				"SECRET": []byte(""),
			},
		},
		{
			name: "duplicate key takes last value",
			input: []store.EnvironmentsTable{
				{Key: "PORT", Val: "5432"},
				{Key: "PORT", Val: "9999"},
			},
			expected: map[string][]byte{
				"PORT": []byte("9999"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, ToEnvMap(tt.input))
		})
	}
}

func TestToEnvString(t *testing.T) {
	tests := []struct {
		name     string
		input    []store.EnvironmentsTable
		expected string
	}{
		{
			name: "sorts alphabetically",
			input: []store.EnvironmentsTable{
				{Key: "PORT", Val: "5432"},
				{Key: "DB", Val: "postgres"},
				{Key: "USER", Val: "admin"},
			},
			expected: "DB=postgres,PORT=5432,USER=admin",
		},
		{
			name: "single entry",
			input: []store.EnvironmentsTable{
				{Key: "PORT", Val: "5432"},
			},
			expected: "PORT=5432",
		},
		{
			name:     "empty slice",
			input:    []store.EnvironmentsTable{},
			expected: "",
		},
		{
			name: "empty value",
			input: []store.EnvironmentsTable{
				{Key: "SECRET", Val: ""},
				{Key: "PORT", Val: "5432"},
			},
			expected: "PORT=5432,SECRET=",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, ToEnvString(tt.input))
		})
	}
}

func TestCompareDiff(t *testing.T) {
	tests := []struct {
		name      string
		file1     string
		file2     string
		deletions []string
		additions []string
	}{
		{
			name:      "no changes",
			file1:     "proj-abc\nsvc-xyz default\n",
			file2:     "proj-abc\nsvc-xyz default\n",
			deletions: nil,
			additions: nil,
		},
		{
			name:      "line added",
			file1:     "proj-abc\n",
			file2:     "proj-abc\nsvc-xyz default\n",
			deletions: nil,
			additions: []string{"svc-xyz default"},
		},
		{
			name:      "line removed",
			file1:     "proj-abc\nsvc-xyz default\n",
			file2:     "proj-abc\n",
			deletions: []string{"svc-xyz default"},
			additions: nil,
		},
		{
			name:      "both added and removed",
			file1:     "proj-abc\nsvc-old default\n",
			file2:     "proj-abc\nsvc-new default\n",
			deletions: []string{"svc-old default"},
			additions: []string{"svc-new default"},
		},
		{
			name:      "empty files",
			file1:     "",
			file2:     "",
			deletions: nil,
			additions: nil,
		},
		{
			name:      "blank lines ignored",
			file1:     "proj-abc\n\n",
			file2:     "proj-abc\n",
			deletions: nil,
			additions: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deletions, additions := CompareDiff(tt.file1, tt.file2)
			assert.ElementsMatch(t, tt.deletions, deletions)
			assert.ElementsMatch(t, tt.additions, additions)
		})
	}
}

func TestGetEnv(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		fallback string
		envVal   string
		expected string
	}{
		{"returns env value", "TEST_KEY", "fallback", "actual", "actual"},
		{"returns fallback when empty", "TEST_KEY", "fallback", "", "fallback"},
		{"returns fallback when unset", "TEST_KEY_UNSET", "fallback", "", "fallback"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envVal != "" {
				os.Setenv(tt.key, tt.envVal)
				defer os.Unsetenv(tt.key)
			}
			assert.Equal(t, tt.expected, GetEnv(tt.key, tt.fallback))
		})
	}
}

func TestReadWriteFile(t *testing.T) {
	tmp, err := os.CreateTemp("", "tysoncloud-test-*.txt")
	assert.NoError(t, err)
	defer os.Remove(tmp.Name())
	tmp.Close()

	content := "proj-abc\nsvc-xyz default\n"
	err = WriteFile(content, tmp.Name())
	assert.NoError(t, err)

	result, err := ReadFile(tmp.Name())
	assert.NoError(t, err)
	assert.Equal(t, content, result)
}
