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
	defer close(urls)

	if err != nil {
		return err
	}
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

func worker(result chan<- FinalResult, urls <-chan string, wg *sync.WaitGroup) {
	defer wg.Done()
	client := &http.Client{}

	for url := range urls {
		func(url string) {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
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

func url_health(path string, workers int) ([]FinalResult, error) {

	var result []FinalResult

	if workers <= 0 {
		return result, fmt.Errorf("workers must be greater than 0")
	}

	urls := make(chan string)
	result_channel := make(chan FinalResult)
	var wg sync.WaitGroup

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go worker(result_channel, urls, &wg)
	}
	go func() {
		wg.Wait()
		close(result_channel)
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
