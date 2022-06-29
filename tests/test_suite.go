package tests

import bunyan "github.com/Dewberry/paul-bunyan"

func TestSuite() {

	passedIDW := testIDW()
	passedGPKGRead := testGPKGRead()
	passedCleaner := testCleaner()
	bunyan.Infof("Passed All IDW Tests?      %v", passedIDW)
	bunyan.Infof("GPKG read correct?         %v", passedGPKGRead)
	bunyan.Infof("Passed All Cleaner Tests?  %v", passedCleaner)
}
