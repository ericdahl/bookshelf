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
		name          string
		args          []string
		wantPort      int
		wantDB        string
		wantWeb       string
		wantVerbose   bool
		wantLogFormat string
	}{
		{
			name:          "default values",
			args:          []string{"cmd"},
			wantPort:      8080,
			wantDB:        "./bookshelf.db",
			wantWeb:       "./web",
			wantVerbose:   false,
			wantLogFormat: "text",
		},
		{
			name:          "custom port",
			args:          []string{"cmd", "--port", "9090"},
			wantPort:      9090,
			wantDB:        "./bookshelf.db",
			wantWeb:       "./web",
			wantVerbose:   false,
			wantLogFormat: "text",
		},
		{
			name:          "custom db file",
			args:          []string{"cmd", "--db-file", "/tmp/test.db"},
			wantPort:      8080,
			wantDB:        "/tmp/test.db",
			wantWeb:       "./web",
			wantVerbose:   false,
			wantLogFormat: "text",
		},
		{
			name:          "custom web dir",
			args:          []string{"cmd", "--web-dir", "/tmp/web"},
			wantPort:      8080,
			wantDB:        "./bookshelf.db",
			wantWeb:       "/tmp/web",
			wantVerbose:   false,
			wantLogFormat: "text",
		},
		{
			name:          "verbose mode",
			args:          []string{"cmd", "--verbose"},
			wantPort:      8080,
			wantDB:        "./bookshelf.db",
			wantWeb:       "./web",
			wantVerbose:   true,
			wantLogFormat: "text",
		},
		{
			name:          "json log format",
			args:          []string{"cmd", "--log-format", "json"},
			wantPort:      8080,
			wantDB:        "./bookshelf.db",
			wantWeb:       "./web",
			wantVerbose:   false,
			wantLogFormat: "json",
		},
		{
			name:          "all custom values",
			args:          []string{"cmd", "--port", "9090", "--db-file", "/tmp/test.db", "--web-dir", "/tmp/web", "--verbose", "--log-format", "json"},
			wantPort:      9090,
			wantDB:        "/tmp/test.db",
			wantWeb:       "/tmp/web",
			wantVerbose:   true,
			wantLogFormat: "json",
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
			verbose := flag.Bool("verbose", false, "Enable verbose logging (Debug level)")
			logFormat := flag.String("log-format", "text", "Log format: 'json' or 'text' (default: text)")
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
			if *verbose != tc.wantVerbose {
				t.Errorf("verbose = %v; want %v", *verbose, tc.wantVerbose)
			}
			if *logFormat != tc.wantLogFormat {
				t.Errorf("logFormat = %s; want %s", *logFormat, tc.wantLogFormat)
			}
		})
	}
}
