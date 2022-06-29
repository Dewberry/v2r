package cleaner

import (
	"github.com/dewberry/v2r/tools"
)

type blob struct {
	Elements      []tools.OrderedPair
	NumFixed      int
	ThresholdSize int
	IsWater       byte
}

type square struct {
	IsWater byte
	Status  status
}

type status byte

const (
	Unsearched status = iota
	Searched
	Changed
	Finalized
)

//Helper methods for blob and square
func getSquarePair(areaMap *[][]square, loc tools.OrderedPair) square {
	return (*areaMap)[loc.R][loc.C]
}

func getSquareRC(areaMap *[][]square, r int, c int) square {
	return (*areaMap)[r][c]
}

func setStatus(areaMap *[][]square, loc tools.OrderedPair, s status) {
	(*areaMap)[loc.R][loc.C].Status = s
}

func getStatus(areaMap *[][]square, loc tools.OrderedPair) status {
	return (*areaMap)[loc.R][loc.C].Status
}

func setWet(areaMap *[][]square, loc tools.OrderedPair, wet byte) {
	(*areaMap)[loc.R][loc.C].IsWater = wet
}

func sameBlob(areaMap *[][]square, loc1 tools.OrderedPair, loc2 tools.OrderedPair) bool {
	return getSquarePair(areaMap, loc1).IsWater == getSquarePair(areaMap, loc2).IsWater
}

func isBigBlob(b *blob) bool {
	return b.NumFixed > 0
}

func updateMapFromBlob(areaMap *[][]square, b *blob, wet byte, s status) {
	for _, loc := range b.Elements {
		b.NumFixed++
		setWet(areaMap, loc, wet)

		setStatus(areaMap, loc, s)
		// setChanged(areaMap, loc, wet != b.IsWater)
	}
	b.Elements = nil
}

func isMinimumStatus(areaMap *[][]square, loc tools.OrderedPair, s status) bool {
	return getSquarePair(areaMap, loc).Status >= s
}

//also no data values represented as 255
func isInBounds(areaMap *[][]square, loc tools.OrderedPair) bool {
	return loc.R >= 0 && loc.R < len(*areaMap) && loc.C >= 0 && loc.C < len((*areaMap)[0]) && getSquarePair(areaMap, loc).IsWater != 255
}

func growBlob(areaMap *[][]square, b *blob, loc tools.OrderedPair) {
	if b.NumFixed > 0 {
		b.NumFixed++
		setWet(areaMap, loc, b.IsWater)
		setStatus(areaMap, loc, Finalized)
	} else {
		b.Elements = append(b.Elements, loc)
		if len(b.Elements) >= b.ThresholdSize {
			updateMapFromBlob(areaMap, b, b.IsWater, Finalized)
		}
	}

}
