package flags

import (
	"fmt"
	"strings"
)

func CheckError(err error, functionName string, wantErr string) string {
	if err != nil {
		if len(wantErr) != 0 {
			if !strings.Contains(err.Error(), wantErr) {
				return fmt.Sprintf("expected error %q from %q function, but it was not found", wantErr, functionName)
			}
		} else {
			return fmt.Sprintf("unexpected error occurred from %q function: %#v", functionName, err)
		}
	} else if err == nil {
		if len(wantErr) != 0 {
			return fmt.Sprintf("expected error %q from %q function, but it was not found", wantErr, functionName)
		}
	}

	return ""
}
