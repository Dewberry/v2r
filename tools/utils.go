package tools

import (
	"fmt"
	"math"
	"strings"
	"time"

	bunyan "github.com/Dewberry/paul-bunyan"
	"github.com/natefinch/lumberjack"
	"github.com/pbnjay/memory"
)

//Start Basic Utilites
func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func Transpose(a [][]float64) [][]float64 {
	m := len(a[0])
	n := len(a)

	b := make([][]float64, m)
	for i := 0; i < m; i++ {
		b[i] = make([]float64, n)
		for j := 0; j < n; j++ {
			b[i][j] = a[j][i]
		}
	}

	return b
}

//End Basic Utilities

//Start Point/OrderedPair/Coord and helper functions
type Point struct {
	X      float64
	Y      float64
	Weight float64
}

type OrderedPair struct {
	R int
	C int
}

type Coord struct {
	P    Point
	Pair OrderedPair
}

func MakePoint(x, y, w float64) Point {
	return Point{x, y, w}
}

func MakePair(r int, c int) OrderedPair {
	return OrderedPair{r, c}
}

func PointToPair(p Point, xInfo Info, yInfo Info) OrderedPair {
	return OrderedPair{int((p.Y - yInfo.Min) / yInfo.Step), int((p.X - xInfo.Min) / xInfo.Step)}
}

func PairToPoint(pair OrderedPair, xInfo Info, yInfo Info) Point {
	return RCToPoint(pair.R, pair.C, xInfo, yInfo)
}

func RCToPoint(r int, c int, xInfo Info, yInfo Info) Point {
	px := xInfo.Min + float64(c)*xInfo.Step
	py := yInfo.Min + float64(r)*yInfo.Step
	return (Point{px, py, 0})
}

func RCToPair(r, c int) OrderedPair {
	return OrderedPair{r, c}
}

func PairToRC(pair OrderedPair) (int, int) {
	return pair.R, pair.C
}

func createCoordPoint(p Point, xInfo Info, yInfo Info, ch chan Coord) {
	pair := PointToPair(p, xInfo, yInfo)
	ch <- Coord{Point{p.X, p.Y, p.Weight}, pair}
}

func MakeCoordSpace(listPoints *[]Point, xInfo Info, yInfo Info) map[OrderedPair]Point {
	seen := map[OrderedPair]Point{}

	channel := make(chan Coord, len(*listPoints))
	for _, p := range *listPoints {
		go createCoordPoint(p, xInfo, yInfo, channel)
	}

	for i := 0; i < len(*listPoints); i++ {
		dataPoint := <-channel

		pair := dataPoint.Pair
		p := dataPoint.P
		elev, exists := seen[pair]
		if exists {
			newElev := (p.Weight + elev.Weight) / 2
			bunyan.Debugf("%v already exists | old elev: %v | this elev%v | ave elev: %v", pair, elev, p.Weight, newElev)
			p.Weight = newElev
		}
		seen[pair] = p
	}
	return seen
}

//End Point/OrderedPair/Coord and helper functions

// Info Structure
// For Conversions to and from Euclidean Space
type Info struct {
	Min  float64 `default:"math.Inf(1)"`
	Max  float64 `default:"math.Inf(-1)"`
	Step float64 `default:"1.0"`
}

func MakeInfo() Info {
	return Info{math.Inf(1), math.Inf(-1), 1.0}
}

func GetDimensions(xInfo Info, yInfo Info) (int, int) {
	cols := int((1 + xInfo.Max - xInfo.Min) / xInfo.Step) // (1 + max - min)/step
	rows := int((1 + yInfo.Max - yInfo.Min) / yInfo.Step)
	return rows, cols
}

// Start Math Functions
func RoundUp(num, denom int) int {
	if num%denom == 0 {
		return num / denom
	}
	return 1 + num/denom
}

func GetChunkBlock(row, chunkR, chunkC int, xInfo Info, yInfo Info) (int, int) {
	_, numCols := GetDimensions(xInfo, yInfo)
	numChunksOnCol := int(math.Ceil(float64(numCols) / float64(chunkC)))

	start := row / chunkR * numChunksOnCol
	return start, start + numChunksOnCol

}

func euclidDist(p1, p2 Point) float64 {
	total := math.Pow(p1.X-p2.X, 2) + math.Pow(p1.Y-p2.Y, 2)
	return math.Pow(total, .5)
}

// p0 is grid location, p is weighted point to compare to
func PartialWeight(p0, p Point, exp float64) float64 {
	return p.Weight / (math.Pow(euclidDist(p, p0), exp))
}

func DistExp(p0, p Point, exp float64) float64 {
	return math.Pow(euclidDist(p, p0), -exp)
}

func CalculateWeight(cell Point, data *[]Point, exp float64) float64 {
	total := 0.0
	for _, p := range *data {
		total += p.Weight / (math.Pow(euclidDist(cell, p), exp))
	}
	return math.Pow(total, .5)
}

// End Math Functions

//Memory Management
func ChannelSize(appxSubprocess uint64, appxOverhead uint64) int {
	bunyan.Debugf("Total system memory: %d bytes", memory.TotalMemory())
	bunyan.Infof("Free memory: %d bytes", memory.FreeMemory())

	bunyan.Debugf("Appx Memory/subprocess: %d bytes", appxSubprocess)
	bunyan.Debugf("Appx Overhead: %v bytes", appxOverhead)

	calculated := int((memory.FreeMemory()*8/10 - appxOverhead) / appxSubprocess)
	bunyan.Infof("using 80%% of free memory")
	bunyan.Debugf("allocated channel size: %v ", calculated)

	return Max(1, calculated)
}

func ChangeExtension(filename string, ext string) string {
	period := strings.LastIndex(filename, ".")
	return fmt.Sprintf("%s%s", filename[:period], ext)
}

func SetLogging() {
	logger := bunyan.New()
	date := time.Now().Format("2006-02-01_15:04:05") // YYYY-MM-DD
	file := fmt.Sprintf("logs/%s.txt", date)
	logger.SetOutput(&lumberjack.Logger{
		Filename:   file,
		MaxSize:    1, // megabytes
		MaxBackups: 100,
		MaxAge:     90,   //days
		Compress:   true, // disabled by default
	})
}
