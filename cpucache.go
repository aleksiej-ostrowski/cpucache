/*

# ------------------------------ #
#                                #
#  version 0.0.1                 #
#                                #
#  Aleksiej Ostrowski, 2023      #
#                                #
#  https://aleksiej.com          #
#                                #
# ------------------------------ #

*/

package main

import (
	"encoding/json"
	"fmt"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	"io/ioutil"
	"math"
	"math/rand"
	"runtime"
	"runtime/debug"
	"time"
)

type MyOut struct {
	Data    [][]time.Duration
	Labels  []string
	X       []int
	Xlabel  string
	Xfilter string
	Ylabel  string
	Title   string
}

// https://stackoverflow.com/questions/39868029/how-to-generate-a-sequence-of-numbers
func makeRange(min, max int) []int {
	a := make([]int, max-min+1)
	for i := range a {
		a[i] = min + i
	}
	return a
}

func main() {

	// https://stackoverflow.com/questions/38972003/how-to-stop-the-golang-gc-and-trigger-it-manually
	// DISABLE the GC

	debug.SetGCPercent(-1)
	debug.SetMemoryLimit(math.MaxInt64)

	rand.Seed(time.Now().UTC().UnixNano())

	ITER := 5

	SIZE_IN_MB := makeRange(1, 20)

	MET := []string{"COLD L-cache", "HOT L-cache"}

	DATA_TIMES := make([][]time.Duration, len(MET))

	for i := range DATA_TIMES {
		DATA_TIMES[i] = make([]time.Duration, len(SIZE_IN_MB))
	}

	for_clear_cash := make([]int, (1<<20)*20)

	for iter := 0; iter < ITER; iter++ {
		for idx_MB, MB := range SIZE_IN_MB {

			size := (1 << 20) * MB

			{
				// https://stackoverflow.com/questions/3446138/how-to-clear-cpu-l1-and-l2-cache
				// This is an attempt to RESET the CPU CACHE

				for i := 0; i < len(for_clear_cash); i++ {
					for_clear_cash[i] = int(math.Sin(float64(rand.Intn(math.MaxInt32)))*0.5) + 5
				}

				_ = for_clear_cash[len(for_clear_cash)-1]

			}

			data := make([]int, size)

			start := time.Now()
			// MAIN code
			data[0] = 17
			for i := 1; i < size; i++ {
				data[i] = data[i-1] + 3
			}
			// end of the MAIN code
			DATA_TIMES[0][idx_MB] += time.Since(start)

			_ = data[size-1]

			for i := 0; i < size; i++ {
				data[i] = 0
			}

			// temp := make([]int, size)
			// copy(data, temp)

			start = time.Now()
			// TWIN of the MAIN code
			data[0] = 17
			for i := 1; i < size; i++ {
				data[i] = data[i-1] + 3
			}
			// end of the TWIN

			DATA_TIMES[1][idx_MB] += time.Since(start)

			_ = data[size-1]
		}
	}

	for idx_MB := range SIZE_IN_MB {
		for idx_MET := range MET {
			DATA_TIMES[idx_MET][idx_MB] /= time.Duration(ITER)
			DATA_TIMES[idx_MET][idx_MB] /= time.Duration(ITER)
		}
	}

	cpuStat, _ := cpu.Info()
	vmStat, _ := mem.VirtualMemory()

	cpu := cpuStat[0].ModelName
	// cache := cpuStat[0].CacheSize
	ram := vmStat.Total / 1024 / 1024 / 1024

	info := fmt.Sprintf("%s, logical CPUs: %d, %d GB RAM", cpu, runtime.NumCPU(), ram)

	mydata := &MyOut{
		Data:    DATA_TIMES,
		Labels:  MET,
		X:       SIZE_IN_MB,
		Xlabel:  "MB",
		Xfilter: "(x > 20_000_000)", // "(x > 10_000) and not((x < 60_000) and (y < 5.))",
		Ylabel:  "Time",
		Title:   "Compare, " + info,
	}

	// fmt.Println(mydata)

	b, err := json.Marshal(mydata)

	if err != nil {
		fmt.Println(err)
	}

	RESULT := "./result.xml"

	_ = ioutil.WriteFile(RESULT, b, 0644)

	fmt.Println("ok")
}
