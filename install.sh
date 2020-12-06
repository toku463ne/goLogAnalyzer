echo "Installling goLogAnalyzer"
echo "Installing dependencies"
go get "github.com/damnever/bitarray"
go get "github.com/pkg/errors"
go get "github.com/go-ini/ini"
go get "golang.org/x/text/unicode/norm"
echo go build -o loganal main.go
go build -o loganal main.go
echo "installing loganal"
sudo cp loganal /usr/local/bin/loganal
