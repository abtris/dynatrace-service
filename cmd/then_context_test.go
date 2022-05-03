package main

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const allowedDeltaSeconds = 0.5

// TestNewThenWithTimeoutContext_Simple tests that a ThenWithTimeout context is done timeoutSeconds after being the waitCtx is done.
func TestNewThenWithTimeoutContext_Simple(t *testing.T) {
	const timeoutSeconds = 4

	initialCtx, trigger := context.WithCancel(context.Background())

	thenWithTimeoutCtx, _ := newThenWithTimeoutContext(context.Background(), initialCtx, timeoutSeconds*time.Second, "thenWithTimeoutCtx1")

	startTime := time.Now()
	trigger()
	<-thenWithTimeoutCtx.Done()
	duration := time.Now().Sub(startTime)

	assert.InDelta(t, timeoutSeconds, duration/time.Second, allowedDeltaSeconds)
}

// TestNewThenWithTimeoutContext_Chained tests that two ThenWithTimeout contexts can be chained together with the expected total timeout.
func TestNewThenWithTimeoutContext_Chained(t *testing.T) {
	const timeoutSeconds1 = 4
	const timeoutSeconds2 = 2

	initialCtx, trigger := context.WithCancel(context.Background())

	thenWithTimeoutCtx1, _ := newThenWithTimeoutContext(context.Background(), initialCtx, timeoutSeconds1*time.Second, "thenWithTimeoutCtx1")
	thenWithTimeoutCtx2, _ := newThenWithTimeoutContext(context.Background(), thenWithTimeoutCtx1, timeoutSeconds2*time.Second, "thenWithTimeoutCtx2")

	startTime := time.Now()
	trigger()
	<-thenWithTimeoutCtx2.Done()
	duration := time.Now().Sub(startTime)

	assert.InDelta(t, timeoutSeconds1+timeoutSeconds2, duration/time.Second, allowedDeltaSeconds)
}

// TestNewThenWithTimeoutContext_Early tests cancelling the first of two chained ThenWithTimeout contexts causes its timeout to be skipped.
func TestNewThenWithTimeoutContext_Early(t *testing.T) {
	const timeoutSeconds1 = 3
	const timeoutSeconds2 = 1

	initialCtx, trigger := context.WithCancel(context.Background())

	thenWithTimeoutCtx1, cancel1 := newThenWithTimeoutContext(context.Background(), initialCtx, timeoutSeconds1*time.Second, "thenWithTimeoutCtx1")
	thenWithTimeoutCtx2, _ := newThenWithTimeoutContext(context.Background(), thenWithTimeoutCtx1, timeoutSeconds2*time.Second, "thenWithTimeoutCtx2")

	startTime := time.Now()
	trigger()
	cancel1()
	<-thenWithTimeoutCtx2.Done()
	duration := time.Now().Sub(startTime)

	assert.InDelta(t, timeoutSeconds2, duration/time.Second, allowedDeltaSeconds)
}
