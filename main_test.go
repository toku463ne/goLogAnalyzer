package main

import (
	"os"
	"testing"
)

func Test_main(t *testing.T) {
	os.Args = []string{"test", "cleanup", "-c", "test.ini"}
	main()
	//os.Args = []string{"test", "run", "-c", "test.ini", "-v"}
	os.Args = []string{"test", "run", "-c", "test.ini"}
	main()
}

func Test_main2(t *testing.T) {
	os.Args = []string{"test", "run", "-p", "c:\\Users\\kot\\loganal\\realtest\\test.log*"}
	main()
}

func Test_main3(t *testing.T) {
	os.Args = []string{"test", "cleanup", "-c", "test2.ini"}
	main()
	os.Args = []string{"test", "run", "-c", "test2.ini"}
	main()
}
