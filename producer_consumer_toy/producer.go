package producer

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"os"
	"strings"
	"sync"
)

type FinalResult struct {
	TotalLines   int               `json:"totallines"`
	ValidLines   int               `json:"validlines"`
	InvalidLines int               `json:"invalidlines"`
	INFO         int               `json:"info"`
	WARN         int               `json:"warn"`
	ERROR        int               `json:"error"`
	UniqueUsers  int               `json:"unique_users"`
	Users        map[string]struct{} `json:"-"`
}

func Producer(lines chan<- string, path string) error {
	file, err := os.Open(path)
	defer close(lines)

	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		lines <- scanner.Text()
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

func worker(lines <-chan string, results chan<- FinalResult, wg *sync.WaitGroup) {
	defer wg.Done()

	result := FinalResult{
		Users: make(map[string]struct{}),
	}

	for msg := range lines {
		result.TotalLines++

		reader := csv.NewReader(strings.NewReader(msg))

		fields, err := reader.Read()

		if err != nil ||
			len(fields) != 3 ||
			(fields[0] != "INFO" && fields[0] != "WARN" && fields[0] != "ERROR") ||
			fields[1] == "" ||
			fields[2] == "" {

			result.InvalidLines++
			continue
		}

		result.ValidLines++
		result.Users[fields[1]] = struct{}{}

		if fields[0] == "ERROR" {
			result.ERROR++
		} else if fields[0] == "INFO" {
			result.INFO++
		} else if fields[0] == "WARN" {
			result.WARN++
		}
	}

	result.UniqueUsers = len(result.Users)
	results <- result
}

func aggregator(results <-chan FinalResult) FinalResult {
	final := FinalResult{
		Users: make(map[string]struct{}),
	}

	for result := range results {
		final.TotalLines += result.TotalLines
		final.ValidLines += result.ValidLines
		final.InvalidLines += result.InvalidLines
		final.INFO += result.INFO
		final.WARN += result.WARN
		final.ERROR += result.ERROR

		for user := range result.Users {
			final.Users[user] = struct{}{}
		}
	}

	final.UniqueUsers = len(final.Users)

	return final
}

func ProcessFile(path string, workers int) (FinalResult, error) {
	if workers <= 0 {
		return FinalResult{}, fmt.Errorf("workers must be greater than 0")
	}

	lines := make(chan string)
	results := make(chan FinalResult)

	var wg sync.WaitGroup

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go worker(lines, results, &wg)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	producerErr := make(chan error, 1)

	go func() {
		producerErr <- Producer(lines, path)
	}()

	result := aggregator(results)

	if err := <-producerErr; err != nil {
		return FinalResult{}, err
	}

	return result, nil
}