echo "Installling goLogAnalyzer"
go get "github.com/damnever/bitarray"
go get "github.com/pkg/errors"
go get "github.com/go-ini/ini"
go get "golang.org/x/text/unicode/norm"
echo go build -o logan main.go
go build -o logan main.go
echo "installing loganal"
sudo cp logan /usr/local/bin/logan
