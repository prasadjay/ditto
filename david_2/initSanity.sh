#!/bin/ksh 

PATH="$HOME:/usr/bin:/bin:/usr/sbin:/sbin:/usr/ucb"
export PATH


#
# 	initial sanity checks
#

if [ ! -x $curl ] ; then 
	print "$1: SANITY: /usr/bin/curl missing!  Will install later ..." 1>&2
	continue
fi

if [ ! -x $apt ] ; then
	print "$1: SANITY: /usr/bin/apt-get missing!  Quitting..." 1>&2
	rm -f $tmp $tmp2
	exit 1
fi
if [ ! -x $nawk ] ; then
	print "$1: SANITY: /usr/bin/nawk missing!  Quitting..." 1>&2
	rm -f $tmp $tmp2
	exit 1
fi
if [ "x$ostype" != "xLinux 4" ] ; then
	print "$1: FATAL: sorry, OS type $ostype not supported" 1>&2
	exit 1
fi
case "$1" 
in
	"")	
	;;
	*)	
	print "$usage: $prog - no flags at this time " 1>&2
	exit 1 
	;;
esac
if [ ! -x $id ] ; then
	print "$prog: SANITY: /usr/bin/id missing!  Quitting..." 1>&2
	rm -f $tmp $tmp2
	exit 1
fi
if [ ! -x $swapoff ] ; then
	print "$prog: SANITY: /sbin/swapoff failed!  Quitting..." 1>&2
	rm -f $tmp $tmp2
	exit 1
fi
id=`/usr/bin/id`
if [ $? != 0 ] ; then
	print "$prog: SANITY: /usr/bin/id failed!  Quitting..." 1>&2
	rm -f $2 $3
	exit 1
fi
case "$id" in
uid=0*)	;;
*)	print "$prog: FATAL: must be run as root" 1>&2
	exit 1 ;;
esac
if [ ! -x /bin/df ] ; then
	print "$prog: SANITY: /bin/df missing!  Quitting..." 1>&2
	rm -f $tmp $tmp2
	exit 1
fi

#
# 	handle CTRL-C
#

