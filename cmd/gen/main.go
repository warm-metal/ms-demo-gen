package main

import (
	"flag"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/warm-metal/ms-demo-gen.git/pkg/dag"
	"gonum.org/v1/gonum/graph/encoding/dot"
)

var (
	numServices      = flag.Int("services", 1, "Number of services in the demo")
	maxCaller        = flag.Int("max-caller", 1, "Maxinum number of callers for each service except the root service")
	maxCallee        = flag.Int("max-callee", 1, "Maxinum number of callees for each service except leaf services")
	longestCallChain = flag.Int("longest-call-chain", -1, "Number of services in the longest call chain. -1 means not limit")
)

func main() {
	flag.Parse()
	g := dag.New(&dag.Options{
		NumberVertices: *numServices,
		InDegreeRange:  [2]int{1, *maxCaller},
		OutDegreeRange: [2]int{0, *maxCallee},
		LongestWalk:    *longestCallChain,
	})

	dotBin, err := dot.Marshal(g, "", "", "")
	if err != nil {
		panic(err)
	}

	err = ioutil.WriteFile(filepath.Join(os.TempDir(), "ws-demo.dot"), dotBin, 0755)
	if err != nil {
		panic(err)
	}
}
