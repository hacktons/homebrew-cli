/**
 * __________________________________________________________________________________________
 *     __     ____     _    _        __       __     _   _    ____     __     ____     _____
 *     / |    /    )   /  ,'       /    )   /    )   /  /|    /    )   / |    /    )   /    '
 * ---/__|---/____/---/_.'--------/--------/----/---/| /-|---/____/---/__|---/___ /---/__----
 *   /   |  /        /  \        /        /    /   / |/  |  /        /   |  /    |   /
 * _/____|_/________/____\______(____/___(____/___/__/___|_/________/____|_/_____|__/____ ___
 *
 *
 * Simple APK Analyzer. Can be used to compare size changes between apks. Compoents inside APK
 * are group by major types, such as：classesN.dex, assets/*, resources.arsc, res/*, lib/*,
 * META-INF/*, and other files include by third-part sdk.
 */
package main

import (
	"archive/zip"
	"bytes"
	"compress/gzip"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"text/template"

	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/fatih/color"
	flag "github.com/ogier/pflag"
)

// APKSizeInfo represents the major compoents whose size will be analyzed
type APKSizeInfo struct {
	Name          string `json:"name"`     // apk file name
	Sha1          string `json:"sha1"`     // hex string of sha1
	Dex           uint64 `json:"dex"`      // classesN.dex
	ResourcesArsc uint64 `jspn:"arsc"`     // resources.arsc
	Res           uint64 `json:"res"`      // res/*
	Assets        uint64 `json:"assets"`   // assets/*
	Lib           uint64 `json:"lib"`      // lib/*
	MetaINF       uint64 `json:"meta"`     // META-INF/*
	Others        uint64 `json:"others"`   // any other files
	Total         uint64 `json:"total"`    // raw size
	Download      uint64 `json:"download"` // download size compressed by gzip -9
}

func (*APKSizeInfo) formatSize(size uint64) float32 {
	return float32(size) / 1024 / 1024
}

func parseApk(apkFile string) *APKSizeInfo {
	r, e := zip.OpenReader(apkFile)
	_debug("file: %s", apkFile)
	if e != nil {
		_debug("Failed open apk: %v", e)
		return nil
	}
	apkSizeInfo := APKSizeInfo{
		Name:          filepath.Base(apkFile),
		Dex:           0,
		ResourcesArsc: 0,
		Res:           0,
		Assets:        0,
		Lib:           0,
		MetaINF:       0,
		Others:        0,
	}
	for _, f := range r.File {
		if f.Name == "resources.arsc" {
			apkSizeInfo.ResourcesArsc = f.FileHeader.CompressedSize64
		} else if strings.Contains(f.Name, "classes.") {
			apkSizeInfo.Dex += f.FileHeader.CompressedSize64
		} else if strings.Contains(f.Name, "res/") {
			apkSizeInfo.Res += f.FileHeader.CompressedSize64
		} else if strings.Contains(f.Name, "assets/") {
			apkSizeInfo.Assets += f.FileHeader.CompressedSize64
		} else if strings.Contains(f.Name, "lib/") {
			apkSizeInfo.Lib += f.FileHeader.CompressedSize64
		} else if strings.Contains(f.Name, "META-INF/") {
			apkSizeInfo.MetaINF += f.FileHeader.CompressedSize64
		} else {
			apkSizeInfo.Others += f.FileHeader.CompressedSize64
		}
	}
	compressedApk, _ := gzipFile(apkFile, apkFile+".gz")
	defer os.Remove(compressedApk.Name())
	apk, _ := os.Open(apkFile)
	s, _ := apk.Stat()
	apkSizeInfo.Total = uint64(s.Size())
	if e != nil {
		_debug("Failed to calculate download size: %v", e)
		apkSizeInfo.Download = uint64(s.Size())
	} else {
		size, _ := compressedApk.Stat()
		apkSizeInfo.Download = uint64(size.Size())
	}

	apkSizeInfo.Sha1 = sha14File(apkFile)
	_info("Apk Size, dex=%d, arsc=%d, res=%d, assets=%d, lib=%d, metainfo=%d, others=%d, total=%d, download=%d",
		apkSizeInfo.Dex, apkSizeInfo.ResourcesArsc, apkSizeInfo.Res, apkSizeInfo.Assets, apkSizeInfo.Lib, apkSizeInfo.MetaINF, apkSizeInfo.Others, apkSizeInfo.Total, apkSizeInfo.Download)
	return &apkSizeInfo
}

