#!/bin/bash

if [[ "" != $(which lxd) ]]
then
	LXD_VERSION=$(lxd version)
	if dpkg --compare-version 3.0.3 le "${LXD_VERSION}" ;
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

