package processing

import (
	"bufio"
	"os"
	"testing"

	bunyan "github.com/Dewberry/paul-bunyan"
	"github.com/dewberry/v2r/tools"
)

func isListCorrect(listPoints []tools.Point) bool {
	correct := []tools.Point{
		tools.MakePoint(1.203154304903684e+07, 3.953694544257536e+06, 0),
		tools.MakePoint(1.2039094668343753e+07, 3.9424537521008854e+06, 0),
		tools.MakePoint(1.2052295156865465e+07, 3.93148004717589e+06, 0),
		tools.MakePoint(1.2076264960092723e+07, 3.917394837044359e+06, 0),
		tools.MakePoint(1.209133966962186e+07, 3.899012514405147e+06, 0),
		tools.MakePoint(1.2020401943588393e+07, 3.9463317049700455e+06, 2.5),
		tools.MakePoint(1.2025601603362886e+07, 3.9380775540176313e+06, 2.9),
		tools.MakePoint(1.203495724450095e+07, 3.92078279952575e+06, 7),
		tools.MakePoint(1.2055886381444821e+07, 3.898699015764322e+06, 6.3),
		tools.MakePoint(1.207451364898982e+07, 3.885548111378168e+06, 2),
		tools.MakePoint(1.2049260510355683e+07, 3.952539493517972e+06, 3.1),
		tools.MakePoint(1.205195120503224e+07, 3.947886314968573e+06, 2.8),
		tools.MakePoint(1.2096825067172093e+07, 3.9268244968121317e+06, 3.9),
		tools.MakePoint(1.2080323349286443e+07, 3.9412021225619004e+06, 4.6)}

	if len(correct) != len(listPoints) {
		bunyan.Debug(correct)
		bunyan.Debug(listPoints)

		return false
	}

	for i, p := range correct {
		if p != listPoints[i] {
			bunyan.Debug(correct)
			bunyan.Debug(listPoints)

			return false
		}
	}

	return true
}

func correctSRS(proj string, folder string) bool {
	filename := folder + "gpkg_proj.txt"
	f, err := os.Open(filename)
	if err != nil {
		bunyan.Errorf("srs, incorrect filename: %v", filename)
		return false
	}
	sc := bufio.NewScanner(f)
	defer f.Close()

	sc.Scan()
	correctProj := sc.Text()
	return correctProj == proj
}

func correctXInfo(xInfo tools.Info) bool {
	return xInfo.Min == 1.2020401943588393e+07 && xInfo.Max == 1.2096825067172093e+07 && xInfo.Step == 1.0
}

func correctYInfo(yInfo tools.Info) bool {
	return yInfo.Min == 3.885548111378168e+06 && yInfo.Max == 3.953694544257536e+06 && yInfo.Step == 2.0
}

func TestGPKGRead(t *testing.T) {
	bunyan.Info("____________________________")
	bunyan.Info("read gpkg")

	dataFolder := "../../features/idw/idw_test_files/"
	filepath := dataFolder + "input.gpkg"
	bunyan.Debug("filepath ", filepath)

	listPoints, proj, xInfo, yInfo, err := ReadGeoPackage(filepath, "wsels", "elevation", 1.0, 2.0)
	if err != nil {
		t.Error(err)
	}
	bunyan.Debug("listPoints ", listPoints)
	bunyan.Debug("proj", proj)
	bunyan.Debug("xInfo", xInfo)
	bunyan.Debug("yInfo", yInfo)

	t.Run("data", func(t *testing.T) {
		if !isListCorrect(listPoints) {
			t.Error("points generated from geopackage incorrect")
		}
	})
	t.Run("proj", func(t *testing.T) {
		if !correctSRS(proj, dataFolder) {
			t.Error("srs generated from geopackage incorrect")
		}
	})
	t.Run("xInfo", func(t *testing.T) {
		if !correctXInfo(xInfo) {
			t.Error("xInfo incorrect")
		}
	})
	t.Run("yInfo", func(t *testing.T) {
		if !correctYInfo(yInfo) {
			t.Error("yInfo incorrect")
		}
	})

	bunyan.Info("____________________________")

}
