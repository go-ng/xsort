package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/cep21/benchparse"
)

func syntaxError() {
	fmt.Fprintf(flag.CommandLine.Output(), "syntax: benchmark_csv <benchmarks file path> <Sort/Slice CSV output> <Appended CSV output>\n")
	flag.CommandLine.ErrorHandling()
	os.Exit(2)
}

func main() {
	flag.Parse()
	if flag.NArg() != 3 {
		syntaxError()
	}
	benchPath := flag.Arg(0)
	sliceResultsPath := flag.Arg(1)
	appendedResultsPath := flag.Arg(2)

	run, err := parseFile(benchPath)
	if err != nil {
		panic(err)
	}

	sliceBenchmarks, err := scanSliceBenchmarks(run)
	if err != nil {
		panic(err)
	}

	appendedBenchmarks, err := scanAppendedBenchmarks(run)
	if err != nil {
		panic(err)
	}

	err = generateCSVForSlice(sliceResultsPath, sliceBenchmarks)
	if err != nil {
		panic(err)
	}

	err = generateCSVForAppended(appendedResultsPath, appendedBenchmarks)
	if err != nil {
		panic(err)
	}
}

func parseFile(filePath string) (*benchparse.Run, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("unable to open file '%s': %w", filePath, err)
	}
	defer f.Close()

	run, err := benchparse.Decoder{}.Decode(f)
	if err != nil {
		return nil, fmt.Errorf("unable to parse file '%s': %w", filePath, err)
	}

	return run, nil
}

const (
	benchmarkName = "Benchmark"
)

type sliceBenchmarks map[string]map[uint64][]*benchparse.BenchmarkResult

func scanSliceBenchmarks(run *benchparse.Run) (sliceBenchmarks, error) {
	m := sliceBenchmarks{}
	for idx, result := range run.Results {
		nameParts := strings.Split(result.Name, "/")
		testFullName := nameParts[0]
		if !strings.HasPrefix(testFullName, benchmarkName) {
			return nil, fmt.Errorf("invalid result (%#v), the name is not Benchmark*", result)
		}

		if !strings.Contains(testFullName, "Slice") && !strings.Contains(testFullName, "Sort") {
			continue
		}

		testName := testFullName[len(benchmarkName):]
		if m[testName] == nil {
			m[testName] = map[uint64][]*benchparse.BenchmarkResult{}
		}

		sliceSizePart := strings.Split(nameParts[1], "-")[0]
		sliceSize, err := strconv.ParseUint(sliceSizePart, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("unable to parse sliceSize in '%s': %w", sliceSizePart, err)
		}

		m[testName][sliceSize] = append(m[testName][sliceSize], &run.Results[idx])
	}
	return m, nil
}

type appendedBenchmarks map[string]map[uint64][]*benchparse.BenchmarkResult

func scanAppendedBenchmarks(run *benchparse.Run) (appendedBenchmarks, error) {
	m := appendedBenchmarks{}
	for idx, result := range run.Results {
		nameParts := strings.Split(result.Name, "/")
		testFullName := nameParts[0]
		if !strings.HasPrefix(testFullName, benchmarkName) {
			return nil, fmt.Errorf("invalid result (%#v), the name is not Benchmark*", result)
		}

		if testFullName != "BenchmarkAppended" {
			continue
		}

		totalSizeStr := strings.Split(nameParts[1], "-")[1]
		tailSizeStr := strings.Split(nameParts[2], "-")[1]
		funcNameParts := strings.Split(nameParts[3], "-")
		caseName := fmt.Sprintf("%s-%s", strings.Join(funcNameParts[:len(funcNameParts)-1], "-"), totalSizeStr)

		tailSize, err := strconv.ParseUint(tailSizeStr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("unable to parse tailSize in '%s': %w", tailSizeStr, err)
		}

		if m[caseName] == nil {
			m[caseName] = make(map[uint64][]*benchparse.BenchmarkResult)
		}

		m[caseName][tailSize] = append(m[caseName][tailSize], &run.Results[idx])
	}
	return m, nil
}

