package main

import (
//	"bufio"
	"bytes"
	"datatree"
	"fmt"
//	"github.com/fatih/color"
//	"io"
//	"os"
//	"path/filepath"
//	"sort"
	"strconv"
	"strings"
)

const (
	BREAKPOINT        = "breakpoint"             // godebug breakpoint
	
	fStartConf	      = "start.conf"
	
	bDefaults         = "StartConf"
	bCurrent          = "AppConf"
	bConfList         = "AppConfsList"
	bConfData         = "AppConfsData"
	bPatterns         = "Patterns"
	bSelected         = "Selected"
	
	kApp              = "App"
	kDef              = "DefApp"
	kDb               = "Database"
	kSys              = "System"
	
	kName             = "Name"
	kConfSrc          = "ConfSrc"
	kConfDst          = "ConfDst"
	
	isDir             = true
	isFile            = false

	separator         = "\n--------------------------\n"
	bigseparator      = "\n----------------------------------------------------\n"
)

type ConfKeeper struct {
	datatree.DTree
}

func (confKeeper *ConfKeeper) Run() (err error) {
	var confManager ConfManager
	confManager.ConfKeeper = confKeeper
_ = BREAKPOINT
	err = confManager.Prepare()
	if err != nil { return }
	err = confManager.CheckData()
	if err != nil { return }
	err = confManager.LoadSystem()
	return
}

type ConfManager struct{
	ConfKeeper   *ConfKeeper
	datatree.JsonHandler
}

func (confManager *ConfManager) Prepare() (err error) {
	err = confManager.AddData(bDefaults, fStartConf)
	if err != nil {
		return
	}
	_, appConfName, err := confManager.ConfKeeper.Get(Key(bDefaults, kApp, kConfDst))
	if err != nil {
		return
	}
	confManager.AddData(bCurrent, appConfName.(string))
	return
}

func (confManager *ConfManager) ReadFile(fileName string) (err error) {
	err = confManager.JsonHandler.ReadFile(fileName)
	if err != nil { return }
	err = confManager.JsonHandler.Decode()
	return
}

func (confManager *ConfManager) AddData(path, fileName string) (err error) {
	path = strings.TrimSpace(path)
	if path == "" {
		err = fmt.Errorf("ConfManager.AddData().path is empty!")
		return
	} 
	err = confManager.ReadFile(fileName)
	if err != nil { return }
	confManager.ConfKeeper.Set(path, confManager.Value)
	return
}

func (confManager *ConfManager) CheckData() (err error) {
	return
}

func (confManager *ConfManager) LoadSystem() (err error) {
	return
}

func Key(args ...interface{}) (result string){
	var buffer bytes.Buffer

    for _, val := range args {
		buffer.WriteString(".")
		switch typedVal := val.(type) {
		case int:
			buffer.WriteString(strconv.Itoa(typedVal))
		case string:
			buffer.WriteString(typedVal)
		default:
			err := fmt.Errorf("Key has member with type other than int or string!")
			panic(err)
		}
		
    }

    result = string(buffer.Bytes()[1:])
	return
}
