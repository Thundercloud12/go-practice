package ratelimiter

import (
	"sync"
	"testing"
	"time"
)

func TestRateLimiter_FirstRequestAllowed(t *testing.T) {
	rl := NewRateLimiter(5, time.Second)

	if !rl.Allow("client-1") {
		t.Fatal("first request should be allowed")
	}
}

func TestRateLimiter_MaxRequests(t *testing.T) {
	const maxRequests = 5

	rl := NewRateLimiter(maxRequests, time.Second)

	for i := 0; i < maxRequests; i++ {
		if !rl.Allow("client-1") {
			t.Fatalf("request %d should be allowed", i+1)
		}
	}

	if rl.Allow("client-1") {
		t.Fatal("request beyond maxRequests should be rejected")
	}
}

func TestRateLimiter_DifferentClients(t *testing.T) {
	rl := NewRateLimiter(2, time.Second)

	// Client 1 uses its entire limit.
	if !rl.Allow("client-1") {
		t.Fatal("client-1 request 1 should be allowed")
	}

	if !rl.Allow("client-1") {
		t.Fatal("client-1 request 2 should be allowed")
	}

	if rl.Allow("client-1") {
		t.Fatal("client-1 request 3 should be rejected")
	}

	// Client 2 should have its own independent limit.
	if !rl.Allow("client-2") {
		t.Fatal("client-2 request 1 should be allowed")
	}

	if !rl.Allow("client-2") {
		t.Fatal("client-2 request 2 should be allowed")
	}

	if rl.Allow("client-2") {
		t.Fatal("client-2 request 3 should be rejected")
	}
}

func TestRateLimiter_WindowExpiration(t *testing.T) {
	rl := NewRateLimiter(2, 50*time.Millisecond)

	if !rl.Allow("client-1") {
		t.Fatal("request 1 should be allowed")
	}

	if !rl.Allow("client-1") {
		t.Fatal("request 2 should be allowed")
	}

	if rl.Allow("client-1") {
		t.Fatal("request 3 should be rejected while inside window")
	}

	// Wait for the previous requests to expire.
	time.Sleep(60 * time.Millisecond)

	if !rl.Allow("client-1") {
		t.Fatal("request should be allowed after window expires")
	}
}

func TestRateLimiter_OnlyExpiredRequestsRemoved(t *testing.T) {
	rl := NewRateLimiter(2, 50*time.Millisecond)

	if !rl.Allow("client-1") {
		t.Fatal("request 1 should be allowed")
	}

	time.Sleep(30 * time.Millisecond)

	if !rl.Allow("client-1") {
		t.Fatal("request 2 should be allowed")
	}

	// First request may expire while second request is still inside
	// the window.
	time.Sleep(30 * time.Millisecond)

	if !rl.Allow("client-1") {
		t.Fatal("request should be allowed because the first request expired")
	}
}

func TestRateLimiter_ClientCreatedAutomatically(t *testing.T) {
	rl := NewRateLimiter(5, time.Second)

	if _, ok := rl.clients["unknown"]; ok {
		t.Fatal("client should not exist before first request")
	}

	if !rl.Allow("unknown") {
		t.Fatal("first request should be allowed")
	}

	if _, ok := rl.clients["unknown"]; !ok {
		t.Fatal("client should have been created")
	}
}

func TestRateLimiter_EmptyClientID(t *testing.T) {
	rl := NewRateLimiter(2, time.Second)

	if !rl.Allow("") {
		t.Fatal("empty client ID should still be treated as a client")
	}

	if !rl.Allow("") {
		t.Fatal("second request should be allowed")
	}

	if rl.Allow("") {
		t.Fatal("third request should be rejected")
	}
}

func TestRateLimiter_ZeroMaxRequests(t *testing.T) {
	rl := NewRateLimiter(0, time.Second)

	if rl.Allow("client-1") {
		t.Fatal("no requests should be allowed when maxRequests is zero")
	}
}

func TestRateLimiter_ZeroWindow(t *testing.T) {
	rl := NewRateLimiter(1, 0)

	// With a zero window, requests should immediately fall outside
	// the window once they have a measurable age.
	rl.Allow("client-1")

	time.Sleep(time.Millisecond)

	if !rl.Allow("client-1") {
		t.Fatal("request should be allowed after a zero-length window")
	}
}

func TestRateLimiter_ManyClients(t *testing.T) {
	rl := NewRateLimiter(2, time.Second)

	const clients = 100

	for i := 0; i < clients; i++ {
		clientID := "client-" + string(rune(i))

		if !rl.Allow(clientID) {
			t.Fatalf("first request for %s should be allowed", clientID)
		}
	}

	if len(rl.clients) != clients {
		t.Fatalf(
			"expected %d clients, got %d",
			clients,
			len(rl.clients),
		)
	}
}

func TestRateLimiter_ConcurrentSameClient(t *testing.T) {
	const (
		maxRequests = 10
		goroutines  = 100
	)

	rl := NewRateLimiter(maxRequests, time.Second)

	var wg sync.WaitGroup

	var allowed int
	var allowedMu sync.Mutex

	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()

			if rl.Allow("client-1") {
				allowedMu.Lock()
				allowed++
				allowedMu.Unlock()
			}
		}()
	}

	wg.Wait()

	if allowed != maxRequests {
		t.Fatalf(
			"expected exactly %d allowed requests, got %d",
			maxRequests,
			allowed,
		)
	}
}

func TestRateLimiter_ConcurrentDifferentClients(t *testing.T) {
	const (
		maxRequests = 5
		clients     = 20
	)

	rl := NewRateLimiter(maxRequests, time.Second)

	var wg sync.WaitGroup

	for i := 0; i < clients; i++ {
		clientID := "client-" + string(rune(i))

		wg.Add(1)

		go func(id string) {
			defer wg.Done()

			for j := 0; j < maxRequests; j++ {
				if !rl.Allow(id) {
					t.Errorf(
						"request %d for %s should be allowed",
						j+1,
						id,
					)
				}
			}
		}(clientID)
	}

	wg.Wait()
}

func TestRateLimiter_ConcurrentSameClientRepeatedly(t *testing.T) {
	rl := NewRateLimiter(100, time.Second)

	const (
		goroutines    = 100
		requestsPerGo = 100
	)

	var wg sync.WaitGroup

	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()

			for j := 0; j < requestsPerGo; j++ {
				rl.Allow("client-1")
			}
		}()
	}

	wg.Wait()
}

func TestRateLimiter_ConcurrentClientCreation(t *testing.T) {
	rl := NewRateLimiter(5, time.Second)

	const goroutines = 100

	var wg sync.WaitGroup

	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()

			rl.Allow("same-client")
		}()
	}

	wg.Wait()

	rl.mapMU.Lock()
	defer rl.mapMU.Unlock()

	if len(rl.clients) != 1 {
		t.Fatalf(
			"expected exactly 1 client, got %d",
			len(rl.clients),
		)
	}
}
