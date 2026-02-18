package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"testing"
	"time"

	pb "github.com/CPNext-hub/calendar-reg-main-api/proto/gen/coursepb"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

func TestFetchByCode_Success(t *testing.T) {
	// Setup faster sleep
	originalSleep := sleepDuration
	sleepDuration = 1 * time.Millisecond
	defer func() { sleepDuration = originalSleep }()

	s := &server{}
	req := &pb.FetchByCodeRequest{Code: "CP353004"}
	res, err := s.FetchByCode(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, "CP353004", res.Code)
}

func TestFetchByCode_NotFound(t *testing.T) {
	originalSleep := sleepDuration
	sleepDuration = 1 * time.Millisecond
	defer func() { sleepDuration = originalSleep }()

	s := &server{}
	req := &pb.FetchByCodeRequest{Code: "INVALID"}
	res, err := s.FetchByCode(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, res)
	st, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.NotFound, st.Code())
}

func TestRun(t *testing.T) {
	// Test that run starts the server and stops on context cancel
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ready := make(chan string, 1)
	errCh := make(chan error, 1)
	go func() {
		errCh <- run(ctx, ":0", ready)
	}()

	// Wait for server to be ready
	var addr string
	select {
	case addr = <-ready:
	case err := <-errCh:
		t.Fatalf("run failed to start: %v", err)
	case <-time.After(1 * time.Second):
		t.Fatal("timed out waiting for server to start")
	}

	// Connect to verify it's running
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewCourseServiceClient(conn)
	// We need to set sleep duration short for this test too, as run uses the global variable
	originalSleep := sleepDuration
	sleepDuration = 1 * time.Millisecond
	defer func() { sleepDuration = originalSleep }()

	_, err = client.FetchByCode(context.Background(), &pb.FetchByCodeRequest{Code: "CP353004"})
	assert.NoError(t, err)

	// Cancel context to stop server
	cancel()

	// Wait for run to return
	select {
	case err := <-errCh:
		// Accept nil (clean shutdown) or error depending on grpc version/implementation of Stop() vs Serve()
		// Serve usually returns nil on GracefulStop? No, Serve returns nil if Stop is called.
		assert.NoError(t, err)
	case <-time.After(2 * time.Second):
		t.Error("server did not stop in time")
	}
}

func TestRun_PortInUse(t *testing.T) {
	// Start a listener on a random port
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Skipf("could not bind to random port: %v", err)
	}
	port := l.Addr().String()
	defer l.Close()

	// Try running the server on the same port
	err = run(context.Background(), port, nil)
	assert.Error(t, err)
}

func TestMain_Success(t *testing.T) {
	origRun := runFunc
	origFatal := logFatal
	defer func() {
		runFunc = origRun
		logFatal = origFatal
	}()

	runFunc = func(ctx context.Context, addr string, ready chan<- string) error {
		return nil
	}

	fatalCalled := false
	logFatal = func(format string, args ...interface{}) {
		fatalCalled = true
	}

	main()

	assert.False(t, fatalCalled, "logFatal should not be called on success")
}

func TestMain_Error(t *testing.T) {
	origRun := runFunc
	origFatal := logFatal
	defer func() {
		runFunc = origRun
		logFatal = origFatal
	}()

	runFunc = func(ctx context.Context, addr string, ready chan<- string) error {
		return errors.New("bind error")
	}

	var fatalMsg string
	logFatal = func(format string, args ...interface{}) {
		fatalMsg = fmt.Sprintf(format, args...)
	}

	main()

	assert.Contains(t, fatalMsg, "bind error")
}
