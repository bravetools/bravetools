#!/bin/bash

if [[ "" != $(which lxd) ]]
then
	LXD_VERSION=$(lxd version)
	LXD_VERSION_CLEANED=$(echo "$LXD_VERSION" | sed 's/ .*//')
	dpkg --compare-versions 3.0.3 le "${LXD_VERSION_CLEANED}" 
	if [[ $? != 0 ]]	
	then
		echo 'lxd version not new enough' ;
		exit 1 ;
	fi
else
	echo 'lxd is not installed' ;
	exit 1 ;
fi


if [[ "" == $(which go) ]]
then
	echo 'go is not installed' ;
	exit 1 ;
fi
