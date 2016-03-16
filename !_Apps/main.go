package main

import (
//	"datatree"
	"fmt"
	"github.com/eiannone/keyboard"
//	"github.com/fatih/color"
	"log"
	"os"
	"path/filepath"
	"strings"
)

const (
	fDir       = "_tmp/log/"
	fExt       = ".log"
)

var flags struct {
	Debug bool
	Verbose bool
}

func NewLogFile(runFileName string) (logFile *os.File, err error) {
	err = os.MkdirAll(fDir, os.ModeDir)
	logFileName := filepath.Base(runFileName)
	logFileName  = strings.Replace(logFileName, filepath.Ext(logFileName), "", -1)
	logFileName  = fDir + logFileName + fExt
	logFile, err = os.Create(logFileName)
	if err != nil { panic(err) }
	return
}

func LogF(format string, args ...interface{}) {
	log.Printf(format, args...)
	if flags.Debug {
		fmt.Printf(format, args...)
	}
}

func LogLn(args ...interface{}) {
	log.Println(args...)
	if flags.Debug {
		fmt.Println(args...)
	}
}

func Print(args ...interface{}) {
	fmt.Print(args...)
}

func PrintF(format string, args ...interface{}) {
	fmt.Printf(format, args...)
}

func PrintLn(args ...interface{}) {
	fmt.Println(args...)
}

func main() {

// logfailo sukūrimas
	logFile, err := NewLogFile(os.Args[0])
	defer logFile.Close()
	log.SetOutput(logFile)


// logfailo pildymo pradžia
	LogLn(os.Args[0], "started")
	LogLn(bigseparator)
	
/*	//red := color.New(color.FgRed).SprintFunc() */

	PrintLn()
	LogLn("This is log file of", os.Args[0])
	LogLn(bigseparator)
	
	args := os.Args
	for _, arg := range args {
        switch arg {
		case "--debug" :
            flags.Debug = true
		case "--verbose" :
			flags.Verbose = true
        }
    }
	LogLn("args:", args, "\n")
	LogLn("flags.Debug:",   flags.Debug, "\n")
	LogLn("flags.Verbose:", flags.Verbose, "\n")

// darbo pradžia

	var confKeeper ConfKeeper
	err = confKeeper.Run()
	if err != nil {
		PrintLn(err.Error(), "\n")
	}
	PrintLn("Press any key to exit...")
	keyboard.GetSingleKey()

}