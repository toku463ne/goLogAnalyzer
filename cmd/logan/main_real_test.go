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

func Test_real_sbc_gateway2(t *testing.T) {
	config := "/data/Documents/202508_sug5k_noans/loganconfs/gateway.yaml"
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

func Test_real_flip(t *testing.T) {
	config := "/data/Documents/202510_flip3/new/loganal.yml"
	//os.Args = []string{"logan", "clean", "-c", config, "-silent"}
	//main()
	os.Args = []string{"logan", "groups", "-c", config}
	main()

	os.Args = []string{"logan", "groups", "-c", config, "-N", "100"}
	main()

}
