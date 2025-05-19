package bench

import (
	"fmt"
	"image/color"
	"sort"
	"time"

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
		p.Legend.Top = true
		p.Legend.Left = false
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
func PlotCombinedMetrics(results []Result, rpcTime time.Duration, filename string) error {
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
	p.Legend.Top = true
	p.Legend.Left = false
	p.Add(plotter.NewGrid())

	// Add baseline line for median eth_blockNumber call time
	if rpcTime != 0 {
		p.Y.Min = float64(rpcTime.Milliseconds() / 2)
		medianMs := float64(rpcTime.Milliseconds())
		baselineLine := plotter.NewFunction(func(x float64) float64 { return medianMs })
		baselineLine.Color = color.RGBA{R: 255, G: 0, B: 0, A: 128} // semi-transparent red
		baselineLine.Width = vg.Points(1.5)
		baselineLine.Dashes = []vg.Length{vg.Points(5), vg.Points(5)}

		p.Add(baselineLine)
		p.Legend.Add("RPC Time", baselineLine)
	}

	err := plotutil.AddLinePoints(p,
		"Total Time", totalPts,
	)
	if err != nil {
		return fmt.Errorf("failed to add line points: %w", err)
	}

	if err := p.Save(12*vg.Inch, 5*vg.Inch, filename); err != nil {
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
	p.Legend.Top = true
	p.Legend.Left = false
	p.Add(plotter.NewGrid())

	err := plotutil.AddLinePoints(p,
		"Async Total Time", asyncPts,
		"Sync Total Time", syncPts,
	)
	if err != nil {
		return fmt.Errorf("failed to add line points: %w", err)
	}

	if err := p.Save(12*vg.Inch, 5*vg.Inch, filename); err != nil {
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

	asyncTotalPts := make(plotter.XYs, len(asyncResults))
	syncTotalPts := make(plotter.XYs, len(syncResults))
	asyncSendPts := make(plotter.XYs, len(asyncResults))

	for i := range asyncResults {
		x := float64(i + 1)
		asyncTotalPts[i].X, asyncTotalPts[i].Y = x, float64(asyncResults[i].TotalTime)
		syncTotalPts[i].X, syncTotalPts[i].Y = x, float64(syncResults[i].TotalTime)
		asyncSendPts[i].X, asyncSendPts[i].Y = x, float64(asyncResults[i].SendTime)
	}

	p := plot.New()
	p.Title.Text = "Benchmark Time Comparison"
	p.X.Label.Text = "Transaction #"
	p.Y.Label.Text = "Time (ms)"
	p.Legend.Top = true
	p.Legend.Left = false
	p.Add(plotter.NewGrid())

	err := plotutil.AddLinePoints(p,
		"Async Total Time", asyncTotalPts,
		"Sync Total Time", syncTotalPts,
		"Async Send Time", asyncSendPts,
	)
	if err != nil {
		return fmt.Errorf("failed to add line points: %w", err)
	}

	if err := p.Save(12*vg.Inch, 5*vg.Inch, filename); err != nil {
		return fmt.Errorf("failed to save plot: %w", err)
	}
	return nil
}

func PlotWithBlockNumberBaseline(asyncResults, syncResults []Result, rpcTime time.Duration, filename string) error {
	n := len(asyncResults)
	if n == 0 || len(syncResults) != n {
		return fmt.Errorf("result length mismatch or empty")
	}

	asyncTotalPts := make(plotter.XYs, n)
	syncTotalPts := make(plotter.XYs, n)
	asyncSendPts := make(plotter.XYs, n)

	for i := 0; i < n; i++ {
		x := float64(i + 1)
		asyncTotalPts[i].X, asyncTotalPts[i].Y = x, float64(asyncResults[i].TotalTime)
		syncTotalPts[i].X, syncTotalPts[i].Y = x, float64(syncResults[i].TotalTime)
		asyncSendPts[i].X, asyncSendPts[i].Y = x, float64(asyncResults[i].SendTime)
	}

	p := plot.New()
	p.Title.Text = "Sync vs Async"
	p.X.Label.Text = "Tx #"
	p.Y.Label.Text = "Time (ms)"
	p.Legend.Top = true
	p.Legend.Left = false
	p.Add(plotter.NewGrid())

	// Add benchmark lines
	err := plotutil.AddLinePoints(p,
		"Async Total Time", asyncTotalPts,
		"Sync Total Time", syncTotalPts,
	)
	if err != nil {
		return err
	}

	// Add baseline line for median eth_blockNumber call time
	if rpcTime != 0 {
		medianMs := float64(rpcTime.Milliseconds())
		baselineLine := plotter.NewFunction(func(x float64) float64 { return medianMs })
		baselineLine.Color = color.RGBA{R: 255, G: 0, B: 0, A: 128} // semi-transparent red
		baselineLine.Width = vg.Points(1.5)
		baselineLine.Dashes = []vg.Length{vg.Points(5), vg.Points(5)}

		p.Add(baselineLine)
		p.Legend.Add("RPC Time", baselineLine)
	}

	if err := p.Save(12*vg.Inch, 5*vg.Inch, filename); err != nil {
		return err
	}
	return nil
}