func generateCSVForSlice(outputPath string, m sliceBenchmarks) (err error) {
	var funcNames []string
	sizesMap := map[uint64]struct{}{}
	for funcName, m := range m {
		funcNames = append(funcNames, funcName)
		for sliceSize := range m {
			sizesMap[sliceSize] = struct{}{}
		}
	}
	sort.Strings(funcNames)

	var sizes []uint64
	for size := range sizesMap {
		sizes = append(sizes, size)
	}
	sort.Slice(sizes, func(i, j int) bool {
		return sizes[i] < sizes[j]
	})

	latencies := make([][][]float64, len(sizes))
	for idx := range sizes {
		latencies[idx] = make([][]float64, len(funcNames))
	}
	for sizeIdx, size := range sizes {
		for funcIdx, funcName := range funcNames {
			results := m[funcName][size]
			for _, result := range results {
				for _, value := range result.Values {
					if value.Unit == benchparse.UnitRuntime {
						latencies[sizeIdx][funcIdx] = append(latencies[sizeIdx][funcIdx], value.Value)
					}
				}
			}
		}
	}

	latenciesForCSV := make([][]string, len(sizes))
	for sizeIdx := range sizes {
		latenciesForCSV[sizeIdx] = make([]string, len(funcNames))
		for funcIdx := range latencies[sizeIdx] {
			var sum float64
			for _, value := range latencies[sizeIdx][funcIdx] {
				sum += value
			}
			latenciesForCSV[sizeIdx][funcIdx] = strconv.FormatFloat(sum/float64(len(latencies[sizeIdx][funcIdx])), 'f', 2, 64)
		}
	}

	f, err := os.OpenFile(outputPath, os.O_WRONLY|os.O_EXCL|os.O_CREATE, 0640)
	if err != nil {
		return fmt.Errorf("unable to create file '%s': %w", outputPath, err)
	}
	defer f.Close()

	w := csv.NewWriter(f)

	if err := w.Write(append([]string{"size"}, funcNames...)); err != nil {
		return fmt.Errorf("unable to write CSV: %w", err)
	}

	for sizeIdx, size := range sizes {
		outLine := append([]string{fmt.Sprintf("%d", size)}, latenciesForCSV[sizeIdx]...)
		if err := w.Write(outLine); err != nil {
			return fmt.Errorf("unable to write CSV: %w", err)
		}
	}

	w.Flush()
	return w.Error()
}

func generateCSVForAppended(outputPath string, m appendedBenchmarks) error {
	var caseNames []string
	tailSizeMap := map[uint64]struct{}{}
	for caseName, m := range m {
		if !strings.HasSuffix(caseName, "-1048576") {
			continue
		}
		caseNames = append(caseNames, caseName)
		for tailSize := range m {
			tailSizeMap[tailSize] = struct{}{}
		}
	}
	sort.Strings(caseNames)

	var tailSizes []uint64
	for tailSize := range tailSizeMap {
		tailSizes = append(tailSizes, tailSize)
	}
	sort.Slice(tailSizes, func(i, j int) bool {
		return tailSizes[i] < tailSizes[j]
	})

	latencies := make([][][]float64, len(tailSizes))
	for idx := range tailSizes {
		latencies[idx] = make([][]float64, len(caseNames))
	}
	for tailSizeIdx, tailSize := range tailSizes {
		for caseIdx, caseName := range caseNames {
			results := m[caseName][tailSize]
			for _, result := range results {
				for _, value := range result.Values {
					if value.Unit == benchparse.UnitRuntime {
						latencies[tailSizeIdx][caseIdx] = append(latencies[tailSizeIdx][caseIdx], value.Value)
					}
				}
			}
		}
	}

	latenciesForCSV := make([][]string, len(tailSizes))
	for sizeIdx := range tailSizes {
		latenciesForCSV[sizeIdx] = make([]string, len(caseNames))
		for funcIdx := range latencies[sizeIdx] {
			var sum float64
			for _, value := range latencies[sizeIdx][funcIdx] {
				sum += value
			}
			latenciesForCSV[sizeIdx][funcIdx] = strconv.FormatFloat(sum/float64(len(latencies[sizeIdx][funcIdx])), 'f', 2, 64)
		}
	}

	f, err := os.OpenFile(outputPath, os.O_WRONLY|os.O_EXCL|os.O_CREATE, 0640)
	if err != nil {
		return fmt.Errorf("unable to create file '%s': %w", outputPath, err)
	}
	defer f.Close()

	w := csv.NewWriter(f)

	if err := w.Write(append([]string{"tailSize"}, caseNames...)); err != nil {
		return fmt.Errorf("unable to write CSV: %w", err)
	}

	for sizeIdx, size := range tailSizes {
		outLine := append([]string{fmt.Sprintf("%d", size)}, latenciesForCSV[sizeIdx]...)
		if err := w.Write(outLine); err != nil {
			return fmt.Errorf("unable to write CSV: %w", err)
		}
	}

	w.Flush()
	return w.Error()
}
