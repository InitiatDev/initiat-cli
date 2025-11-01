package setup

import "time"

func ParseDuration(s string) (time.Duration, error) {
	return time.ParseDuration(s)
}

func ParseRetryPolicy(retries *Retries) *RetryPolicy {
	if retries == nil {
		return nil
	}

	backoff := time.Second * 1
	if retries.Backoff != "" {
		if parsed, err := ParseDuration(retries.Backoff); err == nil {
			backoff = parsed
		}
	}

	return &RetryPolicy{
		Attempts: retries.Attempts,
		Backoff:  backoff,
	}
}
