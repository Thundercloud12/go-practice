package ratelimiterbucket

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

	defer close(urls)
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		urls <- scanner.Text()
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

func worker(result chan<- FinalResult, urls <-chan string, wg *sync.WaitGroup, tokens <-chan struct{}, ctx context.Context) {
	defer wg.Done()
	client := &http.Client{}

	for url := range urls {
		func(url string) {
			ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
			defer cancel()

			select {
			case <-tokens:
			case <-ctx.Done():
				return
			}

			req, err := http.NewRequest(http.MethodGet, url, nil)

			if err != nil {
				result <- FinalResult{
					Url:           url,
					Status_code:   0,
					Success:       false,
					Response_time: 0,
					Error:         err,
				}
				return
			}

			start := time.Now()
			resp, err := client.Do(req.WithContext(ctx))
			if err != nil {
				result <- FinalResult{
					Url:           url,
					Status_code:   0,
					Success:       false,
					Response_time: 0,
					Error:         err,
				}
				return
			}

			defer resp.Body.Close()
			elapsed := time.Since(start)

			result <- FinalResult{
				Url:           url,
				Status_code:   resp.StatusCode,
				Success:       true,
				Response_time: elapsed,
				Error:         nil,
			}

		}(url)
	}
}

func url_health(path string, workers int, limit int) ([]FinalResult, error) {

	var result []FinalResult

	if workers <= 0 || limit <= 0 {
		return result, fmt.Errorf("workers and limit must be greater than 0")
	}

	urls := make(chan string)
	result_channel := make(chan FinalResult)
	var wg sync.WaitGroup

	tokens := make(chan struct{})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		intervali := time.Second / time.Duration(limit)
		tickeri := time.NewTicker(intervali)
		defer tickeri.Stop()
		defer close(tokens)

		for {
			select {
			case <-tickeri.C:
				select {
				case tokens <- struct{}{}:
				case <-ctx.Done():
					return
				}

			case <-ctx.Done():
				return
			}
		}
	}()

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go worker(result_channel, urls, &wg, tokens, ctx)
	}

	go func() {
		wg.Wait()
		close(result_channel)
		cancel()
	}()

	producerErr := make(chan error, 1)

	go func() {
		producerErr <- producer(urls, path)
	}()

	for re := range result_channel {
		result = append(result, re)
	}

	if err := <-producerErr; err != nil {
		return result, err
	}

	return result, nil
}
