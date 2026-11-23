package service

import (
	"errors"
	"math/big"
	"testing"
	"time"
)

func TestRetryBlockReadStopsAtFirstSuccess(t *testing.T) {
	oldDelay := blockRetryDelay
	blockRetryDelay = time.Millisecond
	defer func() { blockRetryDelay = oldDelay }()

	calls := 0
	err := retryBlockRead(func() error {
		calls++
		if calls < 3 {
			return errBlockNotFound
		}
		return nil
	})
	if err != nil {
		t.Fatalf("expected success on the third call, got %v", err)
	}
	if calls != 3 {
		t.Fatalf("expected 3 calls, got %d", calls)
	}
}

func TestRetryBlockReadGivesLastErrorAfterAllAttempts(t *testing.T) {
	oldDelay := blockRetryDelay
	blockRetryDelay = time.Millisecond
	defer func() { blockRetryDelay = oldDelay }()

	calls := 0
	want := errors.New("RPC error: header not found")
	err := retryBlockRead(func() error {
		calls++
		return want
	})
	if !errors.Is(err, want) {
		t.Fatalf("expected the last error, got %v", err)
	}
	if calls != blockRetryAttempts {
		t.Fatalf("expected %d calls, got %d", blockRetryAttempts, calls)
	}
}

func TestLastGoodSupply(t *testing.T) {
	readErr := errors.New("RPC error: header not found")
	cached := big.NewFloat(361720565.97)

	// No cached value yet: the error goes to the caller.
	if got, err := lastGoodSupply("total", nil, time.Now(), readErr); got != nil || !errors.Is(err, readErr) {
		t.Fatalf("expected the read error with no cache, got %v, %v", got, err)
	}

	// A recent cached value replaces the failed read.
	got, err := lastGoodSupply("total", cached, time.Now().Add(-time.Minute), readErr)
	if err != nil || got == nil || got.Cmp(cached) != 0 {
		t.Fatalf("expected the cached value, got %v, %v", got, err)
	}

	// A cached value older than the limit does not hide the failure.
	old := time.Now().Add(-staleSupplyLimit - time.Second)
	if got, err := lastGoodSupply("total", cached, old, readErr); got != nil || !errors.Is(err, readErr) {
		t.Fatalf("expected the read error with a stale cache, got %v, %v", got, err)
	}
}
