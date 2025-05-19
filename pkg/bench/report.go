package bench

import (
	"fmt"
	"image/color"
	"sort"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
	"log"
)

func median(durations []int64) int64 {
	sorted := make([]int64, len(durations))
	copy(sorted, durations)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })
	mid := len(sorted) / 2
	if len(sorted)%2 == 0 {
		return (sorted[mid-1] + sorted[mid]) / 2
	}
	return sorted[mid]
}

func medianInt64(data []int64) int64 {
	sorted := make([]int64, len(data))
	copy(sorted, data)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })
	mid := len(sorted) / 2
	if len(sorted)%2 == 0 {
		return (sorted[mid-1] + sorted[mid]) / 2
	}
	return sorted[mid]
}

func truncateHash(h string) string {
	if len(h) < 14 {
		return h
	}
	return h[:8] + "…" + h[len(h)-4:]
}

func PrintReport(results []Result) {
	if len(results) == 0 {
		fmt.Println("No results to report")
		return
	}

	var totalElapsed int64
	sendTimes := make([]int64, len(results))
	confirmTimes := make([]int64, len(results))
	totalTimes := make([]int64, len(results))

	for i, r := range results {
		sendTimes[i] = r.SendTime
		confirmTimes[i] = r.ConfirmTime
		totalTimes[i] = r.TotalTime
		totalElapsed += r.TotalTime
	}

	fmt.Printf("\nTotal time for all transactions: %.3fs\n\n", float64(totalElapsed)/1000)

	fmt.Println("Individual Transaction Results:")
	fmt.Println("TX#   SEND (ms)    CONFIRM (ms) TOTAL (ms)   HASH")
	fmt.Println("------------------------------------------------------------------------------------------------------------------------")

	for _, r := range results {
		fmt.Printf("%-5d %-12d %-13d %-12d %s\n",
			r.TxIndex,
			r.SendTime,
			r.ConfirmTime,
			r.TotalTime,
			truncateHash(r.TxHash),
		)
	}

	fmt.Println("\nLATENCY STATISTICS:")
	fmt.Println("              MIN (ms)   MAX (ms)   AVG (ms)   MEDIAN (ms)")
	fmt.Println("-------------------------------------------------------")

	computeStats := func(times []int64) (min, max, avg, med int64) {
		min = times[0]
		max = times[0]
		var sum int64
		for _, t := range times {
			sum += t
			if t < min {
				min = t
			}
			if t > max {
				max = t
			}
		}
		avg = sum / int64(len(times))
		med = median(times)
		return
	}

	minSend, maxSend, avgSend, medSend := computeStats(sendTimes)
	minConfirm, maxConfirm, avgConfirm, medConfirm := computeStats(confirmTimes)
	minTotal, maxTotal, avgTotal, medTotal := computeStats(totalTimes)

	fmt.Printf("%-13s %-10d %-10d %-10d %-10d\n", "Send time:", minSend, maxSend, avgSend, medSend)
	fmt.Printf("%-13s %-10d %-10d %-10d %-10d\n", "Confirm time:", minConfirm, maxConfirm, avgConfirm, medConfirm)
	fmt.Printf("%-13s %-10d %-10d %-10d %-10d\n", "Total time:", minTotal, maxTotal, avgTotal, medTotal)
}

// PlotMetrics generates PNG plots for send, confirm, and total times.
func PlotMetrics(results []Result, filenamePrefix string) error {
	if len(results) == 0 {
		return fmt.Errorf("no results to plot")
	}

	sendPts := make(plotter.XYs, len(results))
	confirmPts := make(plotter.XYs, len(results))
	totalPts := make(plotter.XYs, len(results))

	for i, r := range results {
		x := float64(i + 1)
		sendPts[i].X, sendPts[i].Y = x, float64(r.SendTime)
		confirmPts[i].X, confirmPts[i].Y = x, float64(r.ConfirmTime)
		totalPts[i].X, totalPts[i].Y = x, float64(r.TotalTime)
	}

	savePlot := func(title, filename string, pts plotter.XYs) error {
		p := plot.New()
		p.Title.Text = title
		p.X.Label.Text = "Transaction #"
		p.Y.Label.Text = "Time (ms)"
		p.Add(plotter.NewGrid())

		err := plotutil.AddLinePoints(p, title, pts)
		if err != nil {
			return err
		}

		if err := p.Save(6*vg.Inch, 4*vg.Inch, filename); err != nil {
			return err
		}
		return nil
	}

	if err := savePlot("Send Time", filenamePrefix+"_send.png", sendPts); err != nil {
		log.Printf("Failed to save send time plot: %v", err)
	}
	if err := savePlot("Confirm Time", filenamePrefix+"_confirm.png", confirmPts); err != nil {
		log.Printf("Failed to save confirm time plot: %v", err)
	}
	if err := savePlot("Total Time", filenamePrefix+"_total.png", totalPts); err != nil {
		log.Printf("Failed to save total time plot: %v", err)
	}

	return nil
}

