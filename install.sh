#!/bin/bash
set -e

orgdir=`pwd`
if [ ! -d $HOME/go ];then
    cd $HOME
    wget https://go.dev/dl/go1.18.linux-amd64.tar.gz
    tar -xfz go1.18.linux-amd64.tar.gz
fi

if [ ! -f "go.mod" ];then
    go mod init goLogAnalyzer
fi
go mod tidy

echo "Installing goLogAnalyzer"
echo go build -o logan main.go
go build -o logan main.go
echo "sudo cp logan /usr/local/bin/logan"
sudo cp logan /usr/local/bin/logan
echo "OK!"
echo ""
