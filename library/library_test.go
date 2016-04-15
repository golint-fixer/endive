package library

import (
	"fmt"
	"testing"

	cfg "github.com/barsanuphe/endive/config"
)

func TestLibrarySearch(t *testing.T) {
	c := cfg.Config{}
	k := cfg.KnownHashes{}
	ldb := LibraryDB{DatabaseFile: "test/endive.json"}
	l := Library{c, k, ldb}

	err := l.Load()
	if err != nil {
		t.Errorf("Error loading epubs from database: " + err.Error())
	}
	results, err := l.RunQuery("language:fr")
	if err != nil {
		t.Errorf("Error runnig query: " + err.Error())
	}
	fmt.Println(results)
}
