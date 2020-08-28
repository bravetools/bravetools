#!/usr/bin/env bash

VERSION=$(cat ../VERSION)
BINARY_NAME="brave"
APP_DIR="$HOME/.bravetools"

COMMAND=${0##*/}

usage() {
  echo "
USAGE: bash $COMMAND [mac|ubuntu] FLAGS

FLAGS:
  -f   forces install, overwriting all existing folders, files, and configurtaions:
        \$HOME/.bravetools folder containing pre-built images
        non-snap LXD will be removed and updated to the latest snap version

Installs Bravetools either on a local Multipass instance (mac) or
a native Ubuntu installation (≥18.04).

The default installation directory is \$HOME/.bravetools

Multipass install: https://multipass.run
LXD install: sudo snap install lxd
" >&2
  exit 2
}

prepare_mac() {

  if [[ ! -e $APP_DIR ]]; then
    mkdir -p $HOME/.bravetools/images
    mkdir -p $HOME/.bravetools/certs
    touch $APP_DIR/config.yml

    cp config-multipass.yml $APP_DIR/config.yml
    cp darwin/brave /usr/local/bin/$BINARY_NAME
  fi
}

prepare_linux() {
  WHICH_LXD=$(which lxd)
  if [ "$WHICH_LXD" == "/usr/bin/lxd" ]; then
    if [ "$FORCE" != "f" ]; then
      echo "Bravetools requires snap LXD and you have ${WHICH_LXD}. To upgrade to snap lxd execute:

sudo apt remove -y lxd
sudo apt autoremove -u
sudo apt purge
sudo snap install lxd
"
      exit 2
    fi
  fi

  if [[ ! -e $APP_DIR ]]; then
    mkdir -p $HOME/.bravetools/images
    mkdir -p $HOME/.bravetools/certs
    touch $APP_DIR/config.yml

    cp config-lxd.yml $APP_DIR/config.yml
    sudo cp ubuntu/brave /usr/bin/
  fi
}

cleanup_mac() {
  if [ -d "$APP_DIR" ]; then
    if [ "$FORCE" != "f" ]; then
    echo "
Found existing Bravetools installation. Ensure your images and units are backed up.
Execute rm -r ${APP_DIR} and re-try installation"
      exit 2
    fi
  fi
  rm -rf $APP_DIR

  MPID=$(pgrep multipass)
  if [ "$MPID" != "" ]; then
    multipass delete brave
    multipass purge
  fi

  if [[ -f "/usr/local/bin/$BINARY_NAME" ]]; then
    rm /usr/local/bin/$BINARY_NAME
  fi
}

cleanup_linux() {
  if [ -d "$APP_DIR" ]; then
    if [ "$FORCE" != "f" ]; then
    echo "
Found existing Bravetools installation. Ensure your images and units are backed up.
Execute rm -r ${APP_DIR} and re-try installation"
      exit 2
    fi
  fi

  LPID=$(pgrep lxd)
  if [ "$LPID" != "" ]; then
    UNITS=$(lxc storage info brave | grep -A3 'used by:' | grep -A3 'containers:' | awk '{ print $2}')
    lxc rm -f $UNITS
    lxc profile delete brave

    STORAGE=$(lxc storage list --format csv)
    POOLNAME=$(cut -d',' -f1 <<<"$STORAGE")

    lxc storage delete $POOLNAME
  fi

  rm -rf $APP_DIR

  if [[ -f "/usr/bin/$BINARY_NAME" ]]; then
    sudo rm /usr/bin/$BINARY_NAME
  fi
  
}

install_mac() {
  MPID=$(pgrep multipass)
  if [ "$MPID" == "" ]; then
    echo "Unable to locate Multipass on your system. Multipass can be installed from https://multipass.run"
    exit 1
  fi

  brave init && \
  IP=$(brave info --short true) && \
  sleep 15 && \
  sed -i '' "s/0.0.0.0/$IP/g" $APP_DIR/config.yml && \
  brave remote -i $IP
}

install_linux() {
  #LPID=$(pgrep lxd)
  #if [ "$LPID" == "" ]; then
  #  echo "Unable to located LXD. Install it using sudo snap install lxd"
  #  exit 1
  #fi

  brave init && \
  IP=$(brave info --short true) && \
  sleep 15 && \

  sed -i 's/0.0.0.0/localhost/g' $APP_DIR/config.yml && \
  brave remote -i localhost
}

print_error() {
  echo "Unable to install Bravetools " 1>&2
}

if [ $# -le 0 ]
then
  usage
  exit 2
fi

# Main Installation Script
PLATFORM=$1
FLAG=$2
FORCE=""

case $FLAG in
  -f)
  read -p "You are about to force install bravetools. This will remove any pre-built images and install a fresh snap lxd.
Proceed at your own risk. Would you still like to continue [y/N]? " -n 1 -r
  echo
  if [[ ! $REPLY =~ ^[Yy]$ ]]
  then
    exit 1
  fi
  FORCE="f"
  ;;

  "")
  FORCE=""
  ;;

  *)
    usage;;

esac

echo $FORCE

case $PLATFORM in
  mac)
    clear
    echo "Cleaning environment..."
    cleanup_mac || (print_error && exit 1)
    echo "Preparing installation..."
    prepare_mac || (print_error && cleanup_mac && exit 1)
    echo "Installing Bravetools..."
    install_mac || (print_error && cleanup_mac && exit 1)
    ;;

  ubuntu)
    clear
    echo "Cleaning environment..."
    cleanup_linux || (print_error && exit 1)
    echo "Preparing installation..."
    prepare_linux || (print_error && cleanup_mac && exit 1)
    echo "Installing Bravetools..."
    install_linux || (print_error && cleanup_mac && exit 1)
    ;;
  *)
    usage
    ;;
esac
