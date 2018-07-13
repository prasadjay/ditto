package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
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

	InArguments := os.Args

	prog := InArguments[0]

	tmp := "/tmp/" + prog + "." + strconv.Itoa(os.Getpid())
	tmp2 := "/tmp/." + prog + "." + strconv.Itoa(os.Getpid()) + ".2"
	// osuf := "." + prog + ".orig"
	// nsuf := "." + prog + ".1"
	apt := "/usr/bin/apt-get"
	nawk := "/usr/bin/nawk"
	// apoctl := "/usr/bin/apoctl"
	// enforcerd := "/usr/sbin/enforcerd"
	// systemctl := "/bin/systemctl"
	// awk := "/usr/bin/nawk"
	swapoff := "/sbin/swapoff"
	curl := "/usr/bin/curl"
	id := "/usr/bin/id"
	// k := "/bin/ksh"
	// //ostype := `/bin/uname -a | $awk '{print$1" "substr($3,1,1)}'`
	// adminconf := "/etc/kubernetes/admin.conf"
	// config := usr.HomeDir + "/.kube/config"
	kubeadm := "/usr/bin/kubeadm"
	kubelet := "/usr/bin/kubelet"
	kubectl := "/usr/bin/kubectl"

	log := "/var/tmp/" + prog + ".log.date " + nowtime.Format("2006-Mar-02 15:04:05")
	out := "./Install_Readme_" + prog + ".log.date " + nowtime.Format("2006-Mar-02 15:04:05")
	//version := "1.00"

	_, err = exec.Command("/bin/ksh", "PATH=\"$HOME:/usr/bin:/bin:/usr/sbin:/sbin:/usr/ucb\";export PATH;! -x "+curl, "ksh").Output()
	if err != nil {
		fmt.Println(prog + ": SANITY: /usr/bin/curl missing!  Will install later ...")
	}

	_, err = exec.Command("/bin/ksh", "PATH=\"$HOME:/usr/bin:/bin:/usr/sbin:/sbin:/usr/ucb\";export PATH;! -x "+apt, "ksh").Output()
	if err != nil {
		os.Remove(tmp)
		os.Remove(tmp2)
		fmt.Println(prog + ": SANITY: /usr/bin/apt-get missing!  Quitting...")
		os.Exit(1)
	}

	_, err = exec.Command("/bin/ksh", "PATH=\"$HOME:/usr/bin:/bin:/usr/sbin:/sbin:/usr/ucb\";export PATH;! -x "+nawk, "ksh").Output()
	if err != nil {
		os.Remove(tmp)
		os.Remove(tmp2)
		fmt.Println(prog + ": SANITY: /usr/bin/nawk missing!  Quitting...")
		os.Exit(1)
	}

	if len(InArguments) > 1 {
		switch InArguments[1] {
		default:
			fmt.Println("usage : " + prog + " - no flags at this time.")
			break
		}
	} else {
		fmt.Println("usage : " + prog + " - no flags at this time.")
	}

	_, err = exec.Command("/bin/ksh", "PATH=\"$HOME:/usr/bin:/bin:/usr/sbin:/sbin:/usr/ucb\";export PATH;! -x "+id, "ksh").Output()
	if err != nil {
		os.Remove(tmp)
		os.Remove(tmp2)
		fmt.Println(prog + ": SANITY: /usr/bin/id missing!  Quitting...")
		os.Exit(1)
	}

	_, err = exec.Command("/bin/ksh", "PATH=\"$HOME:/usr/bin:/bin:/usr/sbin:/sbin:/usr/ucb\";export PATH;! -x "+swapoff, "ksh").Output()
	if err != nil {
		os.Remove(tmp)
		os.Remove(tmp2)
		fmt.Println(prog + ": SANITY: /sbin/swapoff missing!  Quitting...")
		os.Exit(1)
	}

	_, err = exec.Command("/bin/ksh", "PATH=\"$HOME:/usr/bin:/bin:/usr/sbin:/sbin:/usr/ucb\";export PATH;! -x /bin/df", "ksh").Output()
	if err != nil {
		os.Remove(tmp)
		os.Remove(tmp2)
		fmt.Println(prog + ": SANITY: /bin/df missing!  Quitting...")
		os.Exit(1)
	}

	//handle Ctrl + C

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

	//APT-Related Requirements
	apt_required_1 := [...]string{"ebtables", "ethtool"}
	//apt_required_2 := [...]string{"docker.io", "golang", "git", "apt-transport-https", "curl"}
	apt_required_2 := [...]string{"docker.io", "git", "apt-transport-https", "curl"}
	k8s_required := [...]string{"kubelet", "kubeadm", "kubectl"}

	//Install apt-required

	for x := 0; x < len(apt_required_1); x++ {
		value := apt_required_1[x]
		fmt.Println("Installing " + value + "...\\c")

		_, err = exec.Command("/bin/ksh", "PATH=\"$HOME:/usr/bin:/bin:/usr/sbin:/sbin:/usr/ucb\";export PATH;"+apt+" install -y "+value+" -q > /dev/null 2>&1", "ksh").Output()
		if err != nil {
			fmt.Println("Error : " + err.Error())
		} else {
			fmt.Println("Done")
		}
	}

	//Perform APT Update
	_, _ = exec.Command("/bin/ksh", "PATH=\"$HOME:/usr/bin:/bin:/usr/sbin:/sbin:/usr/ucb\";export PATH;"+apt+" update > /dev/null 2>&1", "ksh").Output()
	fmt.Println("Done")

	//Install apt-required 2
	for x := 0; x < len(apt_required_2); x++ {
		value := apt_required_2[x]
		fmt.Println("Installing " + value + "...\\c")

		_, err = exec.Command("/bin/ksh", "PATH=\"$HOME:/usr/bin:/bin:/usr/sbin:/sbin:/usr/ucb\";export PATH;"+apt+" install -y "+value+" -q > /dev/null 2>&1", "ksh").Output()
		if err != nil {
			fmt.Println("Error : " + err.Error())
		} else {
			fmt.Println("Done")
		}
	}

	//Restart Docker
	_, err = exec.Command("/bin/ksh", "PATH=\"$HOME:/usr/bin:/bin:/usr/sbin:/sbin:/usr/ucb\";export PATH;systemctl start docker >/dev/null 2>&1 ; sleep 2;$systemctl enable docker >/dev/null 2>&1 ; sleep 2", "ksh").Output()
	if err != nil {
		fmt.Println("Error : " + err.Error())
	} else {
		fmt.Println("Done")
	}

	//Adding apt-key
	fmt.Println("Adding apt-key...\\c")
	_, err = exec.Command("/bin/ksh", "PATH=\"$HOME:/usr/bin:/bin:/usr/sbin:/sbin:/usr/ucb\";export PATH;curl -s https://packages.cloud.google.com/apt/doc/apt-key.gpg | apt-key add - >/dev/null 2>&1", "ksh").Output()
	if err != nil {
		fmt.Println("Error : " + err.Error())
	} else {
		fmt.Println("Done")
	}

	//check/modify /etc/apt/sources.list.d/kubernetes.list
	//generate verifyKList.sh
	ioutil.WriteFile("verifyKList.sh", []byte("#!/bin/ksh\nPATH=\"$HOME:/usr/bin:/bin:/usr/sbin:/sbin:/usr/ucb\"\nexport PATH\nfile=\"/etc/apt/sources.list.d/kubernetes.list\"\nif [ ! -f \"$file\" ] ; then\n	print \"$1: $file not present.  Creating...\\c\"\n	touch $file\ncat << SEOF > $file\n# Added by Aporeto setup `date`\n\ndeb http://apt.kubernetes.io/ kubernetes-xenial main\n# Aporeto-end\nSEOF\nelse [ `head -10 $file | fgrep 'Added by Aporeto'|wc -w` -gt 0 ] \n\n	print \"$1: $file already added.  Skipping...\\c\"\n	#\n	#	sanity check\n	#\n	if [ $? != 0 ] ; then\n		print \"$1: FATAL: modification of $file failed!  Quitting.\"\n		rm -f $2 $3\n		exit 1\n	fi\n	cat \"$file\" >> $2\n	if [ $? != 0 ] ; then\n		print \"$1: FATAL: modification of $file failed!  Quitting.\"\n		rm -f $2 $3\n		exit 1\n	fi\n	cat $2 > \"$file\"\n	if [ $? != 0 ] ; then\n		print \"$1: FATAL: update of $file failed!  Quitting.\"\n		rm -f $2 $3\n		exit 1\n	fi\nfi"), 7777)
	//give permission
	_, _ = exec.Command("/bin/ksh", "PATH=\"$HOME:/usr/bin:/bin:/usr/sbin:/sbin:/usr/ucb\";export PATH;chmod 777 verifyKList.sh", "ksh").Output()
	//execute verifyKLists
	_, err = exec.Command("/bin/ksh", "PATH=\"$HOME:/usr/bin:/bin:/usr/sbin:/sbin:/usr/ucb\";export PATH;./verifyKList.sh", prog, tmp, tmp2, "ksh").Output()
	if err != nil {
		fmt.Println("Error : " + err.Error())
	} else {
		fmt.Println("Done")
	}
	//delete verifyList.sh
	os.Remove("verifyKList.sh")

	//Perform APT Update
	_, _ = exec.Command("/bin/ksh", "PATH=\"$HOME:/usr/bin:/bin:/usr/sbin:/sbin:/usr/ucb\";export PATH;"+apt+" update > /dev/null 2>&1", "ksh").Output()
	fmt.Println("Done")

	//Install K8S Required
	for x := 0; x < len(k8s_required); x++ {
		value := k8s_required[x]
		fmt.Println("Installing " + value + "...\\c")

		_, err = exec.Command("/bin/ksh", "PATH=\"$HOME:/usr/bin:/bin:/usr/sbin:/sbin:/usr/ucb\";export PATH;"+apt+" install -y "+value+" -q > /dev/null 2>&1", "ksh").Output()
		if err != nil {
			fmt.Println("Error : " + err.Error())
		} else {
			fmt.Println("Done")
		}
	}

	//Check for kubeadm, kubelet, kubectl

	_, err = exec.Command("/bin/ksh", "PATH=\"$HOME:/usr/bin:/bin:/usr/sbin:/sbin:/usr/ucb\";export PATH;! -x "+kubeadm, "ksh").Output()
	if err != nil {
		os.Remove(tmp)
		os.Remove(tmp2)
		fmt.Println(prog + ": SANITY: /usr/bin/kubeadm missing!  Quitting...")
		os.Exit(1)
	}

	_, err = exec.Command("/bin/ksh", "PATH=\"$HOME:/usr/bin:/bin:/usr/sbin:/sbin:/usr/ucb\";export PATH;! -x "+kubelet, "ksh").Output()
	if err != nil {
		os.Remove(tmp)
		os.Remove(tmp2)
		fmt.Println(prog + ": SANITY: /usr/bin/kubelet missing!  Quitting...")
		os.Exit(1)
	}

	_, err = exec.Command("/bin/ksh", "PATH=\"$HOME:/usr/bin:/bin:/usr/sbin:/sbin:/usr/ucb\";export PATH;! -x "+kubectl, "ksh").Output()
	if err != nil {
		os.Remove(tmp)
		os.Remove(tmp2)
		fmt.Println(prog + ": SANITY: /usr/bin/kubectl missing!  Quitting...")
		os.Exit(1)
	}

	//Clone cri-tools repo
	fmt.Println("Cloning git repo for cri-tools...\\c")
	_, err = exec.Command("/bin/ksh", "PATH=\"$HOME:/usr/bin:/bin:/usr/sbin:/sbin:/usr/ucb\";export PATH;git clone https://github.com/kubernetes-incubator/cri-tools.git >/dev/null 2>&1", "ksh").Output()
	if err != nil {
		fmt.Println("Error fetching from GIT : " + err.Error())
		os.Exit(1)
	} else {
		fmt.Println("Done")
	}

	//Begin readme

	//generate readme.sh
	ioutil.WriteFile("readme.sh", []byte("PATH=\"$HOME:/usr/bin:/bin:/usr/sbin:/sbin:/usr/ucb\"export PATH\n\ncat << REOF > $1\n#-----------------------------------------------------------------------------------------\n# Added by Aporeto setup `date`\n#-----------------------------------------------------------------------------------------\nCommands you may wish to try:\n\n 1) kubectl get pods --all-namespaces (check to see if pods are running)\n	2) kubectl -n guestbook get svc front-end (interact with frontend)\n	3) kubectl delete namespace guestbook (to remove the guestbook)\n\n#-----------------------------------------------------------------------------------------\nPlease review the following file just created for additional information.\n#-----------------------------------------------------------------------------------------\n\n\n\n REOF"), 7777)
	//give permission
	_, _ = exec.Command("/bin/ksh", "PATH=\"$HOME:/usr/bin:/bin:/usr/sbin:/sbin:/usr/ucb\";export PATH;chmod 777 readme.sh", "ksh").Output()
	//execute
	_, err = exec.Command("/bin/ksh", "PATH=\"$HOME:/usr/bin:/bin:/usr/sbin:/sbin:/usr/ucb\";export PATH;./readme.sh", out, "ksh").Output()
	if err != nil {
		fmt.Println("Error : " + err.Error())
	} else {
		fmt.Println("Done")
	}
	//delete readme.sh
	os.Remove("readme.sh")

}
