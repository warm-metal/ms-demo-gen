package dag

import (
	"io"
	"math/rand"
	"os"
	"strings"

	rands "github.com/xyproto/randomstring"
)

type randNameGen struct {
	words []string
}

func (g randNameGen) Name() string {
	if len(g.words) == 0 {
		return rands.HumanFriendlyEnglishString(10)
	}

	return g.words[rand.Int()%len(g.words)]
}

func readAvailableDictionary() ([]string, error) {
	file, err := os.Open("/usr/share/dict/words")
	if err != nil {
		return nil, err
	}

	bytes, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	return strings.Split(string(bytes), "\n"), nil
}

func createNameGen() *randNameGen {
	words, _ := readAvailableDictionary()
	return &randNameGen{words}
}
