package main

import (
	"flag"
	"os"
	"testing"
)

func TestParseFlags(t *testing.T) {
	// Save original os.Args
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	testCases := []struct {
		name     string
		args     []string
		wantPort int
		wantDB   string
		wantWeb  string
	}{
		{
			name:     "default values",
			args:     []string{"cmd"},
			wantPort: 8080,
			wantDB:   "./bookshelf.db",
			wantWeb:  "./web",
		},
		{
			name:     "custom port",
			args:     []string{"cmd", "--port", "9090"},
			wantPort: 9090,
			wantDB:   "./bookshelf.db",
			wantWeb:  "./web",
		},
		{
			name:     "custom db file",
			args:     []string{"cmd", "--db-file", "/tmp/test.db"},
			wantPort: 8080,
			wantDB:   "/tmp/test.db",
			wantWeb:  "./web",
		},
		{
			name:     "custom web dir",
			args:     []string{"cmd", "--web-dir", "/tmp/web"},
			wantPort: 8080,
			wantDB:   "./bookshelf.db",
			wantWeb:  "/tmp/web",
		},
		{
			name:     "all custom values",
			args:     []string{"cmd", "--port", "9090", "--db-file", "/tmp/test.db", "--web-dir", "/tmp/web"},
			wantPort: 9090,
			wantDB:   "/tmp/test.db",
			wantWeb:  "/tmp/web",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Set command-line arguments
			os.Args = tc.args

			// Reset flags
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

			// Parse flags
			port := flag.Int("port", 8080, "Port number for the HTTP server")
			dbFile := flag.String("db-file", "./bookshelf.db", "Path to the SQLite database file")
			webDir := flag.String("web-dir", "./web", "Directory containing static web assets")
			flag.Parse()

			// Check results
			if *port != tc.wantPort {
				t.Errorf("port = %d; want %d", *port, tc.wantPort)
			}
			if *dbFile != tc.wantDB {
				t.Errorf("dbFile = %s; want %s", *dbFile, tc.wantDB)
			}
			if *webDir != tc.wantWeb {
				t.Errorf("webDir = %s; want %s", *webDir, tc.wantWeb)
			}
		})
	}
}