func gzipFile(name string, dst string) (*os.File, error) {
	os.Remove(dst)
	apk, _ := os.Open(name)
	temp, _ := os.Create(dst)
	zw, _ := gzip.NewWriterLevel(temp, gzip.BestCompression)
	defer zw.Flush()
	defer zw.Close()
	io.Copy(zw, apk)
	return temp, nil
}

func sha14File(name string) (hash string) {
	f, err := os.Open(name)
	if err != nil {
		_debug("invalid file: %v", err)
	}
	defer f.Close()

	h := sha1.New()
	if _, err := io.Copy(h, f); err != nil {
		_debug("calculate sha1 failed: %v", err)
	}
	hash = hex.EncodeToString(h.Sum(nil))

	_debug("%s", hash)
	return hash
}

var apkPath string
var output string
var humanReadableSize bool
var showLog bool
var format string

func init() {
	flag.StringVarP(&apkPath, "path", "p", "./", "Path or Directory of the target apk file")
	flag.StringVarP(&output, "output", "o", "output.xlsx", "Output excel name")
	flag.BoolVar(&humanReadableSize, "readable", true, "output the size in MB, instead of bytes count")
	flag.BoolVar(&showLog, "log", false, "Print all debug log")
	flag.StringVar(&format, "format", "xlsx", "foramt to output: json, xlsx(which is default)")
}

func main() {
	flag.Parse()
	if flag.NFlag() == 0 {
		printUsage()
		os.Exit(0)
	}
	// safty check
	if strings.HasPrefix(apkPath, "~") {
		_debug("calcute apk path: %s", apkPath)
		basePath, _ := user.Current()
		apkPath = filepath.Join(basePath.HomeDir, strings.Replace(apkPath, "~", "", 1))
	}
	// set to the same folder
	if !strings.HasPrefix(output, "~") && !strings.HasPrefix(output, "/") {
		_debug("calculate abs of output: %s", output)
		output = filepath.Join(filepath.Dir(apkPath), output)
	}
	if extension := filepath.Ext(output); extension == "" {
		_debug("append extension: %s", format)
		output += "." + format
	}
	_info("path: %s \toutput: %s", apkPath, output)

	var apks []APKSizeInfo
	stat, _ := os.Stat(apkPath)
	if stat.IsDir() {
		subFiles, e := ioutil.ReadDir(apkPath)
		if e != nil {
			_error("Read directory failed %v", e)
			os.Exit(1)
		}
		for _, f := range subFiles {
			apkAbs := filepath.Join(apkPath, f.Name())
			if strings.HasSuffix(f.Name(), ".apk") {
				apk := parseApk(apkAbs)
				if apk == nil {
					continue
				}
				apks = append(apks, *apk)
			}
		}
	} else if strings.HasSuffix(apkPath, ".apk") {
		apk := parseApk(apkPath)
		if apk != nil {
			apks = append(apks, *apk)
		}
	}

	var success bool
	if format == "xlsx" {
		success = exportXLSX(apks, output)
	} else if format == "json" {
		success = exportJSON(apks, output)
	}
	if success {
		_info("analyzing completed!!")
	} else {
		_error("analyzing failed!!")
	}
}

/**
 * export apk size information into excel with table and chart
 */
