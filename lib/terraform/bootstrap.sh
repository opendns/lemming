#!/bin/bash

# function to install packages using apt
# usage: apt_install <packageName>
function apt_install {
    sudo apt-get -qq update
    sudo apt-get -yf install $1
}

# Workaround the ncurses mysql package prompt
sudo debconf-set-selections <<< 'mysql-server mysql-server/root_password password password'
sudo debconf-set-selections <<< 'mysql-server mysql-server/root_password_again password password'

# Setup MySQL on the host
apt_install mysql-server
