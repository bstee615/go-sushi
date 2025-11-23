package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/sushi-go-game/backend/playtest/runner"
)

var verbose bool
var startServer bool

func main() {
	serverURL := flag.String("server", "ws://localhost:8080/ws", "WebSocket server URL")
	flag.BoolVar(&verbose, "verbose", false, "Print state snapshot after each turn")
	flag.BoolVar(&startServer, "start-server", false, "Start a test server on a random port")
	flag.Parse()

	if flag.NArg() == 0 {
		fmt.Println("Usage: playtest [--server URL] [--verbose] [--start-server] <test-file-or-directory>")
		os.Exit(1)
	}

	path := flag.Arg(0)
	
	// Start test server if requested
	if startServer {
		// Parse the playtest file to get deals if it's a single file
		var deals map[int]map[string][]string
		if !isDirectory(path) {
			playtest, err := runner.ParsePlaytest(path)
			if err == nil && len(playtest.Deals) > 0 {
				deals = playtest.Deals
			}
		}

		testServer, err := runner.StartTestServer(deals)
		if err != nil {
			fmt.Printf("Failed to start test server: %v\n", err)
			os.Exit(1)
		}
		defer testServer.Stop()
		
		*serverURL = testServer.URL
		fmt.Printf("Started test server on %s\n", testServer.URL)
	}
	
	// Check if path is a file or directory
	info, err := os.Stat(path)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	var exitCode int
	if info.IsDir() {
		exitCode = runDirectory(path, *serverURL)
	} else {
		exitCode = runFile(path, *serverURL)
	}

	os.Exit(exitCode)
}

func runFile(filepath string, serverURL string) int {
	testRunner := runner.NewTestRunner(serverURL)
	testRunner.SetVerbose(verbose)
	defer testRunner.Close()

	err := testRunner.RunPlaytest(filepath)
	success := err == nil

	testRunner.PrintResults(success, err)

	if success {
		return 0
	}
	return 1
}

func runDirectory(dirPath string, serverURL string) int {
	// Find all YAML files in the directory
	files, err := filepath.Glob(filepath.Join(dirPath, "*.yaml"))
	if err != nil {
		fmt.Printf("Error finding YAML files: %v\n", err)
		return 1
	}

	if len(files) == 0 {
		fmt.Printf("No YAML files found in %s\n", dirPath)
		return 1
	}

	fmt.Printf("Found %d test file(s)\n\n", len(files))

	passed := 0
	failed := 0

	for _, file := range files {
		fmt.Printf("Running %s...\n", filepath.Base(file))
		
		testRunner := runner.NewTestRunner(serverURL)
		testRunner.SetVerbose(verbose)
		err := testRunner.RunPlaytest(file)
		testRunner.Close()

		if err == nil {
			fmt.Println("✓ PASSED\n")
			passed++
		} else {
			fmt.Printf("✗ FAILED: %v\n\n", err)
			failed++
		}
	}

	fmt.Printf("\n=== Summary ===\n")
	fmt.Printf("Total: %d, Passed: %d, Failed: %d\n", len(files), passed, failed)

	if failed > 0 {
		return 1
	}
	return 0
}

func isDirectory(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}
