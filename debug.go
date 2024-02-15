package incr

import (
	"fmt"
	"os"
)

var _isDebug = os.Getenv("INCR_DEBUG") != ""

func debugf(format string, args ...any) {
	if _isDebug {
		fmt.Fprintf(os.Stderr, "[DEBUG]::"+format+"\n", args...)
	}
}