trap 'print "" ; print Killed | tee $log ; rm -f $tmp $tmp2 ; print "" ; exit 2' 2
(
clear
print ""
print ""
print ""
print "-----------------------------------------------------------------"
print "Aporeto K8S Quick Start [ $prog $version ]"
print "Installation of Kubernetes on Ubuntu"
print "-----------------------------------------------------------------"
print "running on                 `/bin/hostname`"
print "date:                      `date`"
print "log file:                  $log"
print "Install Readme:            $out"
print ""
print "Press ENTER to continue, else CTRL-C to quit "\\c
read foo
print ""
print ""
print ""


#
#	APT-Related Requirements
#

apt_required_1="
ebtables
ethtool
"

apt_required_2="
docker.io
golang
git
apt-transport-https
curl
"

k8s_required="
kubelet 
kubeadm 
kubectl
"


#
#	Install apt-required
#


for i in $apt_required_1 ; do 
	print "Installing $i...\\c"
	$apt install -y $i -q >/dev/null 2>&1 
	print "Done"
done

#
#	Perform APT Update
#

print "Updating apt...\\c"
$apt update >/dev/null 2>&1
print "Done"


#
#	Install apt-required 2
#


for j in $apt_required_2 ; do 
	print "Installing $j...\\c"
	$apt install -y $j -q >/dev/null 2>&1 
	print "Done"
done


#
#	Restart Docker
#


print "Restarting Docker...\\c"
$systemctl start docker >/dev/null 2>&1 ; sleep 2
$systemctl enable docker >/dev/null 2>&1 ; sleep 2
print "Done"


#
#	Adding apt-key
#


print "Adding apt-key...\\c"
$curl -s https://packages.cloud.google.com/apt/doc/apt-key.gpg | apt-key add - >/dev/null 2>&1
print "Done"


#
# 	check/modify /etc/apt/sources.list.d/kubernetes.list
#


file="/etc/apt/sources.list.d/kubernetes.list"
if [ ! -f "$file" ] ; then
	print "$prog: $file not present.  Creating...\\c"
	print "\n"
	touch $file
cat << SEOF > $file
# Added by Aporeto setup `date`
deb http://apt.kubernetes.io/ kubernetes-xenial main
# Aporeto-end
SEOF
else [ `head -10 $file | fgrep 'Added by Aporeto'|wc -w` -gt 0 ] 
	print "$prog: $file already added.  Skipping...\\c"

	#
	#	sanity check
	#

	if [ $? != 0 ] ; then
		print "$prog: FATAL: modification of $file failed!  Quitting."
		rm -f $tmp $tmp2
		exit 1
	fi
	cat "$file" >> $tmp
	if [ $? != 0 ] ; then
		print "$prog: FATAL: modification of $file failed!  Quitting."
		rm -f $tmp $tmp2
		exit 1
	fi
	cat $tmp > "$file"
	if [ $? != 0 ] ; then
		print "$prog: FATAL: update of $file failed!  Quitting."
		rm -f $tmp $tmp2
		exit 1
	fi
	print "Done"
fi


#
#	Install K8S Required
#

print "Updating apt...\\c"
$apt update >/dev/null 2>&1
print "Done"

for k in $k8s_required ; do 
	print "Installing $k...\\c"
	$apt install -y $k -q >/dev/null 2>&1 
	print "$k Done"
done

	
#
#	Check for kubeadm, kubelet, kubectl
#

if [ ! -x $kubeadm ] ; then
	print "$prog: SANITY: /usr/bin/kubeadm missing!  Quitting..." 1>&2
	rm -f $tmp $tmp2
	exit 1
fi
if [ ! -x $kubelet ] ; then
	print "$prog: SANITY: /usr/bin/kubelet missing!  Quitting..." 1>&2
	rm -f $tmp $tmp2
	exit 1
fi
if [ ! -x $kubectl ] ; then
	print "$prog: SANITY: /usr/bin/kubectl missing!  Quitting..." 1>&2
	rm -f $tmp $tmp2
	exit 1
fi


#
#	Clone cri-tools repo
#


print "Cloning git repo for cri-tools...\\c"
$git clone https://github.com/kubernetes-incubator/cri-tools.git >/dev/null 2>&1
print "Done"

#
#	Begin readme
#

cat << REOF > $out
#-----------------------------------------------------------------------------------------
# Added by Aporeto setup `date`
#-----------------------------------------------------------------------------------------
Commands you may wish to try:
	
	1) kubectl get pods --all-namespaces (check to see if pods are running)
	2) kubectl -n guestbook get svc front-end (interact with frontend)
	3) kubectl delete namespace guestbook (to remove the guestbook)


#-----------------------------------------------------------------------------------------
Please review the following file just created for additional information.
#-----------------------------------------------------------------------------------------




REOF


# ================================================================
#
#		Begin Functions
#
# ================================================================



#
#	Install apoctl
#

install_apoctl () {

	if [ ! -x $apoctl ] ; then		
			print "Downloading apoctl....\\c"
			$curl -o $apoctl https://download.aporeto.com/releases/release-1.3.1-r9/apoctl/linux/apoctl
			chmod 755 $apoctl
			print "Done"
	else 
			continue 
	fi

}

#
#	Install Enforcer
#

install_enforcer() {

	if [ ! -x $enforcerd ] ; then		
		print "Downloading enforcerd....\\c"
		#$apt update 
		$curl -o enforcerd.amd64.deb https://download.aporeto.com/releases/release-1.3.1-r9/enforcerd/linux/enforcerd.amd64.deb
		$apt install -y ./enforcerd.amd64.deb -q ; sleep 2

	else 
		register_linux 
	fi
		
}


#
#	Register with Kubernetes
#


