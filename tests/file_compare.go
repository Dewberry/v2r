package tests

import (
	"log"

	"github.com/udhos/equalfile"
)

func sameFiles(f1, f2 string) bool {
	cmp := equalfile.New(nil, equalfile.Options{}) // compare using single mode
	equal, err := cmp.CompareFile(f1, f2)
	if err != nil {
		log.Fatal(err)
	}

	return equal
}
