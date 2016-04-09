package main

import (
	"bufio"
	"bytes"
	. "github.com/re-pe/dtree-go"
	"fmt"
	"github.com/fatih/color"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

const (
//	BREAKPOINT        = "breakpoint"             // godebug breakpoint
	
	fStartConf	      = "start.conf"
	
	bDefaults         = "StartConf"
	bCurrent          = "AppConf"
	bConfList         = "AppConfsList"
	bConfData         = "AppConfsData"
	bPatterns         = "Patterns"
	bSelected         = "Selected"
	
	kApp              = "Application"
	kAppL             = "ApplicationList"
	kAppD             = "DefaultApplication"
	kDb               = "Database"
	kDbL              = "DatabaseList"
	kDbD              = "DefaultDatabase"
	kSys              = "System"
	
	kId               = "ID"
	kName             = "Name"
	kConfSrc          = "ConfSrc"
	kConfDst          = "ConfDst"
	kConfCopier       = "ConfCopier"
	
	isDir             = true
	isFile            = false

	separator         = "\n--------------------------\n"
	bigseparator      = "----------------------------------------------------"
)


type ConfKeeper struct {
	DTree
}

func (confKeeper *ConfKeeper) Run() (err error) {
	var confManager ConfManager
	confManager.ConfKeeper = confKeeper

	err = confManager.Prepare()
	if err != nil { return }
	err = confManager.SetConfsData()
	if err != nil { return }

	err = confManager.SelectConf()
	if err != nil { return }

	selected := confKeeper.Get(bSelected)
	err = selected.Error
	if err != nil { return }
	if selected.Value.(string) == "-1" {
		Out("Bye bye.\n\n")
		return
	}
_ = BREAKPOINT
	err = confManager.CopyFiles()
	return
}

type ConfManager struct{
	ConfKeeper   *ConfKeeper
	JsonHandler
}

func (confManager *ConfManager) Prepare() (err error) {
	if pathExists(fStartConf, isFile).Index < 2 { 
		err = fmt.Errorf("ConfManager.Prepare(): file %s does not exist!", fStartConf)
		return 
	}
	err = confManager.AddData(bDefaults, fStartConf)
	if err != nil { 
		err = fmt.Errorf("ConfManager.Prepare(): file %s exists but have errors:\n\n  %s!", fStartConf, err.Error())
		return 
	}
	retValue := confManager.ConfKeeper.Get(Key(bDefaults, kAppL, kConfDst))
	appConfName, err := retValue.Value.(string), retValue.Error 
	if err != nil { return }
	appConfExists := pathExists(appConfName, isFile)
	if appConfExists.Index < 2 {
		Debug("?:ConfManager.Prepare(): file %s does not exist!\n\n", appConfName)
	} else {
		err = confManager.AddData(bCurrent, appConfName)
	}

	if appConfExists.Index > 1 && err != nil {
		err = fmt.Errorf("ConfManager.Prepare(): file %s exists but have errors:\n\n  %s!", appConfName, err.Error())
	}
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

func (confManager *ConfManager) SetConfsList() (err error) {
	var scanner DirScanner
	scanner.ConfKeeper = confManager.ConfKeeper
	err = scanner.AddPattern(
		bDefaults,
		Key(kAppD, kConfSrc), 
		Key(kAppL, kConfSrc),
	)
	if err != nil { return }
	err = scanner.ScanTo(bConfList)
	return
}

func (confManager *ConfManager) SetConfsData() (err error){
	err = confManager.SetConfsList()
	retValue := confManager.ConfKeeper.Get(bConfList)
	var confList []interface{}
	confList, err = retValue.Value.([]interface{}), retValue.Error

 	for _, fileName := range confList {
		err = confManager.ReadFile(fileName.(string))
		if err != nil { return }
		err = confManager.Decode()
		if err != nil { return }
		err = confManager.ConfKeeper.Set(Key(bConfData, "+"), confManager.Value).Error
		if err != nil { return }
	}
	confManager.Content = []byte(``)
	return
}

func (confManager *ConfManager) SelectConf() (err error){
	var selector Selector
	selector.ConfKeeper = confManager.ConfKeeper
	err = selector.Select()
	return
}

func (confManager *ConfManager) CopyFiles() (err error){
	
	confOperator := ConfOperator{
		ConfManager : confManager,
		FileList    : []CopyData{},
	}

	err = confOperator.PrepareToCopy()
	if err != nil { return }

	err = confOperator.CopyFileList()
	return
}

type DirScanner struct{
	ConfKeeper *ConfKeeper
}

func (scanner *DirScanner) AddPattern(path string, patterns ...string) (err error) {
	path = strings.TrimSpace(path)
	if path == "" {
		err = fmt.Errorf("DirScanner.AddPattern().path is empty!")
		return
	} 
	var value interface{}
	for _, key := range patterns {
		retValue := scanner.ConfKeeper.Get(Key(path, key))
		value, err = retValue.Value, retValue.Error
		if err != nil { return }
		err = scanner.ConfKeeper.Set(Key(bPatterns,"+"), value).Error
		if err != nil { return }
	}
	return
}

func (scanner *DirScanner) ScanTo(path string) (err error){
	path = strings.TrimSpace(path)
	if path == "" {
		err = fmt.Errorf("DirScanner.ScanTo().path is empty!")
		return
	} 
	var appFiles, allAppFiles []string
	var retValue DTree
	var patterns []interface{}
	if retValue = scanner.ConfKeeper.Get(bPatterns); retValue.Error != nil {
		err = retValue.Error
		return
	}
	patterns = retValue.Value.([]interface{})
	for _, value := range patterns {
	
		appFiles, err = filepath.Glob(value.(string))
		sort.Sort(ByFileName(appFiles))
		if err != nil { return }
		allAppFiles = append(allAppFiles, appFiles...)
	}
	key := Key(path, "+")
	for i, fileName := range allAppFiles {
		allAppFiles[i] = filepath.ToSlash(fileName)
		err = scanner.ConfKeeper.Set(key,allAppFiles[i]).Error
		if err != nil { return }
	}
	return
}

type ByFileName []string

func (a ByFileName) Len()           int { return len(a) }
func (a ByFileName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByFileName) Less(i, j int) bool { 
	return strings.ToLower(strings.Replace(a[i], filepath.Ext(a[i]), "", -1)) < 
			strings.ToLower(strings.Replace(a[j], filepath.Ext(a[j]), "", -1))
}

type Selector struct{
	ConfKeeper *ConfKeeper
	Selected string
}

func (selector *Selector) Select() (err error) {
	var choice int
	var input string
	retValue := selector.ConfKeeper.Get(bConfList)
	err = retValue.Error
	if err != nil { return }
	lenConfList := len(retValue.Value.([]interface{}))
	err = fmt.Errorf("")
	stdin := bufio.NewReader(os.Stdin)
	for err != nil {
		err = selector.OutputChoices()
		if err != nil { return }
		_, err = fmt.Scanf("%s", &input)
		stdin.ReadString('\n')

		if err != nil {
			Out("Error:", err.Error(), "\n\n")
			Out("\n", bigseparator, "\n\n")
			continue
		}
		input = strings.TrimSpace(input)
		var outputColor color.Attribute
		switch input {
		case "q", "--":
			choice = -1
			err = nil
		default :
			choice, err = strconv.Atoi(input)
			if err != nil {
				outputColor = color.FgRed
				err = fmt.Errorf("Your input is not integer number or code to exit!")
			}
		}
		if err == nil {
			switch {
			case choice < -1 || choice >= lenConfList:
				outputColor = color.FgMagenta
				err = fmt.Errorf("%d is out of range [ -1, %d ]!", choice, lenConfList - 1)
			default:
				outputColor = color.FgCyan
				err = nil
			}
		}
		Out("\nYou selected ")
		color.Set(outputColor, color.Bold)
		Out(input)
		color.Unset()
		Out(".\n\n")
		if err != nil {
			Out(err.Error(), "\n")
			Out("\n", bigseparator, "\n\n")
		} 
	}
	err = selector.ConfKeeper.Set(bSelected, strconv.Itoa(choice)).Error
	return
}

func (selector *Selector) OutputChoices() (err error) {
	var confList []interface{}
	retValue := selector.ConfKeeper.Get(bConfList)
	confList, err = retValue.Value.([]interface{}), retValue.Error
	if err != nil { return }

	Out("Choose your application:\n\n")
	var (curAppName, curConfSrc string ; appName interface{})

	retValue = selector.ConfKeeper.Get(Key(bCurrent, kApp))
	err = retValue.Error
	if err == nil { 
		curAppName = retValue.Value.(map[string]interface{})[kName].(string)
		curConfSrc = retValue.Value.(map[string]interface{})[kConfSrc].(string)
	}
	color.Set(color.FgGreen, color.Bold)
	for key, value := range confList {
		retValue = selector.ConfKeeper.Get(Key(bConfData, key, kApp, kName))
		appName, err = retValue.Value, retValue.Error
		if err != nil { 
			color.Unset()
			return 
		}
		if curAppName == appName && curConfSrc == value {
			color.Set(color.FgYellow, color.Bold)
			Out("?:Current: %2d. %-30v == %v\n", key, appName, value)
			color.Set(color.FgGreen, color.Bold)
		} else {
			Out("?:         %2d. %-30v == %v\n", key, appName, value)
		} 
	}
	//Out(bigseparator, "\n\n")
	color.Unset()
	Out("\nWrite number of selected application from list printed above\n\nor q, -- or -1 to exit: ")
	return
}

type CopyData struct{
	Key string
	Dst string
	Src string
}

type ConfOperator struct{
	ConfManager *ConfManager
	Selected string
	//ConfStruct ConfStruct
	FileList []CopyData
}

func (confOp *ConfOperator) PrepareToCopy() (err error){
	
	//duomenų medyje randamas pasirinktos aplikacijos eilės numeris
	//in the data tree, it finds the selected application's number
	selected := confOp.ConfManager.ConfKeeper.Get(bSelected)
	err = selected.Error
	if err != nil { return } 

	confOp.Selected = selected.Value.(string)
	
	//rastas numeris įkeliamas objekto laukan .Selected
	//the found number becomes loaded to the object's field .Selected 
	
	
	var selConfKeeper SelectedConfKeeper
	
	//duomenų medyje pagal eilės numerį randami pasirinktos aplikacijos konfigūracinio failo duomenys
	//rastieji duomenys „apvelkami“ DTree tipo objektu
	//in the data tree, by the number of application, it finds data of configuration file of application.   
	//the found data becomes "wrapped" by object of type derived form DTree type
	confData := confOp.ConfManager.ConfKeeper.Get(Key(bConfData, confOp.Selected))
	err = confData.Error
	if err != nil { return }

	selConfKeeper.Selected = confOp.Selected
	selConfKeeper.Value = confData.Value
	selConfKeeper.ConfKeeper = confOp.ConfManager.ConfKeeper
	
	err = selConfKeeper.UpdateAppConf()
	if err != nil { return }
	err = selConfKeeper.SetFileList()
	if err != nil { return }
	
	confOp.FileList = selConfKeeper.FileList
	//confOp.ConfStruct = selConfKeeper.ConfStruct

	confOp.ConfManager.JsonHandler.Value = selConfKeeper.Value
	return
}

func (confOp *ConfOperator) CopyFileList() (err error){
	var dbName interface{}
	for _, file := range confOp.FileList {
		
		err = ClearPath(file.Dst)
		if err != nil {
			return
		}

		if file.Key == kApp && file.Src == "" {
			err = confOp.ConfManager.JsonHandler.Encode()
			err = confOp.ConfManager.JsonHandler.NewFile(file.Dst)
			if err == nil {
				fmt.Printf("New file %s was created.\n\n", file.Dst)
			}
		} else {
			if file.Key == kSys {
				if dbName == nil {
					fmt.Printf("Application will not use any database.\n\n")
				} else {
					fmt.Printf("Application will load %s database.\n\n", dbName)
				}
			}
			_, err = CopyFile(file.Dst, file.Src)
			if err != nil { return }
			
			fmt.Printf("File %s was copied to %s.\n\n", file.Src, file.Dst)
			if file.Key == kDb {
				dbName = confOp.ConfManager.JsonHandler.Get(Key(kDb, kName)).Value
			}
		}
	}
	return
}

type SelectedConfKeeper struct{
	DTree
	ConfKeeper *ConfKeeper
	Selected string
	AppDir string
	ConfStruct ConfStruct
	FileList []CopyData
}

func (selConfKeeper *SelectedConfKeeper) UpdateAppConf() (err error) {

	//konfigūraciniuose duomenyse randamas konfigūracinio failo šaltinio kelias
	//in the data of configuration file, it gets found the source path of configuration file
	
	key := Key(bConfList, selConfKeeper.Selected)
	fileSrc := selConfKeeper.ConfKeeper.Get(bConfList)
	fileSrc = fileSrc.Get(selConfKeeper.Selected)
	err = fileSrc.Error
	if err != nil {
		err = fmt.Errorf("Value with key %s does not exist!", key)
		return
	}

	
	appDir := fileSrc.Value.(string)
	if appDir == "" {
		err = fmt.Errorf("Value with key %s is empty!", key)
		return
	}

	//iš konfigūracinio failo šaltinio kelio sužinomas aplikacijos katalogas
	//from path of the configuration file, it gets cognized directory of application
	selConfKeeper.AppDir = filepath.ToSlash(filepath.Dir(appDir))
	if selConfKeeper.AppDir == "" {
		err = fmt.Errorf("selConfKeeper.AppDir is empty!")
		return
	}
	err = selConfKeeper.CreateConfStruct()
	if err != nil { return }

	err = selConfKeeper.UpdateBrachesData()
	//if err != nil { return }
	
	return
}

func (selConfKeeper *SelectedConfKeeper) CreateConfStruct() (err error) {
	selConfKeeper.ConfStruct = ConfStruct{
		Branch{kApp, "Error", "Selected configuration file have no value with key %s!", Leaves{
			Leaf{kName,    "Error", "Selected configuration file have no value with key %s!" },
			Leaf{kConfSrc, "Redirect", Key(bConfList, selConfKeeper.Selected )               },
			Leaf{kConfDst, "Redirect", Key(bDefaults, kAppL, kConfDst        )               },
			
		}},
 		Branch{kDb, "Ignore", bDefaults, Leaves{
			Leaf{kId,      "Redirect", Key(kDbD, kId           )},
			Leaf{kName,    "Redirect", Key(kDbL, "%s", kName   )},
			Leaf{kConfSrc, "Redirect", Key(kDbD, kConfSrc      )},
			Leaf{kConfDst, "Redirect", Key(kDbL, "%s", kConfDst)},
		}},
		Branch{kSys, "Redirect", Key(bDefaults, kSys), Leaves{
			Leaf{kName,    "Redirect", kName                    },
			Leaf{kConfSrc, "Redirect", kConfSrc                 },
			Leaf{kConfDst, "Redirect", kConfDst                 },
		}},
	}

	if len(selConfKeeper.ConfStruct) < 1 {
		err = fmt.Errorf("selConfKeeper.ConfStruct was not created!")
	}
	return
}

func (selConfKeeper *SelectedConfKeeper) UpdateBrachesData() (err error) {
	//selectedConf := selConfKeeper.Value.(map[string]interface{})
	//Out("selConfKeeper.ConfKeeper.Get(bDefaults).Value:\n\n", selConfKeeper.ConfKeeper.Get(bDefaults).Value, "\n\n")
	//Out("selectedConf:\n\n", selectedConf, "\n\n")

	//Out("selConfKeeper.ConfKeeper.Get(bConfList).Value:\n\n", selConfKeeper.ConfKeeper.Get(bConfList).Value, "\n\n")
	//einama per konfigūracinio failo šakas
	//it is walking over branches of the conguration file

	var branch Branch
branches: for _, branch = range selConfKeeper.ConfStruct{

		err = selConfKeeper.Get(branch.Key).Error
		if err != nil {
			switch branch.Action {
			case "Error" :
				err = fmt.Errorf(branch.Data, branch.Key)
				return
			case "Ignore" :
				continue branches
			case "Redirect" :
			default: 
				err = fmt.Errorf("Unknown action in the branch %s!", branch.Key)
				return
			}
		}
		var leafId string
		var leaf Leaf
leaves: for _, leaf = range branch.Leaves {
			err = selConfKeeper.Get(Key(branch.Key, leaf.Key)).Error
			if err == nil { 
				continue leaves
			}
			switch leaf.Action {
			case "Error" :
				err = fmt.Errorf(leaf.Data, leaf.Key)
				return
			case "Ignore" :
				continue leaves
			case "Redirect" :
			default: 
				err = fmt.Errorf("Unknown action in the leaf %s!", leaf.Key)
				return
			}
			var branchData, leafData string
			if branch.Action != "Error" {
				branchData = strings.TrimSpace(branch.Data)
			}
			if leaf.Action == "Redirect" {
				leafData = strings.TrimSpace(leaf.Data)
			}
			if strings.Count(leafData, "%s") == 1 {
				leafData = fmt.Sprintf(leafData, leafId)
			}
			key := Key(branchData, leafData)
			missingData := selConfKeeper.ConfKeeper.Get(key)
			err = missingData.Error
			if err != nil {
				err = fmt.Errorf("Value with key %s have not found!", key)
				return
			}
			if leaf.Key == kId {
				leafId = missingData.Value.(string)
			}
			key = Key(branch.Key, leaf.Key) 
			err = selConfKeeper.Set(key, missingData.Value).Error
			if err != nil { 
				err = fmt.Errorf("Selected conf key %s was not updated with value %s!", key, missingData.Value)
				return 
			}
		}
	}
	//Out("selectedConf:\n\n", selectedConf, "\n\n")
 	return
}

func (selConfKeeper *SelectedConfKeeper) SetFileList() (err error) {
	//einama per konfigūracinio failo šakas
	//it is walking over branches of the conguration file
	for _, branch := range selConfKeeper.ConfStruct {
_ = BREAKPOINT
		if branch.Key == kDb {
			err = selConfKeeper.Get(branch.Key).Error
			if err != nil { continue }
		}
		fileDst := selConfKeeper.Get(Key(branch.Key, kConfDst))
		if err = fileDst.Error; err != nil { return }
		if branch.Key == kApp {
			selConfKeeper.FileList = append(
										selConfKeeper.FileList, 
										CopyData{branch.Key, fileDst.Value.(string), ""},
									)
		} else {
			fileSrc := selConfKeeper.Get(Key(branch.Key, kConfSrc))
			if err = fileSrc.Error; err != nil { return }
			
			filePath := selConfKeeper.AppDir + "/" + fileSrc.Value.(string)
			
			if pathExists(filePath, isFile).Index < 2 {
				err = fmt.Errorf("File %s does not exist!", filePath)
				return
			}
			selConfKeeper.FileList = append(
										selConfKeeper.FileList, 
										CopyData{branch.Key, fileDst.Value.(string), filePath},
									)
		}
	}
	return
}

type Leaf struct{
	Key string
	Action string
	Data string
}

type Leaves []Leaf

type Branch struct{
	Key string
	Action string
	Data string
	Leaves
}

type ConfStruct []Branch

func Key(args ...interface{}) (result string){
	var buffer bytes.Buffer

    for _, val := range args {
		switch typedVal := val.(type) {
		case int:
			buffer.WriteString(".")
			buffer.WriteString(strconv.Itoa(typedVal))
		case string:
			typedVal = strings.TrimSpace(typedVal)
			if typedVal != "" {
				buffer.WriteString(".")
				buffer.WriteString(typedVal)
			}
		default:
			err := fmt.Errorf("Key has member with type other than int or string!")
			panic(err)
		}
		
    }

    result = string(buffer.Bytes()[1:])
	return
}

type intResult struct {
	Index int
	Error error
}

func pathExists(path string, dir bool) (result intResult) {
	result.Index = -1
	path = strings.TrimSpace(path)
	if path == "" {
		result.Error = fmt.Errorf("PathExists.args.path is empty!")
		return
	}

	dirType := "Directory"
	if !dir {
		dirType = "File" 
	}

	finfo, err := os.Stat(path)
	if err != nil {
		// no such file or dir
		result.Index, result.Error = 0, fmt.Errorf("%s does not exist!", path)
		return
	}

	if dir == finfo.IsDir() { 
		result.Index = 2
	} else {
		result.Index = 1
	}

	if result.Index < 2 {
		result.Error = fmt.Errorf("%s is not %s!", path, strings.ToLower(dirType))
	}
	return
}

func ClearPath(fileName string) (err error) {
	fileDir := filepath.ToSlash(filepath.Dir(fileName))
	if pathExists(fileDir, isDir).Index < 2 {
		err = fmt.Errorf("Destination directory %s cannot be found!", fileDir)
		return
	}
	if pathExists(fileName, isFile).Index > 1 {
		err = os.Remove(fileName)
	}
	return
}

func CopyFile(destName, srcName string) (written int64, err error) {
	srcFile, err := os.Open(srcName) 
	if err != nil {
		err = fmt.Errorf("Source file %s cannot be open!", srcName)
		return 
	} 
	defer srcFile.Close()
	destFile, err := os.Create(destName) 
	if err != nil {
		err = fmt.Errorf("Destination file %s cannot be open!", destName)
		return
	} 
	defer destFile.Close()
	written, err = io.Copy(destFile, srcFile)
	return
}

