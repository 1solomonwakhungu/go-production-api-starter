package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"testing"
	"time"
)

func TestSecondSignalTerminatesDuringGracefulShutdown(t *testing.T) {
	if os.Getenv("GO_TEST_SIGNAL_HELPER") == "1" {
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()

		fmt.Println("ready")
		<-ctx.Done()
		_ = gracefullyShutdown(func(context.Context) error {
			fmt.Println("shutting-down")
			select {}
		}, stop)
	}

	cmd := exec.Command(os.Args[0], "-test.run=^TestSecondSignalTerminatesDuringGracefulShutdown$")
	cmd.Env = append(os.Environ(), "GO_TEST_SIGNAL_HELPER=1")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatalf("create stdout pipe: %v", err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatalf("start helper process: %v", err)
	}

	wait := make(chan error, 1)
	go func() {
		wait <- cmd.Wait()
	}()

	killHelper := func() {
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
		<-wait
	}

	scanner := bufio.NewScanner(stdout)
	if !scanner.Scan() || scanner.Text() != "ready" {
		killHelper()
		t.Fatalf("helper did not become ready: %v", scanner.Err())
	}
	if err := cmd.Process.Signal(syscall.SIGTERM); err != nil {
		killHelper()
		t.Fatalf("send first SIGTERM: %v", err)
	}
	if !scanner.Scan() || scanner.Text() != "shutting-down" {
		killHelper()
		t.Fatalf("helper did not begin graceful shutdown: %v", scanner.Err())
	}
	if err := cmd.Process.Signal(syscall.SIGTERM); err != nil {
		killHelper()
		t.Fatalf("send second SIGTERM: %v", err)
	}

	select {
	case err := <-wait:
		exitErr, ok := err.(*exec.ExitError)
		if !ok {
			t.Fatalf("helper exited without a signal: %v", err)
		}
		status, ok := exitErr.Sys().(syscall.WaitStatus)
		if !ok || !status.Signaled() || status.Signal() != syscall.SIGTERM {
			t.Fatalf("helper exit status = %v, want SIGTERM", exitErr.Sys())
		}
	case <-time.After(2 * time.Second):
		killHelper()
		t.Fatal("second SIGTERM did not terminate blocked graceful shutdown")
	}
}
