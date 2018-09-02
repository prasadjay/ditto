package main

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

func main() {
	fmt.Println(GetDNSForIp("13.127.249.213"))
}

func GetDNSForIp(ip string) (dns string) {
	var b []byte
	var err error

	if runtime.GOOS == "windows" {
		b, err = exec.Command("cmd", "/C", "nslookup "+ip+" | grep Name").Output()
	} else {
		cmd := exec.Command("sh", "-c", "nslookup "+ip+" | grep name")
		b, err = cmd.Output()
	}

	if err != nil {
		fmt.Println(err.Error())
		dns = "N/A"
	} else {
		if runtime.GOOS == "windows" {
			temp := strings.TrimSpace(string(b))
			temp = strings.Replace(temp, "Name", "", -1)
			temp = strings.Replace(temp, ":", "", -1)
			dns = strings.TrimSpace(temp)
		} else {
			temp := strings.TrimSpace(string(b))
			temp = strings.TrimSuffix(temp, ".")
			tempArray := strings.Split(temp, "name")
			dns = strings.Replace(tempArray[1], "=", "", -1)
			dns = strings.TrimSpace(dns)
		}
	}
	return
}
