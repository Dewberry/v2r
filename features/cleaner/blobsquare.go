package cleaner

import (
	"app/tools"
)

type blob struct {
	Elements      []tools.OrderedPair
	NumFixed      int
	ThresholdSize int
	IsWater       byte
}

type square struct {
	IsWater   byte
	Searched  bool
	Finalized bool
	IsChanged bool
}

//Helper methods for blob and square
func getSquarePair(areaMap *[][]square, loc tools.OrderedPair) square {
	return (*areaMap)[loc.R][loc.C]
}

func getSquareRC(areaMap *[][]square, r int, c int) square {
	return (*areaMap)[r][c]
}

func setSearched(areaMap *[][]square, loc tools.OrderedPair, searched bool) {
	(*areaMap)[loc.R][loc.C].Searched = searched
}

func setFinalized(areaMap *[][]square, loc tools.OrderedPair, finalized bool) {
	(*areaMap)[loc.R][loc.C].Finalized = finalized
}

func setWet(areaMap *[][]square, loc tools.OrderedPair, wet byte) {
	(*areaMap)[loc.R][loc.C].IsWater = wet
}

func setChanged(areaMap *[][]square, loc tools.OrderedPair, changed bool) {
	(*areaMap)[loc.R][loc.C].IsChanged = changed
}

func sameBlob(areaMap *[][]square, loc1 tools.OrderedPair, loc2 tools.OrderedPair) bool {
	return getSquarePair(areaMap, loc1).IsWater == getSquarePair(areaMap, loc2).IsWater
}

// length of Elements list unless fixed > 0
// then return fixed
func getNumElements(b *blob) int {
	if b.NumFixed > 0 {
		return b.NumFixed
	}
	return len(b.Elements)
}

func isBigBlob(b *blob) bool {
	return b.NumFixed > 0
}

func updateMapFromBlob(areaMap *[][]square, b *blob, wet byte, finalized bool) {
	for _, loc := range b.Elements {
		b.NumFixed++
		setWet(areaMap, loc, wet)
		setFinalized(areaMap, loc, finalized)
		setChanged(areaMap, loc, wet != b.IsWater)
	}
	b.Elements = nil
}

func beenSearched(areaMap *[][]square, loc tools.OrderedPair) bool {
	return getSquarePair(areaMap, loc).Searched
}

//also no data values represented as 255
func isInBounds(areaMap *[][]square, loc tools.OrderedPair) bool {
	return loc.R >= 0 && loc.R < len(*areaMap) && loc.C >= 0 && loc.C < len((*areaMap)[0]) && getSquarePair(areaMap, loc).IsWater != 255
}

func growBlob(areaMap *[][]square, b *blob, loc tools.OrderedPair) {
	if b.NumFixed > 0 {
		b.NumFixed++
		setWet(areaMap, loc, b.IsWater)
		setFinalized(areaMap, loc, true)
	} else {
		b.Elements = append(b.Elements, loc)
		if len(b.Elements) >= b.ThresholdSize {
			updateMapFromBlob(areaMap, b, b.IsWater, true)
		}
	}

}
