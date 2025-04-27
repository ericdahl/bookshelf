package main

import (
	"os"
	"testing"
)

func TestCheckWebDir(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "bookshelf-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test valid directory
	err = checkWebDir(tempDir)
	if err != nil {
		t.Errorf("checkWebDir(%s) returned error for valid directory: %v", tempDir, err)
	}

	// Test non-existent directory
	nonExistentDir := tempDir + "/nonexistent"
	err = checkWebDir(nonExistentDir)
	if err == nil {
		t.Errorf("checkWebDir(%s) did not return error for non-existent directory", nonExistentDir)
	}
}