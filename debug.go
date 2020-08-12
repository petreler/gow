package gow

import (
	"fmt"
	"os"
)

// DebugPrintError
func DebugPrintError(err error) {
	debugPrintError(err)
}

// DebugPrint
func DebugPrint(format string, v ...interface{}) {
	debugPrint(format, v)
}


//==============private=================

// debugPrintError
func debugPrintError(err error) {
	fmt.Fprintf(os.Stdout, "[debugError] %v \n", err)
}

// debugPrint
func debugPrint(format string, v ...interface{}) {
	fmt.Fprintf(os.Stdout, "[debugPrint] "+format+"\n", v...)
}
