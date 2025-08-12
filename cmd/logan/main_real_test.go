package main

import (
	"os"
	"testing"
)

func Test_real_sbc_gateway(t *testing.T) {
	config := "/home/ubuntu/tests/sbc/gateway/sbc_gateway.yml.j2"
	//os.Args = []string{"logan", "clean", "-c", config, "-silent"}
	//main()

	os.Args = []string{"logan", "patterns", "-c", config}
	main()
}

func Test_real_sophos(t *testing.T) {
	config := "/home/ubuntu/tests/sophos/SOPHOS-01.yml"
	//os.Args = []string{"logan", "clean", "-c", config, "-silent"}
	//main()

	os.Args = []string{"logan", "history", "-c", config, "-o", "/tmp/out2", "-asc", "-N", "10"}
	main()
}
