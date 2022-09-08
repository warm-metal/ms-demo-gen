package dag_test

import (
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/warm-metal/ms-demo-gen.git/pkg/dag"
	rands "github.com/xyproto/randomstring"
)

func TestMain(m *testing.M) {
	rands.Seed()
	rand.Seed(time.Now().UnixNano())
	os.Exit(m.Run())
}

func TestDagGenerationWithoutVersion(t *testing.T) {
	g := dag.New(&dag.Options{
		NumberVertices: 10,
		InDegreeRange:  [2]int{1, 2},
		OutDegreeRange: [2]int{0, 3},
		NumberVersionsRange: [2]int{1, 1},
		LongestWalk:    -1,
	})
	if g.Node(1) == nil {
		t.Log(g)
		t.FailNow()
	}
}

func TestDagGenerationWithVersion(t *testing.T) {
	g := dag.New(&dag.Options{
		NumberVertices: 10,
		InDegreeRange:  [2]int{1, 2},
		OutDegreeRange: [2]int{0, 3},
		NumberVersionsRange: [2]int{1, 3},
		LongestWalk:    -1,
	})
	if g.Node(1) == nil {
		t.Log(g)
		t.FailNow()
	}
}
