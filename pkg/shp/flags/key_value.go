package flags

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"strings"
)

// splitKeyValue splits a key-value string with the format "key=value". Returns the key, value, and
// an error if the string is not formatted correctly.
func splitKeyValue(value string) (string, string, error) {
	s := strings.SplitN(value, "=", 2)
	if len(s) != 2 || s[0] == "" {
		return "", "", fmt.Errorf("informed value '%s' is not in key=value format", value)
	}
	return s[0], s[1], nil
}

// writeAsCSV returns the informed slice of strings as a single string split by comma (",").
func writeAsCSV(slice []string) (string, error) {
	b := &bytes.Buffer{}
	w := csv.NewWriter(b)
	if err := w.Write(slice); err != nil {
		return "", err
	}
	w.Flush()
	return strings.TrimSuffix(b.String(), "\n"), nil
}
