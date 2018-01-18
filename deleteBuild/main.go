/*
 _____     ______     __         ______     ______   ______    
/\  __-.  /\  ___\   /\ \       /\  ___\   /\__  _\ /\  ___\   
\ \ \/\ \ \ \  __\   \ \ \____  \ \  __\   \/_/\ \/ \ \  __\   
 \ \____-  \ \_____\  \ \_____\  \ \_____\    \ \_\  \ \_____\ 
  \/____/   \/_____/   \/_____/   \/_____/     \/_/   \/_____/ 
                                                               
*/
package main

import (
	"fmt"
	"os"
	// aliasing library names
	flag "github.com/ogier/pflag"
	"path/filepath"
	"github.com/fatih/color"
)

var debug bool
var rootPath string
var folderName string

func init() {
	flag.StringVarP(&rootPath, "path", "p", "./", "directory to start with")
	flag.StringVarP(&folderName, "name", "n", "build", "specific file name needs to be delete")
	flag.BoolVarP(&debug, "debug", "d", false, "Skip delele for debug")
}

func printUsage(){
	color.Green("Delete build files recursively")
	color.Green("Usage: %s [options]\n", os.Args[0])
	fmt.Println("Options:")
	flag.PrintDefaults()
	os.Exit(1)
}

func die(a interface{}, e error) {
	fmt.Println(a, e)
	os.Exit(1)
}

func main() {
	flag.Parse()
	if flag.NFlag() == 0 {
		printUsage()
	}
	if _, err := os.Stat(rootPath); err != nil {
		die("directory does not exist", err)
	}
	fileList := []string{}
	filepath.Walk(rootPath, func(path string, info os.FileInfo, err error)error {
		if info.Name() == folderName {
			color.Green("Found %s", path)
			fileList = append(fileList, path)
		}
		return err
	})
	for _, file := range fileList {
		color.Yellow("Delete %s", file)
		if debug {
			continue
		}
		if e := os.RemoveAll(file); e != nil {
			color.Red("Remove failed %v", e)
		}
	}
}