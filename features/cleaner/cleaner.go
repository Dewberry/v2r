package cleaner

import (
	"math"

	"github.com/dewberry/v2r/tools"
	"github.com/dewberry/v2r/tools/processing"

	bunyan "github.com/Dewberry/paul-bunyan"
)

func cleanAreaMap(areaMap *[][]square, tolerance map[byte]int, pixelArea float64, adjType int, ICP innerChunkPartition) cleanerStats {
	statsTotal := cleanerStats{}

	for r := 0; r < ICP.REnd; r++ {
		for c := 0; c < ICP.CEnd; c++ {
			sq := getSquareRC(areaMap, r, c)
			if sq.IsWater == 255 {
				continue
			}
			if isMinimumStatus(areaMap, tools.MakePair(r, c), Searched) {
				loc := tools.MakePair(r, c)
				if getStatus(areaMap, loc) == Changed && isInPartiion(ICP, loc) {
					switch sq.IsWater {
					case byte(0): // from water to land
						statsTotal.VoidArea++
						continue

					case byte(1): //land to water
						statsTotal.IslandArea++
						continue
					}
				}
				continue
			}
			setStatus(areaMap, tools.MakePair(r, c), Searched)
			blob, skip := searchBlob(areaMap, tools.MakePair(r, c), adjType, tolerance[getSquareRC(areaMap, r, c).IsWater], getSquareRC(areaMap, r, c).IsWater)
			if skip {
				continue
			}

			if !isBigBlob(&blob) {
				switch blob.IsWater {
				case byte(0): // update island to water
					updateMapFromBlob(areaMap, &blob, 1, Changed)
					if isInPartiion(ICP, tools.MakePair(r, c)) {
						statsTotal.Islands++
						statsTotal.IslandArea++
					}

				case byte(1): // update island to water
					updateMapFromBlob(areaMap, &blob, 0, Changed)
					if isInPartiion(ICP, tools.MakePair(r, c)) {
						statsTotal.Voids++
						statsTotal.VoidArea++
					}
				}
			}

		}
	}
	return statsTotal
}

func searchBlob(areaMap *[][]square, loc tools.OrderedPair, adjType int, thresholdsize int, wet byte) (blob, bool) {
	blob := blob{make([]tools.OrderedPair, 0, thresholdsize), 0, thresholdsize, wet}
	blob.Elements = append(blob.Elements, loc)
	searchStack := []tools.OrderedPair{loc}

	finalized := false
	for len(searchStack) > 0 {
		n := len(searchStack) - 1
		searchLoc := searchStack[n]
		searchStack = searchStack[:n]

		adjacents, shouldSkip := getSimilarSurrounding(areaMap, searchLoc, adjType)
		finalized = finalized || shouldSkip

		for _, adjLoc := range adjacents {
			growBlob(areaMap, &blob, adjLoc)
			searchStack = append(searchStack, adjLoc)
		}
	}
	finalized = finalized || isBigBlob(&blob)
	return blob, finalized
}

// returns a list of adjacent locations and whether to skip over blob (reached a finalized location)
// adjacent locations must be unsearched and of similar type (wet/nonwet)
// adjacent locations specified by adjType
func getSimilarSurrounding(areaMap *[][]square, loc tools.OrderedPair, adjType int) ([]tools.OrderedPair, bool) {
	vectors := AdjacentVectors(adjType)
	directions := [2]int{-1, 1}

	skip := false
	var validSurrounding []tools.OrderedPair
	for _, vec := range vectors {
		for _, dir := range directions {
			adjLoc := tools.MakePair(loc.R+dir*vec.R, loc.C+dir*vec.C)
			if isInBounds(areaMap, adjLoc) && sameBlob(areaMap, loc, adjLoc) {
				if isMinimumStatus(areaMap, adjLoc, Changed) {
					skip = true
				}
				if !isMinimumStatus(areaMap, adjLoc, Searched) {
					validSurrounding = append(validSurrounding, adjLoc)
					setStatus(areaMap, adjLoc, Searched)
				}
			}
		}
	}
	return validSurrounding, skip
}

func CleanFull(filepath string, outfile string, toleranceIsland float64, toleranceVoid float64, adjType int) error {
	areaMap, gdal, err := readFile(filepath)
	bunyan.Infof("[%v, %v]", len(areaMap), len(areaMap[0]))

	if err != nil {
		bunyan.Fatal(err)
	}
	areaSize := math.Abs(gdal.XCell * gdal.YCell)
	tolerance := map[byte]int{0: int(toleranceIsland / areaSize), 1: int(toleranceVoid / areaSize)}

	ICP := innerChunkPartition{0, len(areaMap), 0, len(areaMap[0])}
	summary := cleanAreaMap(&areaMap, tolerance, areaSize, adjType, ICP)
	printStats(summary, areaSize)
	return processing.WriteTif(flattenAreaMap(areaMap), gdal, outfile, tools.MakePair(0, 0), tools.MakePair(len(areaMap), len(areaMap[0])), tools.MakePair(len(areaMap), len(areaMap[0])), true)

}
