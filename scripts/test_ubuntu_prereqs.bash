#!/bin/bash

if [[ "" != $(which lxd) ]]
then
	LXD_VERSION=$(lxd version)
	dpkg --compare-versions 3.0.3 le "${LXD_VERSION}" 
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

if [[ "" == $(which zfs) ]]
then
  echo 'ZFS not found. Run `sudo apt install zfsutils-linux`';
  exit 1;
fi
