package main

import (
	"fmt"
	"os"
	"os/user"
	"runtime"
	"strconv"
	"time"
)

func main() {
	if runtime.GOOS != "linux" {
		fmt.Println("Not supported version of operating system. Found : " + runtime.GOOS + ". Required: Linux Ubuntu")
		//return
	}

	hostname, _ := os.Hostname()
	nowtime := time.Now().UTC()
	continueCommand := ""

	usr, err := user.Current()
	if err != nil {
		fmt.Println("Error getting user for this session.")
		return
	} else {
		fmt.Println(usr.HomeDir)
	}

	prog := os.Args[0]

	tmp := "/tmp/" + prog + "." + strconv.Itoa(os.Getpid())
	_ = tmp
	/*
		tmp2 := "/tmp/." + prog + "." + strconv.Itoa(os.Getpid()) + ".2"
		osuf := "." + prog + ".orig"
		nsuf := "." + prog + ".1"
		apt := "/usr/bin/apt-get"
		nawk := "/usr/bin/nawk"
		apoctl := "/usr/bin/apoctl"
		enforcerd := "/usr/sbin/enforcerd"
		systemctl := "/bin/systemctl"
		awk := "/usr/bin/nawk"
		swapoff := "/sbin/swapoff"
		curl := "/usr/bin/curl"
		id := "/usr/bin/id"
		k := "/bin/ksh"
		//ostype := `/bin/uname -a | $awk '{print$1" "substr($3,1,1)}'`
		adminconf := "/etc/kubernetes/admin.conf"
		config := usr.HomeDir + "/.kube/config"
		kubeadm := "/usr/bin/kubeadm"
		kubelet := "/usr/bin/kubelet"
		kubectl := "/usr/bin/kubectl"
	*/
	log := "/var/tmp/" + prog + ".log.date " + nowtime.Format("2006-Mar-02 15:04:05")
	out := "./Install_Readme_" + prog + ".log.date " + nowtime.Format("2006-Mar-02 15:04:05")
	//version := "1.00"

	fmt.Println("")
	fmt.Println("")
	fmt.Println("")
	fmt.Println("-----------------------------------------------------------------")
	fmt.Println("Aporeto K8S Quick Start " + prog)
	fmt.Println("Installation of Kubernetes on Ubuntu")
	fmt.Println("-----------------------------------------------------------------")
	fmt.Println("running on                 " + hostname)
	fmt.Println("date:                      " + nowtime.Format(time.RFC1123))
	fmt.Println("log file:                  " + log)
	fmt.Println("Install Readme:            " + out)
	fmt.Println("")
	fmt.Println("Press ENTER to continue, else CTRL-C to quit ")
	fmt.Println("")
	fmt.Println("")
	fmt.Println("")
	_, _ = fmt.Scanln(&continueCommand)

	if continueCommand != "" {
		fmt.Println("Aborted by user...")
		return
	}

}
