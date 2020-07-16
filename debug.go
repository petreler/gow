package gow

import (
	"fmt"
	"os"
)

func DebugPrintError(err error) {
	debugPrintError(err)
}

// debugPrintError
func debugPrintError(err error) {
	fmt.Fprintf(os.Stdout, "[debugError] %v \n", err)
}

// debugPrint
func debugPrint(format string, v ...interface{}) {
	fmt.Fprintf(os.Stdout, "[debugPrint] "+format+"\n", v...)
}
