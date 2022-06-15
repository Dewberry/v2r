package tools

type Blob struct {
	Elements []OrderedPair
	IsWater  byte
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

func sameBlob(areaMap *[][]Square, loc1 OrderedPair, loc2 OrderedPair) bool {
	return getSquarePair(areaMap, loc1).IsWater == getSquarePair(areaMap, loc2).IsWater
}

func updateMapFromBlob(areaMap *[][]Square, blob *Blob, wet byte) {
	for _, loc := range blob.Elements {
		(*areaMap)[loc.R][loc.C].IsWater = wet
		(*areaMap)[loc.R][loc.C].Finalized = true
	}
}

func searchedLoc(areaMap *[][]Square, loc OrderedPair) bool {
	return getSquarePair(areaMap, loc).Searched
}

//also no data values represented as 255
func inBounds(areaMap *[][]Square, loc OrderedPair) bool {
	return loc.R >= 0 && loc.R < len(*areaMap) && loc.C >= 0 && loc.C < len((*areaMap)[0]) && getSquarePair(areaMap, loc).IsWater != 255
}

func growBlob(blob *Blob, loc OrderedPair) {
	blob.Elements = append(blob.Elements, loc)
}