register_kub(){

		#
		#	Register Enforcer 
		#

		print "Please enter your Aporeto username: \\c"
			read r_username
		
			#
			#	apoctl registration
			#

			KUB_TOKEN=$(apoctl auth aporeto --account $r_username --validity 2m)

			print "Please enter your namespace: \\c"
				read KUB_NAMESPACE

			$apoctl account create-k8s-cluster demo --target-namespace=$KUB_$NAMESPACE

			#
			#	enforcerd registration / might need to exec.
			#

		
			tar xzvf my-first-cluster.tgz

			#for i in yaml output
			#	$kubectl create -f $i
			#done
			
			#
			#	Restarting enforcerd services
			#

			#$systemctl enable enforcerd  ; sleep 3
			#$systemctl start enforcerd  ; sleep 3

			print "Services restarted.  Please check the aporeto web console for the registered enforcerd agents"
			exit 0 
}

#
#	End kubernetes registration
#


#
#	Installation for Linux
#


register_linux(){

		#
		#	Register Enforcer 
		#

		print "Please enter your Aporeto username: \\c"
			read a_username
		
			#
			#	apoctl registration
			#

			AUTH_TOKEN=$($apoctl auth aporeto --account $a_username -e)
			print "Done"
	
			#
			#	enforcerd registration / might need to exec.
			#
		
			print "Requesting enforcerd token...\\c"
			APOCTL_TOKEN=$($apoctl auth aporeto --account $a_username --validity 2m)

			print "Please enter the namespace you registered in Aporeto console: \\c"
				read NAMESPACE

				$enforcerd register --token $APOCTL_TOKEN --squall-namespace $NAMESPACE 
				print "Done"
			
			#
			#	Restarting enforcerd services
			#

			$systemctl enable enforcerd  ; sleep 3
			$systemctl start enforcerd  ; sleep 3

			print "Services restarted.  Please check the aporeto web console for the registered enforcerd agents"
}


#
#	End Linux Registration
#

# ================================================================
#
#		End Functions
#
# ================================================================



#
#	Create a network
#

print "For our cluster, please enter a CIDR address (example: 192.168.1.1/24): \\c"
	read cidr
	
	print "Kubeadm Init with $cidr starting and key writeout...\\c"
	$kubeadm init --pod-network-cidr=$cidr >> $out 
	print "Done"

	#
	#	Setup $HOME
	#
	
	if [ ! -d $HOME/.kube ] ; then
		mkdir -p $HOME/.kube
	  	cp -i /etc/kubernetes/admin.conf $HOME/.kube/config
	    	chown $(id -u):$(id -g) $HOME/.kube/config

		#
		#  	Deploy Project Calico
		#

		$kubectl apply -f https://docs.projectcalico.org/v2.6/getting-started/kubernetes/installation/hosted/kubeadm/1.6/calico.yaml
	else

		#
		#	skip if Exists
		#

		print "$prog: $HOME/.kube Exists...skipping" 1>&2
		continue
	fi


	#
	#	Untaint the Master 
	#

	print "Untaint of the master so it will be available for scheduling workloads...\\c"
	$kubectl taint nodes --all node-role.kubernetes.io/master- >/dev/null 2>&1
	print "Done"

	#	
	#	Deploy application
	#

	print "Ready to deploy an application? [yn] \\c"
		read app_r

	#
	#	Take users input
	#

	case "$app_r" in
        ""|"y"|"Y")
			print "Creating Namespace and Grabbing application....\\c"
			$kubectl create namespace guestbook
			$kubectl apply -n guestbook -f "https://raw.githubusercontent.com/dnester/guestbook/master/guestbook.yaml"
			print "Done"
        ;;
        "n"|"N")
			continue
		;;									            
		*)
			print "Creating Namespace and Grabbing application....\\c"
			$kubectl create namespace guestbook
			$kubectl apply -n guestbook -f "https://raw.githubusercontent.com/dnester/guestbook/master/guestbook.yaml"
			print "Done"
		;;
		esac 

	#
	#	At this point you should have a fully-functional kubernetes cluster on which you can run workloads.
	#


#
#	Final message.
#

print ""
print ""
print "The K8 cluster is now coming up.  Please test with the following command:"
print ""
print "         $kubectl get pods --all-namespaces"
print ""
print ""
print "$prog run on `/bin/hostname` completed on `date`.  Be sure to review the "
print "installation readme in $out"
print ""

) | tee $log

rm -f $tmp $tmp2
exit 0





# ============================================================================
# 	END of dk8s
# ============================================================================
