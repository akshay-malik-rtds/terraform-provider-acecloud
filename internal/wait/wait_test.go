package wait

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// toSet
// ---------------------------------------------------------------------------

func TestToSet_Empty(t *testing.T) {
	s := toSet(nil)
	if len(s) != 0 {
		t.Fatalf("expected empty set, got %d items", len(s))
	}
}

func TestToSet_Single(t *testing.T) {
	s := toSet([]string{"ACTIVE"})
	if _, ok := s["ACTIVE"]; !ok {
		t.Fatal("expected ACTIVE in set")
	}
	if len(s) != 1 {
		t.Fatalf("expected 1 item, got %d", len(s))
	}
}

func TestToSet_Multiple(t *testing.T) {
	s := toSet([]string{"ACTIVE", "ERROR", "ACTIVE"}) // duplicate
	if len(s) != 2 {
		t.Fatalf("expected 2 items (deduplicated), got %d", len(s))
	}
}

// ---------------------------------------------------------------------------
// isRetryable
// ---------------------------------------------------------------------------

func TestIsRetryable_Match(t *testing.T) {
	err := fmt.Errorf("resource is immutable while in PENDING_CREATE")
	if !isRetryable(err, defaultRetryableErrors) {
		t.Fatal("expected retryable for 'immutable'")
	}
}

func TestIsRetryable_MatchPending(t *testing.T) {
	err := fmt.Errorf("status is PENDING_DELETE")
	if !isRetryable(err, defaultRetryableErrors) {
		t.Fatal("expected retryable for 'PENDING_'")
	}
}

func TestIsRetryable_MatchInUse(t *testing.T) {
	err := fmt.Errorf("port is still in use by device")
	if !isRetryable(err, defaultRetryableErrors) {
		t.Fatal("expected retryable for 'in use'")
	}
}

func TestIsRetryable_MatchSecurityGroup(t *testing.T) {
	err := fmt.Errorf("SecurityGroupInUse: still attached")
	if !isRetryable(err, defaultRetryableErrors) {
		t.Fatal("expected retryable for 'SecurityGroupInUse'")
	}
}

func TestIsRetryable_NoMatch(t *testing.T) {
	err := fmt.Errorf("not found")
	if isRetryable(err, defaultRetryableErrors) {
		t.Fatal("expected not retryable for 'not found'")
	}
}

func TestIsRetryable_CustomSubstrings(t *testing.T) {
	err := fmt.Errorf("volume is locked for backup")
	custom := []string{"locked", "busy"}
	if !isRetryable(err, custom) {
		t.Fatal("expected retryable with custom substrings")
	}
}

func TestIsRetryable_EmptyRetryableList(t *testing.T) {
	err := fmt.Errorf("immutable")
	if isRetryable(err, []string{}) {
		t.Fatal("expected not retryable with empty list")
	}
}

// ---------------------------------------------------------------------------
// WaitForStatus
// ---------------------------------------------------------------------------

