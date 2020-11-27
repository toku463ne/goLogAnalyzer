package main

import (
	"os"
	"testing"
)

func Test_main(t *testing.T) {
	os.Args = []string{"test", "destroy", "-c", "test.ini"}
	main()
	//os.Args = []string{"test", "run", "-c", "test.ini", "-v"}
	os.Args = []string{"test", "run", "-c", "test.ini"}
	main()
}
