#!/bin/bash
set -e

echo "Installing goLogAnalyzer"
echo go build -o logan main.go
go build -o logan main.go
echo "sudo cp logan /usr/local/bin/logan"
sudo cp logan /usr/local/bin/logan
echo "OK!"
echo ""