func exportXLSX(apks []APKSizeInfo, dst string) bool {
	categories := map[string]string{
		"A1": "Version",
		"B1": "SHA1",
		"C1": "download",
		"D1": "raw",
		"E1": "dex",
		"F1": "arsc",
		"G1": "assets",
		"H1": "res",
		"I1": "lib",
		"J1": "META-INF",
		"K1": "others",
	}
	values := map[string]interface{}{}
	// 58client_v8.3.1_58585858_20180404_13.54_release.apk => v8.3.1
	regx, _ := regexp.Compile("_([v.0-9]+)_")
	for i, apkSizeInfo := range apks {
		rows := strconv.Itoa(i + 2)
		index := regx.FindStringIndex(apkSizeInfo.Name)
		var version string
		if index != nil {
			version = apkSizeInfo.Name[index[0]+1 : index[1]-1]
			_debug("mathed version is %s", version)
		} else {
			version = apkSizeInfo.Name
		}
		values["A"+rows] = version
		values["B"+rows] = apkSizeInfo.Sha1
		values["C"+rows] = apkSizeInfo.formatSize(apkSizeInfo.Download)
		values["D"+rows] = apkSizeInfo.formatSize(apkSizeInfo.Total)
		values["E"+rows] = apkSizeInfo.formatSize(apkSizeInfo.Dex)
		values["F"+rows] = apkSizeInfo.formatSize(apkSizeInfo.ResourcesArsc)
		values["G"+rows] = apkSizeInfo.formatSize(apkSizeInfo.Assets)
		values["H"+rows] = apkSizeInfo.formatSize(apkSizeInfo.Res)
		values["I"+rows] = apkSizeInfo.formatSize(apkSizeInfo.Lib)
		values["J"+rows] = apkSizeInfo.formatSize(apkSizeInfo.MetaINF)
		values["K"+rows] = apkSizeInfo.formatSize(apkSizeInfo.Others)
	}
	xlsx := excelize.NewFile()
	for k, v := range categories {
		xlsx.SetCellValue("Sheet1", k, v)
	}
	for k, v := range values {
		xlsx.SetCellValue("Sheet1", k, v)
	}
	lineChartTemplate := `{
		"type": "line",
		"series": [
			{
				"name": "Sheet1!$C$1",
				"categories": "Sheet1!$A$2:$A${{.size}}",
				"values": "Sheet1!$C$2:$C${{.size}}"
			},
			{
				"name": "Sheet1!$D$1",
				"categories": "Sheet1!$A$2:$A${{.size}}",
				"values": "Sheet1!$D$2:$D${{.size}}"
			}
		],
		"title": {
			"name": "APK Size Line"
		}
	}`
	t, _ := template.New("temp").Parse(lineChartTemplate)
	buf := new(bytes.Buffer)
	t.Execute(buf, map[string]interface{}{"size": len(apks) + 1})
	chartJSON := buf.String()
	_debug("line chart:\n\n\t%s", chartJSON)
	xlsx.AddChart("Sheet1", "A"+strconv.Itoa(len(apks)+3), chartJSON)

	funcs := template.FuncMap{
		"customMethod": func(i int, l int) string {
			index := strconv.Itoa(i + 2)
			item := "{\"name\": \"Sheet1!$A$" + index + "\"," +
				"\"categories\": \"Sheet1!$C$1:$K$1\"," +
				"\"values\": \"Sheet1!$C$" + index + ":$K$" + index + "\"}"
			if i == l-1 {
				return item
			}
			return item + ","
		},
	}

	col3DChartTemplate := `{
		"type": "col3DClustered",
		"series": [
			{{range $key, $value := .apkInfoSize}}
				{{ customMethod $key $.size}}
			{{end}}
		],
		"title": {
			"name": "APK Version Details"
		}
	}`
	buf = new(bytes.Buffer)
	t, _ = template.New("3D").Funcs(funcs).Parse(col3DChartTemplate)
	t.Execute(buf, map[string]interface{}{"apkInfoSize": apks, "size": len(apks)})
	col3DChartJSON := buf.String()
	_debug("column chart\n\n\t%s", col3DChartJSON)
	xlsx.AddChart("Sheet1", "M1", col3DChartJSON)

	err := xlsx.SaveAs(dst)
	if err != nil {
		_error("Save xlsx failed, $v", err)
		return false
	}
	return true
}

/**
 * export apk size information into json format
 */
func exportJSON(apks []APKSizeInfo, dst string) bool {
	jsonBytes, e := json.Marshal(apks)
	if e != nil {
		return false
	}
	jsonText := string(jsonBytes)
	_debug(jsonText)
	if jsonFile, e := os.Create(dst); e == nil {
		jsonFile.WriteString(jsonText)
		jsonFile.Close()
	} else {
		return false
	}
	return true
}

func printUsage() {
	fmt.Printf("Usage: %s [options]", os.Args[0])
	fmt.Println("\nOptions:")
	flag.PrintDefaults()
	fmt.Println("\nExample:")
	color.Green("\tapkcompare [-p directory] [-o output.xlsx]")
}

func tag(format string) string {
	return "[ApkCompare] " + format
}

func _info(format string, params ...interface{}) {
	color.Green(tag(format), params...)
}

func _error(format string, params ...interface{}) {
	color.Red(tag(format), params...)
}

func _debug(format string, params ...interface{}) {
	if showLog {
		color.Yellow(tag(format), params...)
	}
}
