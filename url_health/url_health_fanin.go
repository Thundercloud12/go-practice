package urlhealth

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"
)

type FinalResult struct {
	Url           string        `json:"url"`
	Status_code   int           `json:"statuscode"`
	Success       bool          `json:"success"`
	Response_time time.Duration `json:"response_time"`
	Error         error         `json:"error"`
}

func producer(urls chan<- string, path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	defer close(urls)

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		urls <- scanner.Text()
	}

	return scanner.Err()
}

func worker_fanin(urls <-chan string, wg *sync.WaitGroup) chan FinalResult {
	result := make(chan FinalResult)

	go func() {
		defer wg.Done()
		defer close(result)

		client := &http.Client{}

		for url := range urls {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)

			req, err := http.NewRequest(http.MethodGet, url, nil)

			if err != nil {
				result <- FinalResult{
					Url:     url,
					Success: false,
					Error:   err,
				}
				cancel()
				continue
			}

			start := time.Now()

			resp, err := client.Do(req.WithContext(ctx))

			elapsed := time.Since(start)
			cancel()

			if err != nil {
				result <- FinalResult{
					Url:           url,
					Status_code:   0,
					Success:       false,
					Response_time: elapsed,
					Error:         err,
				}
				continue
			}

			resp.Body.Close()

			result <- FinalResult{
				Url:           url,
				Status_code:   resp.StatusCode,
				Success:       true,
				Response_time: elapsed,
				Error:         nil,
			}
		}
	}()

	return result
}

func fanIn(channels ...chan FinalResult) chan FinalResult {
	var wg sync.WaitGroup

	merged := make(chan FinalResult)

	wg.Add(len(channels))

	for _, ch := range channels {
		go func(ch chan FinalResult) {
			defer wg.Done()

			for result := range ch {
				merged <- result
			}
		}(ch)
	}

	go func() {
		wg.Wait()
		close(merged)
	}()

	return merged
}

func url_health_fanin(path string, workers int) ([]FinalResult, error) {
	var result []FinalResult

	if workers <= 0 {
		return result, fmt.Errorf("workers must be greater than 0")
	}

	urls := make(chan string)

	var workerWG sync.WaitGroup
	channels := make([]chan FinalResult, 0, workers)

	// Fan-out
	for i := 0; i < workers; i++ {
		workerWG.Add(1)

		ch := worker_fanin(urls, &workerWG)
		channels = append(channels, ch)
	}

	// Producer
	producerErr := make(chan error, 1)

	go func() {
		producerErr <- producer(urls, path)
	}()

	// Fan-in
	merged := fanIn(channels...)

	// Collect results
	for r := range merged {
		result = append(result, r)
	}

	// Wait for workers
	workerWG.Wait()

	if err := <-producerErr; err != nil {
		return result, err
	}

	return result, nil
}