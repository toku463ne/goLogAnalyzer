#!/bin/bash
set -e

echo "Installing goLogAnalyzer"
go get "github.com/damnever/bitarray"
go get "github.com/pkg/errors"
go get "github.com/go-ini/ini"
go get "golang.org/x/text/unicode/norm"
go get "github.com/juju/fslock"
echo go build -o logan main.go
go build -o logan main.go
echo "sudo cp logan /usr/local/bin/logan"
sudo cp logan /usr/local/bin/logan
echo "OK!"
echo ""
echo ""
logan -help
echo ""
echo ""
logan rar -help
echo ""
echo ""
logan clean -help
echo ""
echo ""
logan stats -help
echo ""
echo ""
logan topN -help
echo ""
echo ""
logan frq -help
