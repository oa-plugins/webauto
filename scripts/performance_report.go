package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"sort"
	"strconv"
	"time"
)

// BenchmarkResult represents a single benchmark result
type BenchmarkResult struct {
	Command    string
	Iteration  int
	DurationMs int
	TargetMs   int
	Status     string
}

// CommandStats holds aggregated statistics for a command
type CommandStats struct {
	Command    string
	Count      int
	MinMs      int
	MaxMs      int
	AvgMs      int
	TargetMs   int
	PassRate   float64
	TotalMs    int
}

func main() {
	fmt.Println("📊 webauto Performance Report Generator")
	fmt.Println("========================================\n")

	// Read benchmark results
	results, err := readBenchmarkResults("benchmark_results.csv")
	if err != nil {
		fmt.Printf("❌ Error reading benchmark results: %v\n", err)
		os.Exit(1)
	}

	if len(results) == 0 {
		fmt.Println("⚠️  No benchmark results found. Run ./scripts/benchmark.sh first.")
		os.Exit(1)
	}

	// Aggregate stats by command
	stats := aggregateStats(results)

	// Print summary
	printSummary(stats)

	// Print detailed results
	printDetailedResults(stats)

	// Print recommendations
	printRecommendations(stats)

	// Calculate overall score
	printOverallScore(stats)
}

func readBenchmarkResults(filename string) ([]BenchmarkResult, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	var results []BenchmarkResult
	for i, record := range records {
		if i == 0 {
			// Skip header
			continue
		}

		iteration, _ := strconv.Atoi(record[1])
		duration, _ := strconv.Atoi(record[2])
		target, _ := strconv.Atoi(record[3])

		results = append(results, BenchmarkResult{
			Command:    record[0],
			Iteration:  iteration,
			DurationMs: duration,
			TargetMs:   target,
			Status:     record[4],
		})
	}

	return results, nil
}

func aggregateStats(results []BenchmarkResult) map[string]*CommandStats {
	statsMap := make(map[string]*CommandStats)

	for _, result := range results {
		stats, exists := statsMap[result.Command]
		if !exists {
			stats = &CommandStats{
				Command:  result.Command,
				MinMs:    result.DurationMs,
				MaxMs:    result.DurationMs,
				TargetMs: result.TargetMs,
			}
			statsMap[result.Command] = stats
		}

		stats.Count++
		stats.TotalMs += result.DurationMs

		if result.DurationMs < stats.MinMs {
			stats.MinMs = result.DurationMs
		}
		if result.DurationMs > stats.MaxMs {
			stats.MaxMs = result.DurationMs
		}

		if result.Status == "PASS" {
			stats.PassRate++
		}
	}

	// Calculate averages and pass rates
	for _, stats := range statsMap {
		stats.AvgMs = stats.TotalMs / stats.Count
		stats.PassRate = (stats.PassRate / float64(stats.Count)) * 100
	}

	return statsMap
}

func printSummary(stats map[string]*CommandStats) {
	fmt.Println("## Summary")
	fmt.Println("")
	fmt.Printf("**Date**: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Printf("**Total Commands**: %d\n", len(stats))
	fmt.Println("")

	fmt.Println("| Command | Avg (ms) | Min (ms) | Max (ms) | Target (ms) | Pass Rate | Status |")
	fmt.Println("|---------|----------|----------|----------|-------------|-----------|--------|")

	// Sort commands by name
	commands := make([]string, 0, len(stats))
	for cmd := range stats {
		commands = append(commands, cmd)
	}
	sort.Strings(commands)

	for _, cmd := range commands {
		s := stats[cmd]
		status := "✅"
		if s.PassRate < 100 {
			status = "❌"
		}

		fmt.Printf("| %s | %d | %d | %d | %d | %.1f%% | %s |\n",
			s.Command, s.AvgMs, s.MinMs, s.MaxMs, s.TargetMs, s.PassRate, status)
	}
	fmt.Println("")
}

func printDetailedResults(stats map[string]*CommandStats) {
	fmt.Println("## Detailed Analysis")
	fmt.Println("")

	// Sort by average time (slowest first)
	type kv struct {
		Command string
		Stats   *CommandStats
	}
	var sorted []kv
	for cmd, s := range stats {
		sorted = append(sorted, kv{cmd, s})
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Stats.AvgMs > sorted[j].Stats.AvgMs
	})

	for _, pair := range sorted {
		s := pair.Stats

		fmt.Printf("### %s\n", s.Command)
		fmt.Printf("- **Average**: %dms (target: %dms)\n", s.AvgMs, s.TargetMs)
		fmt.Printf("- **Range**: %dms - %dms\n", s.MinMs, s.MaxMs)
		fmt.Printf("- **Pass Rate**: %.1f%% (%d/%d iterations)\n", s.PassRate, int(s.PassRate*float64(s.Count)/100), s.Count)

		variance := float64(s.MaxMs - s.MinMs)
		avgVariance := variance / float64(s.AvgMs) * 100
		fmt.Printf("- **Variance**: %.1f%%\n", avgVariance)

		// Performance assessment
		if s.AvgMs <= s.TargetMs {
			fmt.Printf("- ✅ **Status**: MEETING TARGET\n")
		} else {
			overhead := s.AvgMs - s.TargetMs
			fmt.Printf("- ❌ **Status**: %dms OVER TARGET (%.1f%% slower)\n",
				overhead, float64(overhead)/float64(s.TargetMs)*100)
		}

		fmt.Println("")
	}
}

