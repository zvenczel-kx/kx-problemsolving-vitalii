package services

import (
	"context"
	"errors"
	"gateway/internal/registry"
	"log"
	"net/http"
	"time"
)

type Gateway struct {
	Client      *http.Client
	Registry    *registry.ServiceRegistry
	MaxRetries  int
	Timeout     time.Duration
	BackoffBase time.Duration
}

func (f *Gateway) isRetryableStatus(code int) bool {
	switch code {
	case 500, 502, 503, 504:
		return true
	default:
		return false
	}
}

func (f *Gateway) Forward(ctx context.Context) (*http.Response, error) {
	if f.Client == nil || f.Registry == nil {
		return nil, errors.New("gateway not configured")
	}

	ctx, cancel := context.WithTimeout(ctx, f.Timeout)
	defer cancel()

	target := f.Registry.NextInstance()
	if target == "" {
		return nil, errors.New("no available services")
	}

	backoff := f.BackoffBase
	var resp *http.Response
	var err error

	for attempt := 0; attempt <= f.MaxRetries; attempt++ {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		req, reqErr := http.NewRequestWithContext(ctx, http.MethodGet, target+"/data", nil)
		if reqErr != nil {
			return nil, reqErr
		}

		resp, err = f.Client.Do(req)
		if err != nil {
			// network error; retry if attempts left
			if attempt < f.MaxRetries {
				log.Printf("[forward] network error attempt=%d: %v", attempt, err)
				select {
				case <-time.After(backoff):
					backoff = time.Duration(float64(backoff) * 1.7)
				case <-ctx.Done():
					return nil, ctx.Err()
				}
				continue
			}
			return nil, err
		}

		if f.isRetryableStatus(resp.StatusCode) {
			if attempt < f.MaxRetries {
				log.Printf("[forward] retryable status=%d attempt=%d", resp.StatusCode, attempt)
				resp.Body.Close()
				select {
				case <-time.After(backoff):
					backoff = time.Duration(float64(backoff) * 1.7)
				case <-ctx.Done():
					return nil, ctx.Err()
				}
				continue
			}
		}

		break
	}

	return resp, nil
}
