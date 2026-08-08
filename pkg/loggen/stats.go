package loggen

import (
	"cmp"
	"fmt"
	"slices"
	"strconv"
	"time"
)

type Stats struct {
	Worker   string
	Count    int64
	Duration time.Duration
	Bytes    uint64
}

type MetaStats struct {
	avgBytes     uint64
	avgDuration  time.Duration
	avgEmitted   int
	avgEps       float64
	maxCol2      int
	maxCol4      int
	maxCol5      int
	totalBytes   uint64
	totalEmitted int64
	workerCount  int
}

func columnWidthMin(value any, min int) int {
	var n string
	switch value := value.(type) {
	case int64:
		n = strconv.FormatInt(value, 10)
	case uint64:
		n = strconv.FormatUint(value, 10)
	case float64:
		n = strconv.FormatFloat(value, 'f', 0, 64)
	}
	width := len(n)
	if width >= min {
		return width
	}

	// else width < min
	return min
}

func (lg *Loggen) metaStats() MetaStats {
	var ms MetaStats
	var totalDuration time.Duration

	for _, stats := range lg.statsSlice {
		ms.workerCount++
		ms.totalEmitted += stats.Count
		totalDuration += stats.Duration
		ms.totalBytes += stats.Bytes
	}

	tmp := columnWidthMin(ms.totalEmitted, 6)
	if tmp > ms.maxCol2 {
		ms.maxCol2 = tmp
	}

	tmp = columnWidthMin(ms.totalBytes, 5)
	if tmp > ms.maxCol4 {
		ms.maxCol4 = tmp
	}

	ms.avgBytes = ms.totalBytes / uint64(ms.workerCount)
	ms.avgDuration = totalDuration / time.Duration(ms.workerCount)
	ms.avgEmitted = int(ms.totalEmitted) / ms.workerCount
	ms.avgEps = float64(ms.totalEmitted) / float64(totalDuration.Seconds())

	tmp = columnWidthMin(ms.avgEps, 6)
	if tmp > ms.maxCol5 {
		ms.maxCol5 = tmp
	}

	return ms
}

func (lg *Loggen) presentStats() {
	if lg.opt.Stats {
		slices.SortFunc(lg.statsSlice, func(a, b Stats) int {
			return cmp.Compare(a.Worker, b.Worker)
		})

		ms := lg.metaStats()
		col1 := 14 // column width will not vary
		col2 := ms.maxCol2
		col3 := 20 // column width will not vary
		col4 := ms.maxCol4
		col5 := ms.maxCol5
		//         1         2         3         4         5         6         7         8         9         1
		//123456789012345678901234567890123456789012345678901234567890
		//Worker       |Emit  |Duration           |Bytes |Eps
		header := fmt.Sprintf("%%s%%-%ds %%-%ds %%-%ds %%-%ds %%-%ds%%s", col1, col2, col3, col4, col5)
		fmt.Printf(header, "\n", "Worker", "Emit", "Duration", "Bytes", "Eps", "\n")

		format := fmt.Sprintf("%%-%ds %%-%dd %%-%ds %%-%dd %%-%d.0f%%s", col1, col2, col3, col4, col5)
		for _, stats := range lg.statsSlice {
			fmt.Printf(format, stats.Worker, stats.Count, microSecondFormat(stats.Duration), stats.Bytes, (float64(stats.Count) / float64(stats.Duration.Seconds())), "\n")
		}

		fmt.Println()
		avg := fmt.Sprintf("%%%ds %%-%dd %%-%ds %%-%dd %%-%d.0f%%s", col1, col2, col3, col4, col5)
		fmt.Printf(avg, "AVG", ms.avgEmitted, microSecondFormat(ms.avgDuration), ms.avgBytes, ms.avgEps, "\n")
		total := fmt.Sprintf("%%%ds %%-%dd %%%ds %%-%dd%%s", col1, col2, col3, col4)
		fmt.Printf(total, "TOTAL", ms.totalEmitted, " ", ms.totalBytes, "\n\n")
	}
}
