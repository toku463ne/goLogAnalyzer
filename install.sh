#!/bin/bash
set -e

orgdir=`pwd`
GOSRC=go1.18.linux-amd64.tar.gz
if [ ! -d $HOME/go ];then
    cd $HOME
    wget https://go.dev/dl/$GOSRC
    tar -xzf $GOSRC
fi

GO=$HOME/go/bin/go

if [ ! -f "go.mod" ];then
    $GO mod init goLogAnalyzer
fi
$GO mod tidy

echo "Installing goLogAnalyzer"
echo go build -o logan main.go
$GO build -o logan main.go
echo "sudo cp logan /usr/local/bin/logan"
sudo cp logan /usr/local/bin/logan
echo "OK!"
echo ""
