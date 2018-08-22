package logger

import (
	"log"
	"os"
)

var isVerbose bool
var std = log.New(os.Stdout, "", log.LstdFlags)
var dbg = log.New(os.Stdout, "[DEBUG]: ", log.LstdFlags)

// sets up if logging should be verbose/debug or not
func SetVerbose(verbose bool) {
	isVerbose = verbose
}

// returns global isDebug
func IsVerbose() bool {
	return isVerbose
}

// This functions deal with debug level taking verbose into account

// Print calls Output to print to the standard logger.
// Arguments are handled in the manner of fmt.Print.
func Debug(v ...interface{}) {
	if !isVerbose {
		return
	}
	dbg.Print(v...)
}

// Printf calls Output to print to the standard logger.
// Arguments are handled in the manner of fmt.Printf.
func Debugf(format string, v ...interface{}) {
	if !isVerbose {
		return
	}
	dbg.Printf(format, v...)
}

// Println calls Output to print to the standard logger.
// Arguments are handled in the manner of fmt.Println.
func Debugln(v ...interface{}) {
	if !isVerbose {
		return
	}
	dbg.Println(v...)
}

func Separator() {
	if !isVerbose {
		return
	}
	dbg.Println("-----------------------------------------------------------")
}

// These functions write to the standard logger.

// Print calls Output to print to the standard logger.
// Arguments are handled in the manner of fmt.Print.
func Print(v ...interface{}) {
	std.Print(v...)
}

// Printf calls Output to print to the standard logger.
// Arguments are handled in the manner of fmt.Printf.
func Printf(format string, v ...interface{}) {
	std.Printf(format, v...)
}

// Println calls Output to print to the standard logger.
// Arguments are handled in the manner of fmt.Println.
func Println(v ...interface{}) {
	std.Println(v...)
}

// Fatal is equivalent to Print() followed by a call to os.Exit(1).
func Fatal(v ...interface{}) {
	std.Fatal(v...)
}

// Fatalf is equivalent to Printf() followed by a call to os.Exit(1).
func Fatalf(format string, v ...interface{}) {
	std.Fatalf(format, v...)
}

// Fatalln is equivalent to Println() followed by a call to os.Exit(1).
func Fatalln(v ...interface{}) {
	std.Fatalln(v...)
}

// Panic is equivalent to Print() followed by a call to panic().
func Panic(v ...interface{}) {
	std.Panic(v...)
}

// Panicf is equivalent to Printf() followed by a call to panic().
func Panicf(format string, v ...interface{}) {
	std.Panicf(format, v...)
}

// Panicln is equivalent to Println() followed by a call to panic().
func Panicln(v ...interface{}) {
	std.Panicln(v...)
}