// PlotCombinedMetrics creates a single PNG plot with send, confirm, and total times.
func PlotCombinedMetrics(results []Result, filename string) error {
	if len(results) == 0 {
		return fmt.Errorf("no results to plot")
	}

	sendPts := make(plotter.XYs, len(results))
	confirmPts := make(plotter.XYs, len(results))
	totalPts := make(plotter.XYs, len(results))

	for i, r := range results {
		x := float64(i + 1)
		sendPts[i].X, sendPts[i].Y = x, float64(r.SendTime)
		confirmPts[i].X, confirmPts[i].Y = x, float64(r.ConfirmTime)
		totalPts[i].X, totalPts[i].Y = x, float64(r.TotalTime)
	}

	p := plot.New()
	p.Title.Text = "Benchmark Time Metrics"
	p.X.Label.Text = "Transaction #"
	p.Y.Label.Text = "Time (ms)"
	p.Add(plotter.NewGrid())

	err := plotutil.AddLinePoints(p,
		"Send Time", sendPts,
		"Confirm Time", confirmPts,
		"Total Time", totalPts,
	)
	if err != nil {
		return fmt.Errorf("failed to add line points: %w", err)
	}

	if err := p.Save(8*vg.Inch, 5*vg.Inch, filename); err != nil {
		return fmt.Errorf("failed to save plot: %w", err)
	}
	return nil
}

// PlotCombinedTotalTime plots total times of async and sync benchmarks on the same chart.
func PlotCombinedTotalTime(asyncResults, syncResults []Result, filename string) error {
	if len(asyncResults) == 0 || len(syncResults) == 0 {
		return fmt.Errorf("no results to plot")
	}
	if len(asyncResults) != len(syncResults) {
		return fmt.Errorf("async and sync results length mismatch")
	}

	asyncPts := make(plotter.XYs, len(asyncResults))
	syncPts := make(plotter.XYs, len(syncResults))

	for i := range asyncResults {
		x := float64(i + 1)
		asyncPts[i].X, asyncPts[i].Y = x, float64(asyncResults[i].TotalTime)
		syncPts[i].X, syncPts[i].Y = x, float64(syncResults[i].TotalTime)
	}

	p := plot.New()
	p.Title.Text = "Total Transaction Time Comparison"
	p.X.Label.Text = "Transaction #"
	p.Y.Label.Text = "Total Time (ms)"
	p.Add(plotter.NewGrid())

	err := plotutil.AddLinePoints(p,
		"Async Total Time", asyncPts,
		"Sync Total Time", syncPts,
	)
	if err != nil {
		return fmt.Errorf("failed to add line points: %w", err)
	}

	if err := p.Save(8*vg.Inch, 5*vg.Inch, filename); err != nil {
		return fmt.Errorf("failed to save plot: %w", err)
	}
	return nil
}

func PlotCombinedTotalTimeWithMedian(asyncResults, syncResults []Result, filename string) error {
	if len(asyncResults) == 0 || len(syncResults) == 0 {
		return fmt.Errorf("no results to plot")
	}
	if len(asyncResults) != len(syncResults) {
		return fmt.Errorf("async and sync results length mismatch")
	}

	asyncPts := make(plotter.XYs, len(asyncResults))
	syncPts := make(plotter.XYs, len(syncResults))

	for i := range asyncResults {
		x := float64(i + 1)
		asyncPts[i].X, asyncPts[i].Y = x, float64(asyncResults[i].TotalTime)
		syncPts[i].X, syncPts[i].Y = x, float64(syncResults[i].TotalTime)
	}

	p := plot.New()
	p.Title.Text = "Total Transaction Time Comparison with Median Lines"
	p.X.Label.Text = "Transaction #"
	p.Y.Label.Text = "Total Time (ms)"
	p.Add(plotter.NewGrid())

	// Add async and sync total time lines
	err := plotutil.AddLinePoints(p,
		"Async Total Time", asyncPts,
		"Sync Total Time", syncPts,
	)
	if err != nil {
		return fmt.Errorf("failed to add line points: %w", err)
	}

	// Compute medians
	asyncTimes := make([]int64, len(asyncResults))
	syncTimes := make([]int64, len(syncResults))
	for i := range asyncResults {
		asyncTimes[i] = asyncResults[i].TotalTime
		syncTimes[i] = syncResults[i].TotalTime
	}
	asyncMedian := float64(medianInt64(asyncTimes))
	syncMedian := float64(medianInt64(syncTimes))

	// Create median lines spanning full X range
	xmin := 1.0
	xmax := float64(len(asyncResults))

	asyncMedianLine := plotter.NewFunction(func(x float64) float64 { return asyncMedian })
	asyncMedianLine.Color = color.RGBA{R: 255, A: 255} // red
	asyncMedianLine.Width = vg.Points(1.5)
	asyncMedianLine.Dashes = []vg.Length{vg.Points(5), vg.Points(5)}

	syncMedianLine := plotter.NewFunction(func(x float64) float64 { return syncMedian })
	syncMedianLine.Color = color.RGBA{B: 255, A: 255} // blue
	syncMedianLine.Width = vg.Points(1.5)
	syncMedianLine.Dashes = []vg.Length{vg.Points(5), vg.Points(5)}

	p.Add(asyncMedianLine, syncMedianLine)

	// Add legend entries for median lines
	p.Legend.Add("Async Median", asyncMedianLine)
	p.Legend.Add("Sync Median", syncMedianLine)

	p.X.Min = xmin
	p.X.Max = xmax

	if err := p.Save(8*vg.Inch, 5*vg.Inch, filename); err != nil {
		return fmt.Errorf("failed to save plot: %w", err)
	}
	return nil
}
