package main

import (
	"os"
	"testing"
)

func Test_main(t *testing.T) {
	//loganal rar [-f LOGPATH] [-d DATADIR] [-g GAPVALUE] [-v] [-s SEARCH_KEYS] [-x EXCLUDE_KEYS]
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

func Test_main4(t *testing.T) {
	os.Args = []string{"test", "frq",
		"-f", "c:\\Users\\kot\\loganal\\realtest\\test.log",
		"-m", "100", "-x", "error"}
	main()
}

func Test_main5(t *testing.T) {
	os.Args = []string{"test", "cleanup", "-c", "test3.ini"}
	main()
	os.Args = []string{"test", "run", "-c", "test3.ini"}
	main()
}

func Test_main6(t *testing.T) {
	os.Args = []string{"test", "cleanup", "-c", "test4.ini"}
	main()
	os.Args = []string{"test", "run", "-c", "test4.ini"}
	main()
}