func TestWaitForStatus_ImmediateTarget(t *testing.T) {
	result, err := WaitForStatus(context.Background(), WaitForStatusOpts{
		Refresh: func(ctx context.Context) (*StatusResult, error) {
			return &StatusResult{Status: "ACTIVE", Data: "mydata"}, nil
		},
		TargetStatus: []string{"ACTIVE"},
		ErrorStatus:  []string{"ERROR"},
		Timeout:      2 * time.Second,
		PollInterval: 100 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != "ACTIVE" {
		t.Fatalf("expected ACTIVE, got %s", result.Status)
	}
	if result.Data.(string) != "mydata" {
		t.Fatalf("expected mydata, got %v", result.Data)
	}
}

func TestWaitForStatus_ReachesTargetAfterPolls(t *testing.T) {
	var callCount int32
	result, err := WaitForStatus(context.Background(), WaitForStatusOpts{
		Refresh: func(ctx context.Context) (*StatusResult, error) {
			n := atomic.AddInt32(&callCount, 1)
			if n < 3 {
				return &StatusResult{Status: "BUILD"}, nil
			}
			return &StatusResult{Status: "ACTIVE", Data: "done"}, nil
		},
		TargetStatus: []string{"ACTIVE"},
		ErrorStatus:  []string{"ERROR"},
		Timeout:      5 * time.Second,
		PollInterval: 50 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != "ACTIVE" {
		t.Fatalf("expected ACTIVE, got %s", result.Status)
	}
	if atomic.LoadInt32(&callCount) < 3 {
		t.Fatalf("expected at least 3 calls, got %d", callCount)
	}
}

func TestWaitForStatus_ErrorStatus(t *testing.T) {
	result, err := WaitForStatus(context.Background(), WaitForStatusOpts{
		Refresh: func(ctx context.Context) (*StatusResult, error) {
			return &StatusResult{Status: "ERROR", Data: "fail"}, nil
		},
		TargetStatus: []string{"ACTIVE"},
		ErrorStatus:  []string{"ERROR"},
		Timeout:      2 * time.Second,
		PollInterval: 50 * time.Millisecond,
	})
	if err == nil {
		t.Fatal("expected error for ERROR status")
	}
	if result == nil || result.Status != "ERROR" {
		t.Fatal("expected result with ERROR status")
	}
	expected := "resource entered error status: ERROR"
	if err.Error() != expected {
		t.Fatalf("expected error %q, got %q", expected, err.Error())
	}
}

func TestWaitForStatus_Timeout(t *testing.T) {
	_, err := WaitForStatus(context.Background(), WaitForStatusOpts{
		Refresh: func(ctx context.Context) (*StatusResult, error) {
			return &StatusResult{Status: "BUILD"}, nil
		},
		TargetStatus: []string{"ACTIVE"},
		ErrorStatus:  []string{"ERROR"},
		Timeout:      200 * time.Millisecond,
		PollInterval: 50 * time.Millisecond,
	})
	if err == nil {
		t.Fatal("expected timeout error")
	}
	if !containsStr(err.Error(), "timed out") {
		t.Fatalf("expected timeout message, got: %s", err.Error())
	}
	if !containsStr(err.Error(), "BUILD") {
		t.Fatalf("expected last status BUILD in message, got: %s", err.Error())
	}
}

func TestWaitForStatus_ContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	_, err := WaitForStatus(ctx, WaitForStatusOpts{
		Refresh: func(ctx context.Context) (*StatusResult, error) {
			return &StatusResult{Status: "BUILD"}, nil
		},
		TargetStatus: []string{"ACTIVE"},
		ErrorStatus:  []string{"ERROR"},
		Timeout:      5 * time.Second,
		PollInterval: 50 * time.Millisecond,
	})
	if err == nil {
		t.Fatal("expected context cancelled error")
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got: %v", err)
	}
}

func TestWaitForStatus_TransientRefreshErrors(t *testing.T) {
	var callCount int32
	result, err := WaitForStatus(context.Background(), WaitForStatusOpts{
		Refresh: func(ctx context.Context) (*StatusResult, error) {
			n := atomic.AddInt32(&callCount, 1)
			if n <= 2 {
				return nil, fmt.Errorf("network blip")
			}
			return &StatusResult{Status: "ACTIVE"}, nil
		},
		TargetStatus: []string{"ACTIVE"},
		ErrorStatus:  []string{"ERROR"},
		Timeout:      5 * time.Second,
		PollInterval: 50 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != "ACTIVE" {
		t.Fatalf("expected ACTIVE, got %s", result.Status)
	}
}

func TestWaitForStatus_MultipleTargetStatuses(t *testing.T) {
	result, err := WaitForStatus(context.Background(), WaitForStatusOpts{
		Refresh: func(ctx context.Context) (*StatusResult, error) {
			return &StatusResult{Status: "DEGRADED"}, nil
		},
		TargetStatus: []string{"ACTIVE", "DEGRADED"},
		ErrorStatus:  []string{"ERROR"},
		Timeout:      2 * time.Second,
		PollInterval: 50 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != "DEGRADED" {
		t.Fatalf("expected DEGRADED, got %s", result.Status)
	}
}

func TestWaitForStatus_MultipleErrorStatuses(t *testing.T) {
	_, err := WaitForStatus(context.Background(), WaitForStatusOpts{
		Refresh: func(ctx context.Context) (*StatusResult, error) {
			return &StatusResult{Status: "FAILED"}, nil
		},
		TargetStatus: []string{"ACTIVE"},
		ErrorStatus:  []string{"ERROR", "FAILED"},
		Timeout:      2 * time.Second,
		PollInterval: 50 * time.Millisecond,
	})
	if err == nil {
		t.Fatal("expected error for FAILED status")
	}
	if !containsStr(err.Error(), "FAILED") {
		t.Fatalf("expected FAILED in error, got: %s", err.Error())
	}
}

func TestWaitForStatus_DefaultTimeoutApplied(t *testing.T) {
	// Verify that with zero Timeout, default is applied (not instant timeout).
	result, err := WaitForStatus(context.Background(), WaitForStatusOpts{
		Refresh: func(ctx context.Context) (*StatusResult, error) {
			return &StatusResult{Status: "ACTIVE"}, nil
		},
		TargetStatus: []string{"ACTIVE"},
		// Timeout:0 => DefaultStatusTimeout
		PollInterval: 50 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != "ACTIVE" {
		t.Fatalf("expected ACTIVE, got %s", result.Status)
	}
}

func TestWaitForStatus_NilResult(t *testing.T) {
	var callCount int32
	result, err := WaitForStatus(context.Background(), WaitForStatusOpts{
		Refresh: func(ctx context.Context) (*StatusResult, error) {
			n := atomic.AddInt32(&callCount, 1)
			if n < 3 {
				return nil, nil // nil result, nil error
			}
			return &StatusResult{Status: "ACTIVE"}, nil
		},
		TargetStatus: []string{"ACTIVE"},
		Timeout:      5 * time.Second,
		PollInterval: 50 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != "ACTIVE" {
		t.Fatalf("expected ACTIVE, got %s", result.Status)
	}
}

// ---------------------------------------------------------------------------
// RetryOnConflict
// ---------------------------------------------------------------------------

func TestRetryOnConflict_ImmediateSuccess(t *testing.T) {
	err := RetryOnConflict(context.Background(), RetryOnConflictOpts{
		Operation: func(ctx context.Context) error {
			return nil
		},
		Timeout:  2 * time.Second,
		Interval: 50 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRetryOnConflict_SucceedsAfterRetries(t *testing.T) {
	var callCount int32
	err := RetryOnConflict(context.Background(), RetryOnConflictOpts{
		Operation: func(ctx context.Context) error {
			n := atomic.AddInt32(&callCount, 1)
			if n < 3 {
				return fmt.Errorf("resource is immutable while PENDING_CREATE")
			}
			return nil
		},
		Timeout:  5 * time.Second,
		Interval: 50 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if atomic.LoadInt32(&callCount) < 3 {
		t.Fatal("expected at least 3 attempts")
	}
}

func TestRetryOnConflict_NonRetryableError(t *testing.T) {
	err := RetryOnConflict(context.Background(), RetryOnConflictOpts{
		Operation: func(ctx context.Context) error {
			return fmt.Errorf("resource not found")
		},
		Timeout:  2 * time.Second,
		Interval: 50 * time.Millisecond,
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "resource not found" {
		t.Fatalf("expected original error, got: %v", err)
	}
}

func TestRetryOnConflict_Timeout(t *testing.T) {
	err := RetryOnConflict(context.Background(), RetryOnConflictOpts{
		Operation: func(ctx context.Context) error {
			return fmt.Errorf("resource is immutable")
		},
		Timeout:  200 * time.Millisecond,
		Interval: 50 * time.Millisecond,
	})
	if err == nil {
		t.Fatal("expected timeout error")
	}
	if !containsStr(err.Error(), "timed out") {
		t.Fatalf("expected timeout message, got: %s", err.Error())
	}
}

func TestRetryOnConflict_ContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	var callCount int32
	err := RetryOnConflict(ctx, RetryOnConflictOpts{
		Operation: func(ctx context.Context) error {
			n := atomic.AddInt32(&callCount, 1)
			if n == 2 {
				cancel()
			}
			return fmt.Errorf("resource is immutable")
		},
		Timeout:  5 * time.Second,
		Interval: 50 * time.Millisecond,
	})
	if err == nil {
		t.Fatal("expected error on context cancel")
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got: %v", err)
	}
}

func TestRetryOnConflict_CustomRetryableErrors(t *testing.T) {
	var callCount int32
	err := RetryOnConflict(context.Background(), RetryOnConflictOpts{
		Operation: func(ctx context.Context) error {
			n := atomic.AddInt32(&callCount, 1)
			if n < 3 {
				return fmt.Errorf("volume is locked for backup")
			}
			return nil
		},
		RetryableErrors: []string{"locked", "busy"},
		Timeout:         5 * time.Second,
		Interval:        50 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRetryOnConflict_InUseRetry(t *testing.T) {
	var callCount int32
	err := RetryOnConflict(context.Background(), RetryOnConflictOpts{
		Operation: func(ctx context.Context) error {
			n := atomic.AddInt32(&callCount, 1)
			if n < 2 {
				return fmt.Errorf("port still in use by device")
			}
			return nil
		},
		Timeout:  5 * time.Second,
		Interval: 50 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRetryOnConflict_SecurityGroupInUseRetry(t *testing.T) {
	var callCount int32
	err := RetryOnConflict(context.Background(), RetryOnConflictOpts{
		Operation: func(ctx context.Context) error {
			n := atomic.AddInt32(&callCount, 1)
			if n < 2 {
				return fmt.Errorf("SecurityGroupInUse: still attached")
			}
			return nil
		},
		Timeout:  5 * time.Second,
		Interval: 50 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRetryOnConflict_PleaseRetryAgain(t *testing.T) {
	var callCount int32
	err := RetryOnConflict(context.Background(), RetryOnConflictOpts{
		Operation: func(ctx context.Context) error {
			n := atomic.AddInt32(&callCount, 1)
			if n < 2 {
				return fmt.Errorf("Please try again later")
			}
			return nil
		},
		Timeout:  5 * time.Second,
		Interval: 50 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// PollForResource
// ---------------------------------------------------------------------------

func TestPollForResource_FoundImmediately(t *testing.T) {
	item, err := PollForResource(context.Background(), PollForResourceOpts{
		List: func(ctx context.Context) (interface{}, error) {
			return "found-item", nil
		},
		Timeout:      2 * time.Second,
		PollInterval: 50 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.(string) != "found-item" {
		t.Fatalf("expected found-item, got %v", item)
	}
}

func TestPollForResource_FoundAfterPolls(t *testing.T) {
	var callCount int32
	item, err := PollForResource(context.Background(), PollForResourceOpts{
		List: func(ctx context.Context) (interface{}, error) {
			n := atomic.AddInt32(&callCount, 1)
			if n < 3 {
				return nil, nil
			}
			return "found-item", nil
		},
		Timeout:      5 * time.Second,
		PollInterval: 50 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.(string) != "found-item" {
		t.Fatalf("expected found-item, got %v", item)
	}
}

func TestPollForResource_Timeout(t *testing.T) {
	_, err := PollForResource(context.Background(), PollForResourceOpts{
		List: func(ctx context.Context) (interface{}, error) {
			return nil, nil
		},
		Timeout:      200 * time.Millisecond,
		PollInterval: 50 * time.Millisecond,
	})
	if err == nil {
		t.Fatal("expected timeout error")
	}
	if !containsStr(err.Error(), "timed out") {
		t.Fatalf("expected timeout message, got: %s", err.Error())
	}
}

func TestPollForResource_ContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := PollForResource(ctx, PollForResourceOpts{
		List: func(ctx context.Context) (interface{}, error) {
			return nil, nil
		},
		Timeout:      5 * time.Second,
		PollInterval: 50 * time.Millisecond,
	})
	if err == nil {
		t.Fatal("expected context cancelled error")
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got: %v", err)
	}
}

func TestPollForResource_TransientErrors(t *testing.T) {
	var callCount int32
	item, err := PollForResource(context.Background(), PollForResourceOpts{
		List: func(ctx context.Context) (interface{}, error) {
			n := atomic.AddInt32(&callCount, 1)
			if n <= 2 {
				return nil, fmt.Errorf("network error")
			}
			return "found", nil
		},
		Timeout:      5 * time.Second,
		PollInterval: 50 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.(string) != "found" {
		t.Fatalf("expected found, got %v", item)
	}
}

func TestPollForResource_ReturnsStruct(t *testing.T) {
	type cluster struct {
		ID   string
		Name string
	}
	item, err := PollForResource(context.Background(), PollForResourceOpts{
		List: func(ctx context.Context) (interface{}, error) {
			return &cluster{ID: "c-123", Name: "test"}, nil
		},
		Timeout:      2 * time.Second,
		PollInterval: 50 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	c := item.(*cluster)
	if c.ID != "c-123" || c.Name != "test" {
		t.Fatalf("unexpected cluster: %+v", c)
	}
}

// ---------------------------------------------------------------------------
// Defaults
// ---------------------------------------------------------------------------

func TestDefaultConstants(t *testing.T) {
	if DefaultStatusTimeout != 5*time.Minute {
		t.Fatalf("expected 5m, got %v", DefaultStatusTimeout)
	}
	if DefaultStatusInterval != 5*time.Second {
		t.Fatalf("expected 5s, got %v", DefaultStatusInterval)
	}
	if DefaultRetryTimeout != 2*time.Minute {
		t.Fatalf("expected 2m, got %v", DefaultRetryTimeout)
	}
	if DefaultRetryInterval != 10*time.Second {
		t.Fatalf("expected 10s, got %v", DefaultRetryInterval)
	}
	if DefaultPollTimeout != 2*time.Minute {
		t.Fatalf("expected 2m, got %v", DefaultPollTimeout)
	}
	if DefaultPollInterval != 5*time.Second {
		t.Fatalf("expected 5s, got %v", DefaultPollInterval)
	}
}

func TestDefaultRetryableErrors(t *testing.T) {
	expected := []string{"immutable", "PENDING_", "Please try again", "in use", "SecurityGroupInUse"}
	if len(defaultRetryableErrors) != len(expected) {
		t.Fatalf("expected %d defaults, got %d", len(expected), len(defaultRetryableErrors))
	}
	for i, e := range expected {
		if defaultRetryableErrors[i] != e {
			t.Fatalf("expected %q at index %d, got %q", e, i, defaultRetryableErrors[i])
		}
	}
}

// ---------------------------------------------------------------------------
// helper
// ---------------------------------------------------------------------------

func containsStr(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstr(s, substr))
}

func containsSubstr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
