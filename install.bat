echo "Installling goLogAnalyzer"
go get "github.com/damnever/bitarray"
go get "github.com/pkg/errors"
go get "github.com/go-ini/ini"
go get "golang.org/x/text/unicode/norm"
echo go build -o logan.exe main.go
go build -o logan.exe main.go
