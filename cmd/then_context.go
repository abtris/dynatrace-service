package main

import (
	"context"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// newThenWithTimeoutContext creates a new Context that is done:
// a. after a timeout which is triggered after waitCtx is done, or
// b.  if parentCtx is done.
// It returns the new context and a function that cancels the context.
func newThenWithTimeoutContext(parentCtx context.Context, waitCtx context.Context, timeout time.Duration, name string) (context.Context, func()) {
	logrus.WithField("name", name).WithField("timeout", timeout).Debug("Creating ThenWithTimeoutContext")
	thenCtx, innerCancel := context.WithCancel(parentCtx)

	// start a monitoring goroutine to ultimately set the thenCtx done state
	waitGroup := &sync.WaitGroup{}
	waitGroup.Add(1)
	go func() {
		defer waitGroup.Done()
		defer innerCancel()

		// wait for waitCtx or this thenContext
		select {
		case <-waitCtx.Done():

		case <-thenCtx.Done():
			return
		}

		logrus.WithField("name", name).Debug("Timeout of ThenWithTimeoutContext triggered")

		// wait for the timeout or this thenContext to be cancelled
		select {
		case <-thenCtx.Done():
			return
		case <-time.After(timeout):

		}

		logrus.WithField("name", name).Debug("ThenWithTimeoutContext expired")

	}()

	cancel := func() {
		logrus.WithField("name", name).Debug("Cancelling ThenWithTimeoutContext")
		innerCancel()
		waitGroup.Wait()
	}

	return thenCtx, cancel
}