func printRecommendations(stats map[string]*CommandStats) {
	fmt.Println("## Optimization Recommendations")
	fmt.Println("")

	hasIssues := false
	for cmd, s := range stats {
		if s.AvgMs > s.TargetMs {
			hasIssues = true
			overhead := s.AvgMs - s.TargetMs

			fmt.Printf("### %s (%.1f%% over target)\n", cmd, float64(overhead)/float64(s.TargetMs)*100)

			// Specific recommendations based on command type
			switch cmd {
			case "browser-launch", "browser-close":
				fmt.Println("- Consider browser process reuse or faster launch options")
				fmt.Println("- Investigate Playwright startup overhead")
				fmt.Println("- Profile Node.js subprocess creation time")

			case "session-list":
				fmt.Println("- Implement in-memory session cache")
				fmt.Println("- Reduce file system operations")
				fmt.Println("- Use singleton SessionManager pattern")

			case "session-close":
				fmt.Println("- Optimize process termination")
				fmt.Println("- Reduce file I/O during cleanup")

			case "page-navigate":
				fmt.Println("- This is network-bound, but check:")
				fmt.Println("  - TCP connection reuse")
				fmt.Println("  - IPC overhead")
				fmt.Println("  - Wait strategy optimization")

			case "element-click", "element-type":
				fmt.Println("- Implement TCP connection pooling")
				fmt.Println("- Reduce timeout values for fast operations")
				fmt.Println("- Optimize JSON serialization overhead")

			case "page-screenshot", "page-pdf":
				fmt.Println("- Optimize base64 encoding/decoding")
				fmt.Println("- Consider streaming instead of buffering")
				fmt.Println("- Check Playwright screenshot options")
			}

			fmt.Println("")
		}
	}

	if !hasIssues {
		fmt.Println("✅ All commands are meeting their performance targets!")
		fmt.Println("")
	}
}

func printOverallScore(stats map[string]*CommandStats) {
	fmt.Println("## Overall Performance Score")
	fmt.Println("")

	totalCommands := len(stats)
	passingCommands := 0
	totalOverhead := 0
	totalTarget := 0

	for _, s := range stats {
		totalTarget += s.TargetMs
		if s.AvgMs <= s.TargetMs {
			passingCommands++
		} else {
			totalOverhead += (s.AvgMs - s.TargetMs)
		}
	}

	passRate := float64(passingCommands) / float64(totalCommands) * 100

	fmt.Printf("**Commands Meeting Target**: %d/%d (%.1f%%)\n", passingCommands, totalCommands, passRate)

	if totalOverhead > 0 {
		fmt.Printf("**Total Overhead**: %dms\n", totalOverhead)
	}

	// Calculate grade
	grade := "F"
	emoji := "❌"
	if passRate >= 90 {
		grade = "A"
		emoji = "🎉"
	} else if passRate >= 80 {
		grade = "B"
		emoji = "✅"
	} else if passRate >= 70 {
		grade = "C"
		emoji = "⚠️"
	} else if passRate >= 60 {
		grade = "D"
		emoji = "❌"
	}

	fmt.Printf("\n**Grade**: %s %s\n\n", grade, emoji)

	if passRate < 100 {
		fmt.Println("**Next Steps**:")
		fmt.Println("1. Review optimization recommendations above")
		fmt.Println("2. Profile slow commands using `go test -cpuprofile=cpu.prof -bench=.`")
		fmt.Println("3. Implement targeted optimizations")
		fmt.Println("4. Re-run benchmarks to validate improvements")
		fmt.Println("")
	} else {
		fmt.Println("🎉 Excellent! All performance targets are being met!")
		fmt.Println("")
	}
}
