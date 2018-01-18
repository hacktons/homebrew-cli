 /*
 __     __     __     ______   __        ______     _____     ______    
/\ \  _ \ \   /\ \   /\  ___\ /\ \      /\  __ \   /\  __-.  /\  == \   
\ \ \/ ".\ \  \ \ \  \ \  __\ \ \ \     \ \  __ \  \ \ \/\ \ \ \  __<   
 \ \__/".~\_\  \ \_\  \ \_\    \ \_\     \ \_\ \_\  \ \____-  \ \_____\ 
  \/_/   \/_/   \/_/   \/_/     \/_/      \/_/\/_/   \/____/   \/_____/ 
                                                                        
  */                                                                                                                                                                                                         
package main
import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"strconv"
)

func execCommand(cmd string, args []string) (output string, e error){
	cmdOut, err := exec.Command(cmd, args...).Output()
	return string(cmdOut), err
}

func log(a interface{}) {
	fmt.Println("[WiFi ADB] ", a)
}

func err(a interface{}, err error) {
	if err!=nil {
		fmt.Println("[WiFi ADB] ", a, err)
	} else {
		fmt.Println("[WiFi ADB] ", a)
	}
}

func die(a interface{}, e error) {
	err(a, e)
	os.Exit(1)
}

/**
* check if there are any devices connected via `adb devcies`
* prompt user to chose one if multi-device connected;
*/
func selectDevice()string {
	var (
		checkOutput string
		checkErr error
	)
	if checkOutput, checkErr = execCommand("adb", []string{"devices"});  checkErr!=nil{
		die("devices check failed", checkErr)
	}
	devices:=strings.Split(strings.Trim(checkOutput, "\n"), "\n")
	devicesCount:=cap(devices)-1
	
	log("device count "+strconv.Itoa(devicesCount))
	var slectedDevice string
	if (devicesCount < 1) {
		die("no device found", nil)
	} else if (devicesCount > 1 ) {
		log("find more than one device")
		for i:=1; i <= devicesCount; i++ {
			log(strconv.Itoa(i)+":\t" + devices[i])
		}
		fmt.Println("please input device index as list:")
		var inputDevice string
		fmt.Scanln(&inputDevice)
		var (
			index int
			e error
		)
		if index, e =strconv.Atoi(inputDevice); e!=nil {
			die("inlaid input, please try again!!", e)
		}
		if index <= 0 || index > devicesCount {
			die("inlaid input, please try again!!", nil)
		} else {
			slectedDevice=devices[index]
		}
	} else {
		slectedDevice=devices[1]	
	}
	spliteDevice:=strings.Fields(slectedDevice)
	deviceId:=spliteDevice[0]
	deviceType:=spliteDevice[1]
	if deviceType != "device" {
		die(slectedDevice+" is not valid USB device", nil)
	}
	if strings.Contains(deviceId, "emulator") {
		die(deviceId+" seems to be an emulator", nil)
	}
	if len(strings.Split(deviceId, ".")) == 4 {
		die(deviceId + " seems to be connected via tcpip already", nil)
	}
	return deviceId
}

func main () {
	device:=selectDevice()
	log(device)
	var (
		cmdOut string
		err    error
	)
	cmdName := "adb"
	cmdArgs := []string{"-s", device, "shell", " ip -f inet addr show wlan0"}
	/*
	40: wlan0: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc mq state UP qlen 1000
		inet 10.252.224.182/20 brd 10.252.239.255 scope global wlan0
	*/
	if cmdOut, err = execCommand(cmdName, cmdArgs); err != nil {
		// fmt.Fprintln(os.Stderr, "Execute adb shell failed: ", err)
		// os.Exit(1)
		die("execute adb shell failed:", err)
	}
	log(cmdOut)
	splited := strings.Split(cmdOut, "\n")
	addres := strings.Split(strings.Trim(splited[1], " "), " ");
	/*
	10.252.224.182
	*/
	ip := strings.Split(addres[1], "/")[0]
	log("device ip  "+ip)
	if _, e := execCommand("adb", []string{"-s", device, "tcpip", "5555"}); e!=nil {
		die("adb tcpip failed: ", err)
	}
	log("unplugin USB then press ENTER:")
	//adb connect 10.0.101.192
    var input string
	fmt.Scanln(&input)
	if _, e := execCommand("adb", []string{"connect", ip}); e!=nil {
		die("adb connect failed: ", err)
	}
	log("device connected: "+ip+":55555")
}