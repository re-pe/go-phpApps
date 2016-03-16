package main

import (
	"fmt"
	"github.com/eiannone/keyboard"
	"log"
	"os"
	"path/filepath"
	"strings"
)

const (
	fRunningFile  = "_Settings.exe"
	fDir          = "_tmp/log/"
	fExt          = ".log"
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

func ExitLn(args ...interface{}) {
	fmt.Println(args...)
	log.Fatalln(args...)
}

func CheckFormat(args []interface{}) (format string, hasFormat bool) {
	if len(args) < 1 { return }
	switch arg := args[0].(type){
	case string: 
		format = arg
		hasFormat = true
	}
	if !hasFormat { return }

	if len(format) < 1 || strings.Index(format, "?:") != 0 {
		hasFormat = false
		return
	}
	format = strings.Replace(format, "?:", "", 1)
	return
}

func Print(args ...interface{}) {
	if format, hasFormat := CheckFormat(args); hasFormat {
		fmt.Printf(format, args[1:]...)
	} else {
		fmt.Print(args...)
	}
}

func Log(args ...interface{}) {
	if format, hasFormat := CheckFormat(args); hasFormat {
		log.Printf(format, args[1:]...)
	} else {
		log.Print(args...)
	}
}

func Out(args ...interface{}) {
	Log(args...)
	Print(args...)
}

func Debug(args ...interface{}) {
	if flags.Debug {
		Out(args...)
	}
}

func main() {
	args := os.Args
	for _, arg := range args {
        switch arg {
		case "--debug" :
            flags.Debug = true
		case "--verbose" :
			flags.Verbose = true
        }
    }
	
// logfailo sukūrimas
	logFile, err := NewLogFile(fRunningFile)
	defer logFile.Close()
	log.SetOutput(logFile)

	Debug("\nargs: ", args, "\n\n")
	Debug("flags.Debug: ", flags.Debug, "\n\n")
	Debug("flags.Verbose: ", flags.Verbose, "\n\n")

// logfailo pildymo pradžia
	Debug(fRunningFile, " started\n")
	Out("\n")
	
// darbo pradžia

	var confKeeper ConfKeeper
	
	err = confKeeper.Run()
	if err != nil {
		Out(err.Error(), "\n\n")
	}
	Out("Press any key to exit...\n")
	keyboard.GetSingleKey()

}
 