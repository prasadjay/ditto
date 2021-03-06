package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"runtime"
	"strconv"
	"strings"
	"time"
)

func main() {
	//check for OS
	if runtime.GOOS != "linux" {
		fmt.Println("Not supported version of operating system. Found : " + runtime.GOOS + ". Required: Linux Ubuntu")
		os.Exit(1)
	}

	hostname, _ := os.Hostname()
	nowtime := time.Now().UTC()
	continueCommand := ""
	homeDir := ""
	_ = homeDir
	usr, err := user.Current()
	if err != nil {
		fmt.Println("Error getting user for this session.")
		return
	} else {
		homeDir = usr.HomeDir
	}

	InArguments := os.Args

	tokens := strings.Split(InArguments[0], "/")
	prog := tokens[len(tokens)-1]

	//check for user
	cmdUser := exec.Command("id", "-u")
	usrOut, _ := cmdUser.Output()

	i, _ := strconv.Atoi(string(usrOut[:len(usrOut)-1]))

	if i != 0 {
		fmt.Println(prog + " : FATAL: must be run as root")
		os.Exit(1)
	}

	config := GetConfig()

	tmp := "/tmp/" + prog + "." + strconv.Itoa(os.Getpid())
	tmp2 := "/tmp/." + prog + "." + strconv.Itoa(os.Getpid()) + ".2"
	// osuf := "." + prog + ".orig"
	// nsuf := "." + prog + ".1"
	apt := "/usr/bin/apt-get"
	nawk := "/usr/bin/nawk"
	apoctl := "/usr/bin/apoctl"
	enforcerd := "/usr/sbin/enforcerd"
	systemctl := "/bin/systemctl"
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

	log := "/var/tmp/" + prog + ".log.date " + nowtime.Format(time.RFC3339)
	out := "./Install_Readme_" + prog + ".log." + nowtime.Format(time.RFC3339)
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
	apt_required_2 := [...]string{"docker.io", "golang", "git", "apt-transport-https", "curl"}
	//apt_required_2 := [...]string{"docker.io", "git", "apt-transport-https", "curl"}
	k8s_required := [...]string{"kubelet", "kubeadm", "kubectl"}

	//Install apt-required

	for x := 0; x < len(apt_required_1); x++ {
		value := apt_required_1[x]
		fmt.Print("Installing " + value + "...")

		_, err = exec.Command("/bin/ksh", "PATH=\"$HOME:/usr/bin:/bin:/usr/sbin:/sbin:/usr/ucb\";export PATH;"+apt+" install -y "+value+" -q > /dev/null 2>&1", "ksh").Output()
		if err != nil {
			fmt.Println("Error : " + err.Error())
			os.Exit(1)
		} else {
			fmt.Println("Done")
		}
	}

	//Perform APT Update
	_, _ = exec.Command("/bin/ksh", "PATH=\"$HOME:/usr/bin:/bin:/usr/sbin:/sbin:/usr/ucb\";export PATH;"+apt+" update > /dev/null 2>&1").Output()
	fmt.Println("Done updating apt")

	//Install apt-required 2
	for x := 0; x < len(apt_required_2); x++ {
		value := apt_required_2[x]
		fmt.Print("Installing " + value + "...")

		_, err = exec.Command("/bin/ksh", "PATH=\"$HOME:/usr/bin:/bin:/usr/sbin:/sbin:/usr/ucb\";export PATH;"+apt+" install -y "+value+" -q > /dev/null 2>&1", "ksh").Output()
		if err != nil {
			fmt.Println("Error : " + err.Error())
			os.Exit(1)
		} else {
			fmt.Println("Done")
		}
	}

	fmt.Println("Restarting Docker...")
	//Restart docker
	ioutil.WriteFile("restartDocker.sh", []byte("#!/bin/ksh\nPATH=\"$HOME:/usr/bin:/bin:/usr/sbin:/sbin:/usr/ucb\"\nexport PATH\n"+apt+" update >/dev/null 2>&1\n"+systemctl+" start docker >/dev/null 2>&1 ; sleep 2\n"+systemctl+" enable docker >/dev/null 2>&1 ; sleep 2\n/usr/bin/curl -s https://packages.cloud.google.com/apt/doc/apt-key.gpg | apt-key add -\n/usr/bin/apt-get update >/dev/null 2>&1"), 7777)
	_, _ = exec.Command("/bin/ksh", "PATH=\"$HOME:/usr/bin:/bin:/usr/sbin:/sbin:/usr/ucb\";export PATH;chmod 777 restartDocker.sh", "ksh").Output()
	//execute verifyKLists
	_, err = exec.Command("/bin/ksh", "PATH=\"$HOME:/usr/bin:/bin:/usr/sbin:/sbin:/usr/ucb\";export PATH;./restartDocker.sh", "ksh").Output()
	if err != nil {
		fmt.Println("Error Restarting Docker: " + err.Error())
		os.Exit(1)
	} else {
		fmt.Println("Done Restarting Docker")
	}
	//delete restartDocker.sh
	os.Remove("restartDocker.sh")

	//Adding apt-key
	fmt.Print("Adding apt-key...")
	_, err = exec.Command("/bin/ksh", "PATH=\"$HOME:/usr/bin:/bin:/usr/sbin:/sbin:/usr/ucb\";export PATH;curl -s https://packages.cloud.google.com/apt/doc/apt-key.gpg | apt-key add - >/dev/null 2>&1").Output()
	if err != nil {
		fmt.Println("Error : " + err.Error())
		os.Exit(1)
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
		fmt.Println("Error verify Kubernetes list : " + err.Error())
		os.Exit(1)
	} else {
		fmt.Println("Done verifying Kubernetes list")
	}
	//delete verifyList.sh
	os.Remove("verifyKList.sh")

	//Perform APT Update
	_, _ = exec.Command("/bin/ksh", "PATH=\"$HOME:/usr/bin:/bin:/usr/sbin:/sbin:/usr/ucb\";export PATH;"+apt+" update > /dev/null 2>&1").Output()
	fmt.Println("Done updating apt")

	//Install K8S Required
	for x := 0; x < len(k8s_required); x++ {
		value := k8s_required[x]
		fmt.Print("Installing " + value + "...")

		_, err = exec.Command("/bin/ksh", "PATH=\"$HOME:/usr/bin:/bin:/usr/sbin:/sbin:/usr/ucb\";export PATH;"+apt+" install -y "+value+" -q > /dev/null 2>&1", "ksh").Output()
		if err != nil {
			fmt.Println("Error : " + err.Error())
			os.Exit(1)
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
	fmt.Print("Cloning git repo for cri-tools...")
	_, _ = exec.Command("/bin/ksh", "PATH=\"$HOME:/usr/bin:/bin:/usr/sbin:/sbin:/usr/ucb\";export PATH;git clone https://github.com/kubernetes-incubator/cri-tools.git >/dev/null 2>&1").Output()

	//Begin readme

	//generate readme.sh
	ioutil.WriteFile("readme.sh", []byte("PATH=\"$HOME:/usr/bin:/bin:/usr/sbin:/sbin:/usr/ucb\"\nexport PATH\n\ncat << REOF > $1\n#-----------------------------------------------------------------------------------------\n# Added by Aporeto setup `date`\n#-----------------------------------------------------------------------------------------\nCommands you may wish to try:\n\n 1) kubectl get pods --all-namespaces (check to see if pods are running)\n	2) kubectl -n guestbook get svc front-end (interact with frontend)\n	3) kubectl delete namespace guestbook (to remove the guestbook)\n\n#-----------------------------------------------------------------------------------------\nPlease review the following file just created for additional information.\n#-----------------------------------------------------------------------------------------\n\n\n\n REOF"), 7777)
	//give permission
	_, _ = exec.Command("/bin/ksh", "PATH=\"$HOME:/usr/bin:/bin:/usr/sbin:/sbin:/usr/ucb\";export PATH;chmod 777 readme.sh", "ksh").Output()
	//execute
	_, err = exec.Command("/bin/ksh", "PATH=\"$HOME:/usr/bin:/bin:/usr/sbin:/sbin:/usr/ucb\";export PATH;./readme.sh", out, "ksh").Output()
	if err != nil {
		fmt.Println("Error creating readme : " + err.Error())
	} else {
		fmt.Println("Done creating readme")
	}
	//delete readme.sh
	os.Remove("readme.sh")

	//Create a network

	cidr := ""
	fmt.Print("For our cluster, please enter a CIDR address (example: 192.168.1.1/24): ")
	_, _ = fmt.Scanln(&cidr)
	fmt.Print("Kubeadm Init with " + cidr + " starting and key writeout...")
	//ioutil.WriteFile("network.sh", []byte("PATH=\"$HOME:/usr/bin:/bin:/usr/sbin:/sbin:/usr/ucb\"\nexport PATH\n\n	$1 init --pod-network-cidr=$2 >> $3 \n	print \"Done\"\n\n	#\n	#	Setup $HOME\n	#\n	print \"Home Path : $HOME\"\n\n	if [ ! -d $HOME/.kube ] ; then\n		mkdir -p $HOME/.kube\n 	cp -i /etc/kubernetes/admin.conf $HOME/.kube/config\nchown $(id -u):$(id -g) $HOME/.kube/config\n\n#\n	#  	Deploy Project Calico\n	#\n\n	$4 apply -f https://docs.projectcalico.org/v2.6/getting-started/kubernetes/installation/hosted/kubeadm/1.6/calico.yaml\nelse\n	#\n	#	skip if Exists\n#\n\nprint \"$prog: $HOME/.kube Exists...skipping\" 1>&2\ncontinue\n	fi\n\n\n	#\n	#	Untaint the Master \n	#\n\n	print \"Untaint of the master so it will be available for scheduling workloads...\\c\"\n$4 taint nodes --all node-role.kubernetes.io/master- >/dev/null 2>&1\n	print \"Done\"\n\n\nprint \"Bind to k8s...\"\n	$4 create clusterrolebinding aporeto --clusterrole=cluster-admin --serviceaccount=kube-system:default\n	print \"Done\"\n"), 7777)
	ioutil.WriteFile("network.sh", []byte("PATH=\"$HOME:/usr/bin:/bin:/usr/sbin:/sbin:/usr/ucb\"\nexport PATH\n\n	$1 init --pod-network-cidr=$2 >> $3 \n	print \"Done\"\n\n	#\n	#	Setup $HOME\n	#\n\n	if [ ! -d $HOME/.kube ] ; then\n		print \"Home Directory : $HOME\"		mkdir -p $HOME/.kube\n		cp -i /etc/kubernetes/admin.conf $HOME/.kube/config\n		chown $(id -u):$(id -g) $HOME/.kube/config\n\n#\n	#  	Deploy Project Calico\n	#\n\n		$4 apply -f https://docs.projectcalico.org/v2.6/getting-started/kubernetes/installation/hosted/kubeadm/1.6/calico.yaml\n	else\n	#\n	#	skip if Exists\n#\n\n		print \"$prog: $HOME/.kube Exists...skipping\" 1>&2\n		continue\n	fi\n\n\n	#\n	#	Untaint the Master \n	#\n\n	print \"Untaint of the master so it will be available for scheduling workloads...\\c\"\n	$4 taint nodes --all node-role.kubernetes.io/master- >/dev/null 2>&1\n	print \"Done\"\n\n\n	print \"Bind to k8s...\"\n	$4 create clusterrolebinding aporeto --clusterrole=cluster-admin --serviceaccount=kube-system:default\n	print \"Done\"\n"), 7777)
	//give permission
	_, _ = exec.Command("/bin/ksh", "PATH=\"$HOME:/usr/bin:/bin:/usr/sbin:/sbin:/usr/ucb\";export PATH;chmod 777 network.sh", "ksh").Output()
	//execute
	_, _ = exec.Command("/bin/ksh", "PATH=\"$HOME:/usr/bin:/bin:/usr/sbin:/sbin:/usr/ucb\";export PATH;./network.sh", kubeadm, cidr, out, kubectl, "ksh").Output()
	//delete network.sh
	os.Remove("network.sh")

	//Deploy application

	app_r := ""
	fmt.Print("Ready to deploy an application? [y/n] : ")
	_, _ = fmt.Scanln(&app_r)

	//Take users input

	app_r = strings.ToLower(app_r)

	if app_r == "y" {
		ioutil.WriteFile("deploy.sh", []byte("PATH=\"$HOME:/usr/bin:/bin:/usr/sbin:/sbin:/usr/ucb\"\nexport PATH\n\n$1 create namespace guestbook\n$1 apply -n guestbook -f \"https://raw.githubusercontent.com/dnester/guestbook/master/guestbook.yaml\""), 7777)
		//give permission
		_, _ = exec.Command("/bin/ksh", "PATH=\"$HOME:/usr/bin:/bin:/usr/sbin:/sbin:/usr/ucb\";export PATH;chmod 777 deploy.sh", "ksh").Output()
		//execute
		_, _ = exec.Command("/bin/ksh", "PATH=\"$HOME:/usr/bin:/bin:/usr/sbin:/sbin:/usr/ucb\";export PATH;./deploy.sh", kubectl, "ksh").Output()
		//delete deploy.sh
		os.Remove("deploy.sh")
	} else if app_r == "n" {
		//nothing do herer
	} else {
		ioutil.WriteFile("deploy.sh", []byte("PATH=\"$HOME:/usr/bin:/bin:/usr/sbin:/sbin:/usr/ucb\"\nexport PATH\n\n$1 create namespace guestbook\n$1 apply -n guestbook -f \"https://raw.githubusercontent.com/dnester/guestbook/master/guestbook.yaml\""), 7777)
		//give permission
		_, _ = exec.Command("/bin/ksh", "PATH=\"$HOME:/usr/bin:/bin:/usr/sbin:/sbin:/usr/ucb\";export PATH;chmod 777 deploy.sh", "ksh").Output()
		//execute
		_, _ = exec.Command("/bin/ksh", "PATH=\"$HOME:/usr/bin:/bin:/usr/sbin:/sbin:/usr/ucb\";export PATH;./deploy.sh", kubectl, "ksh").Output()
		//delete deploy.sh
		os.Remove("deploy.sh")
	}

	//At this point you should have a fully-functional kubernetes cluster on which you can run workloads.
	//Now we will download and install the required Aporeto components.

	aposoft_r := ""
	fmt.Print("Ready to download the required Aporeto components? [y/n] : ")
	_, _ = fmt.Scanln(&aposoft_r)

	//Take users input for downloading apoctl and enforcerd

	aposoft_r = strings.ToLower(aposoft_r)

	switch aposoft_r {
	case "y":
		install_apoctl(apoctl, enforcerd, systemctl, apt, curl, config)
		break
	case "n":
		//nothing to do.. continue
		break
	default:
		install_apoctl(apoctl, enforcerd, systemctl, apt, curl, config)
		break
	}

	//Final message.

	fmt.Println("")
	fmt.Println("")
	fmt.Println("The K8 cluster is now coming up.  Please test with the following command:")
	fmt.Println("")
	fmt.Println("         $kubectl get pods --all-namespaces")
	fmt.Println("")
	fmt.Println("")
	fmt.Println(prog + " run on `/bin/hostname` completed on " + nowtime.Format("2006-Mar-02 15:04:05") + ". Be sure to review the installation readme in " + out)
	fmt.Println("")

	os.Remove(tmp)
	os.Remove(tmp2)
	os.Exit(0)

	//============================================================================
	// 	END of dk8s
	//============================================================================
}

func install_apoctl(apoctl, enforcerd, systemctl, apt, curl string, config map[string]interface{}) {
	fmt.Println("made it here")

	fmt.Print("Verifying apoctl... ")
	fmt.Println(apoctl, enforcerd)
	var outBytes []byte
	_, err := exec.Command("/bin/ksh", "PATH=\"$HOME:/usr/bin:/bin:/usr/sbin:/sbin:/usr/ucb\";export PATH;"+apoctl, "ksh").Output()
	if err != nil {
		fmt.Print("Downloading apoctl... ")
		_, _ = exec.Command("/bin/ksh", "PATH=\"$HOME:/usr/bin:/bin:/usr/sbin:/sbin:/usr/ucb\";export PATH;"+curl+" -o "+apoctl+" https://download.aporeto.com/releases/release-1.3.1-r9/apoctl/linux/apoctl").Output()
		_, _ = exec.Command("/bin/ksh", "PATH=\"$HOME:/usr/bin:/bin:/usr/sbin:/sbin:/usr/ucb\";export PATH;chmod 775 "+apoctl, "ksh").Output()
		fmt.Println("Done")
	} else {
		fmt.Println("Nothing to do")
	}

	//Enforcerd
	fmt.Print("Verifying Enforcerd... ")
	_, err = exec.Command("/bin/ksh", "PATH=\"$HOME:/usr/bin:/bin:/usr/sbin:/sbin:/usr/ucb\";export PATH;"+enforcerd, "ksh").Output()
	if err != nil {
		fmt.Print("Downloading Enforcerd... ")
		_, _ = exec.Command("/bin/ksh", "PATH=\"$HOME:/usr/bin:/bin:/usr/sbin:/sbin:/usr/ucb\";export PATH;"+curl+" -o enforcerd.amd64.deb https://download.aporeto.com/releases/release-1.3.1-r9/enforcerd/linux/enforcerd.amd64.deb").Output()
		_, _ = exec.Command("/bin/ksh", "PATH=\"$HOME:/usr/bin:/bin:/usr/sbin:/sbin:/usr/ucb\";export PATH;"+apt+" install -y ./enforcerd.amd64.deb -q ; sleep 2", "ksh").Output()
		fmt.Println("Done")
	} else {
		fmt.Println("Nothing to do")
	}

	//Let's register our Demo Namespace
	r_username := config["username"].(string)
	if r_username == "" {
		fmt.Print("Please enter your Aporeto username: : ")
		_, _ = fmt.Scanln(&r_username)
	}

	r_password := config["password"].(string)
	if r_password == "" {
		fmt.Print("Please enter your Aporeto password: : ")
		_, _ = fmt.Scanln(&r_password)
	}

	//apoctl registration

	var TOKEN string
	outBytes, err = exec.Command("/bin/ksh", "PATH=\"$HOME:/usr/bin:/bin:/usr/sbin:/sbin:/usr/ucb\";export PATH;apoctl auth aporeto --account "+r_username+" --password "+r_password+" --validity 2000m").Output()
	if err != nil {
		fmt.Println("Error : Failed to get token. Reason : " + err.Error())
		os.Exit(1)
	} else {
		TOKEN = string(outBytes)
		fmt.Println("TOKEN : " + TOKEN)
	}

	//Get namespace
	NAMESPACE := config["namespace"].(string)
	if NAMESPACE == "" {
		fmt.Print("Please enter your namespace : ")
		_, _ = fmt.Scanln(&NAMESPACE)
	}

	/*var curlme string
	outBytes, err = exec.Command("/bin/ksh", "PATH=\"$HOME:/usr/bin:/bin:/usr/sbin:/sbin:/usr/ucb\";export PATH;"+curl+" -s -X POST -H \"Content-Type: application/json\" -H \"Authorization: Bearer "+TOKEN+"\" -d '{\"name\":\"apodemo3\",\"targetNamespace\":\""+NAMESPACE+"\"}' \"https://api.console.aporeto.com/kubernetesclusters?\" | grep -Po '\"kubernetesDefinitions\":(\\d*?,|.*?[^\\]\",)' | awk -F\" '{print$4}' >> curlme").Output()
	if err != nil {
		fmt.Println("Error : Failed to execute curl. Reason : " + err.Error())
		fmt.Println("Another object exists with the same key... Please delete your cluster before you reinitiate..")
		os.Exit(1)
	} else {
		curlme = string(outBytes)
		fmt.Println("curlme : " + curlme)
	}*/

	URL := "https://api.console.aporeto.com/kubernetesclusters"
	Method := "POST"
	retryCount := 0
	excludeResponseCodes := make([]string, 1)
	excludeResponseCodes[0] = "201"

	headerTokens := make(map[string]string)
	headerTokens["Authorization"] = "Bearer " + TOKEN
	headerTokens["Content-Type"] = "application/json"

	postMap := make(map[string]interface{})
	postMap["name"] = "aporetodemo10"
	postMap["targetNamespace"] = NAMESPACE

	bytess, _ := json.Marshal(postMap)

	mapObject := make(map[string]interface{})

	err, response := SendRequest(Method, URL, string(bytess), headerTokens, excludeResponseCodes, retryCount)
	if err != nil {
		fmt.Println(err.Error())
		fmt.Println("Resource Kubernetes cluster already available. Please delete it from app console first...")
	} else {
		json.Unmarshal([]byte(response), &mapObject)
		ioutil.WriteFile("./curlme", []byte(mapObject["kubernetesDefinitions"].(string)), 0666)
	}

	_, _ = exec.Command("/bin/ksh", "PATH=\"$HOME:/usr/bin:/bin:/usr/sbin:/sbin:/usr/ucb\";export PATH;mkdir ./myaporeto/;/usr/bin/base64 -d curlme >> ./myaporeto/myfiles").Output()
	_, _ = exec.Command("/bin/ksh", "PATH=\"$HOME:/usr/bin:/bin:/usr/sbin:/sbin:/usr/ucb\";export PATH;mv myaporeto/myfiles myaporeto/myfiles.tar.gz ; cd myaporeto;gzip -df myfiles.tar.gz;tar xvf myfiles.tar;").Output()

	files, err := ioutil.ReadDir("./myaporeto/")
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	for _, f := range files {
		if !strings.Contains(f.Name(), "myfiles") { //excluding myfiles.tar
			_, err = exec.Command("/bin/ksh", "PATH=\"$HOME:/usr/bin:/bin:/usr/sbin:/sbin:/usr/ucb\";export PATH;kubectl create -f ./myaporeto/"+f.Name()).Output()
			if err != nil {
				fmt.Println("Pod for " + f.Name() + " already exists... Please remove it first...")
			} else {
				fmt.Println("Done : " + f.Name())
			}
		}
	}

	time.Sleep(5 * time.Second)

	_, err = exec.Command("/bin/ksh", "PATH=\"$HOME:/usr/bin:/bin:/usr/sbin:/sbin:/usr/ucb\";export PATH;/lib/systemd/systemd-sysv-install enable enforcerd").Output()
	if err != nil {
		fmt.Println("Error : Enabling enforcerd. Reason : " + err.Error())
		os.Exit(1)
	} else {
		fmt.Println("Enabled enforcerd")
	}

	time.Sleep(5 * time.Second)

	_, _ = exec.Command("/bin/ksh", "PATH=\"$HOME:/usr/bin:/bin:/usr/sbin:/sbin:/usr/ucb\";export PATH;"+systemctl+" start enforcerd").Output()
	/*if err != nil {
		fmt.Println("Error : Starting enforcerd. Reason :  " + err.Error())
		os.Exit(1)
	} else {
		fmt.Println("Started enforcerd")
	}*/

	time.Sleep(5 * time.Second)

	fmt.Println("Services restarted.  Please check the aporeto web console for the registered enforcerd agents")

}

func SendRequest(Method string, url string, BodyString string, headerTokens map[string]string, excludeResponseCodes []string, retryCount int) (err error, response string) {
	retryCount = 0
	var req *http.Request
	var resp *http.Response
	req, err = http.NewRequest(Method, url, bytes.NewBuffer([]byte(BodyString)))
	if err != nil {
		response = err.Error()
	} else {
		var body []byte

		if len(headerTokens) > 0 {
			for key, value := range headerTokens {
				req.Header.Set(key, value)
			}
		}

		client := &http.Client{}
		resp, err = client.Do(req)

		if err != nil {
			if retryCount > 0 && !strings.Contains(err.Error(), "permission") && (strings.Contains(err.Error(), "dial tcp") || strings.Contains(err.Error(), "EOF")) {
				err = nil
				time.Sleep(1 * time.Second)
				return SendRequest(Method, url, BodyString, headerTokens, excludeResponseCodes, (retryCount - 1))
			} else {
				response = "Error : " + err.Error()
			}
		} else {
			if resp.Body != nil {
				defer resp.Body.Close()
			}

			body, err = ioutil.ReadAll(resp.Body)
			if err != nil {
				response = "Error : " + err.Error()
			} else {
				excludeResponses := ""

				if len(excludeResponseCodes) > 0 {
					for _, value := range excludeResponseCodes {
						excludeResponses += value + " "
					}
				}

				if resp.StatusCode != 200 && !strings.Contains(excludeResponses, strconv.Itoa(resp.StatusCode)) {
					if len(body) < 4 {
						err = errors.New("Error! Empty Response Body!")
					} else {
						err = errors.New(string(body))

					}
				} else {
					response = string(body)
				}
			}
		}

	}

	if strings.Contains(response, "Error :") {
		//fmt.Println(response)
		if err != nil {
			//fmt.Println(err.Error())
		}
	} else {
		//fmt.Println("HTTP_DefaultRequest successful.")
		//fmt.Println("HTTP Response : " + response)
	}

	return
}

func GetConfig() (config map[string]interface{}) {
	config = make(map[string]interface{})

	content, err := ioutil.ReadFile("./config.json")

	if err != nil {
		//config file not Available// regenerate it...
		fmt.Println("WARNING : Configuration file not found. Generating an empty configuration file. Please update username and password.")
		config["username"] = ""
		config["password"] = ""
		config["namespace"] = ""
		byteArray, _ := json.Marshal(config)
		_ = ioutil.WriteFile("./config.json", byteArray, 0777)
	} else {
		json.Unmarshal(content, &config)
	}

	return
}
