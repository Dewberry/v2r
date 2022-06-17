package tools

type Blob struct {
	Elements      []OrderedPair
	NumFixed      int
	ThresholdSize int
	IsWater       byte
}

type Square struct {
	IsWater   byte
	Searched  bool
	Finalized bool
}

//Helper methods for blob and square
func getSquarePair(areaMap *[][]Square, loc OrderedPair) Square {
	return (*areaMap)[loc.R][loc.C]
}

func getSquareRC(areaMap *[][]Square, r int, c int) Square {
	return (*areaMap)[r][c]
}

func setSearched(areaMap *[][]Square, loc OrderedPair, searched bool) {
	(*areaMap)[loc.R][loc.C].Searched = searched
}

func setFinalized(areaMap *[][]Square, loc OrderedPair, finalized bool) {
	(*areaMap)[loc.R][loc.C].Finalized = finalized
}

func setWet(areaMap *[][]Square, loc OrderedPair, wet byte) {
	(*areaMap)[loc.R][loc.C].IsWater = wet
}

func sameBlob(areaMap *[][]Square, loc1 OrderedPair, loc2 OrderedPair) bool {
	return getSquarePair(areaMap, loc1).IsWater == getSquarePair(areaMap, loc2).IsWater
}

// length of Elements list unless fixed > 0
// then return fixed
func GetNumElements(blob *Blob) int {
	if blob.NumFixed > 0 {
		return blob.NumFixed
	}
	return len(blob.Elements)
}

func BigBlob(blob *Blob) bool {
	return blob.NumFixed > 0
}

func updateMapFromBlob(areaMap *[][]Square, blob *Blob, wet byte, finalized bool) {
	for _, loc := range blob.Elements {
		blob.NumFixed++
		setWet(areaMap, loc, wet)
		setFinalized(areaMap, loc, finalized)
	}
	blob.Elements = nil
}

func searchedLoc(areaMap *[][]Square, loc OrderedPair) bool {
	return getSquarePair(areaMap, loc).Searched
}

//also no data values represented as 255
func inBounds(areaMap *[][]Square, loc OrderedPair) bool {
	return loc.R >= 0 && loc.R < len(*areaMap) && loc.C >= 0 && loc.C < len((*areaMap)[0]) && getSquarePair(areaMap, loc).IsWater != 255
}

func growBlob(areaMap *[][]Square, blob *Blob, loc OrderedPair) {
	if blob.NumFixed > 0 {
		blob.NumFixed++
		setWet(areaMap, loc, blob.IsWater)
		setFinalized(areaMap, loc, true)
	} else {
		blob.Elements = append(blob.Elements, loc)
		if len(blob.Elements) >= blob.ThresholdSize {
			updateMapFromBlob(areaMap, blob, blob.IsWater, true)
		}
	}

}
