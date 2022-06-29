package tests

import (
	bunyan "github.com/Dewberry/paul-bunyan"
	"github.com/udhos/equalfile"
)

func sameFiles(f1, f2 string) bool {
	cmp := equalfile.New(nil, equalfile.Options{})
	equal, err := cmp.CompareFile(f1, f2)
	if err != nil {
		bunyan.Fatal(err)
	}

	return equal
}
