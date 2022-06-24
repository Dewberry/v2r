package tests

import bunyan "github.com/Dewberry/paul-bunyan"

func TestSuite() {

	passedIDW := testIDW()
	passedCleaner := testCleaner()
	bunyan.Infof("Passed All IDW Tests? \t%v", passedIDW)
	bunyan.Infof("Passed All Cleaner Tests?\t%v", passedCleaner)
}
