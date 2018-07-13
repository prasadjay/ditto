#!/bin/ksh 

PATH="$HOME:/usr/bin:/bin:/usr/sbin:/sbin:/usr/ucb"
export PATH

file="/etc/apt/sources.list.d/kubernetes.list"
if [ ! -f "$file" ] ; then
	print "$1: $file not present.  Creating...\\c"
	print "\n"
	touch $file
cat << SEOF > $file
# Added by Aporeto setup `date`
deb http://apt.kubernetes.io/ kubernetes-xenial main
# Aporeto-end
SEOF
else [ `head -10 $file | fgrep 'Added by Aporeto'|wc -w` -gt 0 ] 
	print "$1: $file already added.  Skipping...\\c"

	#
	#	sanity check
	#

	if [ $? != 0 ] ; then
		print "$1: FATAL: modification of $file failed!  Quitting."
		rm -f $2 $3
		exit 1
	fi
	cat "$file" >> $2
	if [ $? != 0 ] ; then
		print "$1: FATAL: modification of $file failed!  Quitting."
		rm -f $2 $3
		exit 1
	fi
	cat $2 > "$file"
	if [ $? != 0 ] ; then
		print "$1: FATAL: update of $file failed!  Quitting."
		rm -f $2 $3
		exit 1
	fi
fi