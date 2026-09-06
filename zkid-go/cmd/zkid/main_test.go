package main_test

import (
	"bytes"
	"os/exec"
	"strings"
	"testing"
)

func TestCliDecodeDateTimeFormat(t *testing.T) {
	cmd := exec.Command("go", "run", ".", "-decode")
	cmd.Stdin = strings.NewReader("268F6e\n")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("CLI execution failed: %v, stderr: %s", err, stderr.String())
	}

	expected := "2026-08-15 06:40:00\n"
	if stdout.String() != expected {
		t.Errorf("Expected output %q, got %q", expected, stdout.String())
	}
}

func TestCliDecodeMultipleLines(t *testing.T) {
	cmd := exec.Command("go", "run", ".", "-decode")
	cmd.Stdin = strings.NewReader("268F6e\n2695I7\n")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("CLI execution failed: %v, stderr: %s", err, stderr.String())
	}

	lines := strings.Split(strings.TrimSpace(stdout.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("Expected 2 lines, got %d: %q", len(lines), stdout.String())
	}
	if lines[0] != "2026-08-15 06:40:00" {
		t.Errorf("Line 1 expected '2026-08-15 06:40:00', got %q", lines[0])
	}
	if lines[1] != "2026-09-05 18:07:00" {
		t.Errorf("Line 2 expected '2026-09-05 18:07:00', got %q", lines[1])
	}
}

func TestCliEncodeDateTimeStdin(t *testing.T) {
	cmd := exec.Command("go", "run", ".")
	cmd.Stdin = strings.NewReader("2026-08-15 06:40:00\n")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("CLI execution failed: %v, stderr: %s", err, stderr.String())
	}

	expected := "268F6e\n"
	if stdout.String() != expected {
		t.Errorf("Expected output %q, got %q", expected, stdout.String())
	}
}

func TestCliEncodeDateTimeStdinMultipleLines(t *testing.T) {
	cmd := exec.Command("go", "run", ".")
	cmd.Stdin = strings.NewReader("2026-08-15 06:40:00\n2026-09-05 18:07:00\n")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("CLI execution failed: %v, stderr: %s", err, stderr.String())
	}

	lines := strings.Split(strings.TrimSpace(stdout.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("Expected 2 lines, got %d: %q", len(lines), stdout.String())
	}
	if lines[0] != "268F6e" {
		t.Errorf("Line 1 expected '268F6e', got %q", lines[0])
	}
	if lines[1] != "2695I7" {
		t.Errorf("Line 2 expected '2695I7', got %q", lines[1])
	}
}

func TestCliEncodeDateTimeStdinWithPrecision(t *testing.T) {
	cmd := exec.Command("go", "run", ".", "-p", "1")
	cmd.Stdin = strings.NewReader("2026-08-15 06:40:08\n")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("CLI execution failed: %v, stderr: %s", err, stderr.String())
	}

	expected := "268F6e-8\n"
	if stdout.String() != expected {
		t.Errorf("Expected output %q, got %q", expected, stdout.String())
	}
}

func TestCliEncodeInvalidDateTime(t *testing.T) {
	cmd := exec.Command("go", "run", ".")
	cmd.Stdin = strings.NewReader("not-a-datetime\n")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err == nil {
		t.Fatal("Expected CLI error for invalid datetime, got nil")
	}
	if !strings.Contains(stderr.String(), "error parsing datetime") {
		t.Errorf("Expected stderr to mention error parsing datetime, got: %s", stderr.String())
	}
}
