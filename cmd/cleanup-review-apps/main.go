package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/superfly/fly-go"
	"github.com/superfly/fly-go/flaps"
)

const (
	prAppPrefix   = "go-sushi-pr-"
	productionApp = "go-sushi"
	maxAge        = 7 * 24 * time.Hour // 1 week
)

type appStatus struct {
	name    string
	status  string // "live", "scaled-down", "error"
	message string
	age     time.Duration
}

func main() {
	ctx := context.Background()

	// Get API token from environment
	token := os.Getenv("FLY_API_TOKEN")
	if token == "" {
		log.Fatal("FLY_API_TOKEN environment variable is required")
	}

	// Create Fly client
	client := fly.NewClientFromOptions(fly.ClientOptions{
		AccessToken: token,
	})

	fmt.Println("Checking for review apps older than 1 week...")

	// List all apps
	apps, err := client.GetApps(ctx, nil)
	if err != nil {
		log.Fatalf("Failed to list apps: %v", err)
	}

	// Filter review apps (PR-based only)
	var reviewApps []fly.App
	for _, app := range apps {
		if strings.HasPrefix(app.Name, prAppPrefix) && app.Name != productionApp {
			reviewApps = append(reviewApps, app)
		}
	}

	if len(reviewApps) == 0 {
		fmt.Println("No review apps found.")
		return
	}

	fmt.Printf("Found %d review apps to check:\n", len(reviewApps))
	for _, app := range reviewApps {
		fmt.Printf("  - %s\n", app.Name)
	}
	fmt.Println()

	var results []appStatus
	var errors []string

	// Process each review app
	for _, app := range reviewApps {
		fmt.Printf("Checking app: %s\n", app.Name)

		result := processApp(ctx, client, app.Name)
		results = append(results, result)

		if result.status == "error" {
			errors = append(errors, fmt.Sprintf("%s: %s", app.Name, result.message))
		}

		fmt.Println()
	}

	// Print summary
	printSummary(results)

	// Fail if there were errors
	if len(errors) > 0 {
		fmt.Println("\n❌ Workflow failed with errors:")
		for _, err := range errors {
			fmt.Printf("  - %s\n", err)
		}
		os.Exit(1)
	}

	fmt.Println("\n✅ Cleanup complete!")
}

func processApp(ctx context.Context, client *fly.Client, appName string) appStatus {
	result := appStatus{name: appName}

	// Get releases to find the most recent deployment
	// We check the most recent release (not the first) so that redeploying
	// to an existing review app resets the 7-day timer
	releases, err := client.GetAppReleasesMachines(ctx, appName, "", 1) // Get only the most recent release
	if err != nil {
		result.status = "error"
		result.message = fmt.Sprintf("Failed to fetch releases: %v", err)
		return result
	}

	if len(releases) == 0 {
		// No releases means app was never deployed - this is ok, just skip
		result.status = "skipped"
		result.message = "No releases found (app never deployed)"
		fmt.Printf("⚠️  No releases found for %s (app never deployed), skipping...\n", appName)
		return result
	}

	// Get the most recent release (first in the list)
	mostRecentRelease := releases[0]
	releaseTime := mostRecentRelease.CreatedAt

	if releaseTime.IsZero() {
		result.status = "error"
		result.message = "Could not determine creation date"
		fmt.Printf("❌ Could not determine creation date for %s\n", appName)
		return result
	}

	age := time.Since(releaseTime)
	result.age = age
	ageDays := int(age.Hours() / 24)

	if age < maxAge {
		result.status = "live"
		result.message = fmt.Sprintf("%d days old (less than 7 days)", ageDays)
		fmt.Printf("✓ App %s is %d days old, skipping (less than 7 days)\n", appName, ageDays)
		return result
	}

	// App is old enough to scale down
	fmt.Printf("📅 App %s is %d days old (last deployed: %s)\n", appName, ageDays, releaseTime.Format(time.RFC3339))
	fmt.Printf("⚙️  Scaling app to 0 machines...\n")

	// Get flaps client to scale machines
	flapsClient, err := flaps.NewWithOptions(ctx, flaps.NewClientOpts{
		AppName: appName,
	})
	if err != nil {
		result.status = "error"
		result.message = fmt.Sprintf("Failed to create flaps client: %v", err)
		fmt.Printf("❌ Failed to create flaps client for %s: %v\n", appName, err)
		return result
	}

	// List machines
	machines, err := flapsClient.List(ctx, "")
	if err != nil {
		result.status = "error"
		result.message = fmt.Sprintf("Failed to list machines: %v", err)
		fmt.Printf("❌ Failed to list machines for %s: %v\n", appName, err)
		return result
	}

	if len(machines) == 0 {
		result.status = "scaled-down"
		result.message = fmt.Sprintf("%d days old (already scaled to 0)", ageDays)
		fmt.Printf("✓ App %s already has 0 machines\n", appName)
		return result
	}

	// Stop all machines
	scaledCount := 0
	for _, machine := range machines {
		if machine.State == "stopped" {
			continue
		}

		input := fly.StopMachineInput{
			ID: machine.ID,
		}
		err := flapsClient.Stop(ctx, input, machine.LeaseNonce)
		if err != nil {
			result.status = "error"
			result.message = fmt.Sprintf("Failed to stop machine %s: %v", machine.ID, err)
			fmt.Printf("❌ Failed to stop machine %s: %v\n", machine.ID, err)
			return result
		}
		scaledCount++
	}

	result.status = "scaled-down"
	result.message = fmt.Sprintf("%d days old (scaled %d machines to 0)", ageDays, scaledCount)
	fmt.Printf("✅ Successfully scaled %s to 0 (%d machines stopped)\n", appName, scaledCount)
	return result
}

func printSummary(results []appStatus) {
	// Print to console
	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("CLEANUP SUMMARY")
	fmt.Println(strings.Repeat("=", 70))

	live := 0
	scaledDown := 0
	skipped := 0
	errorCount := 0

	fmt.Println("\nReview Apps Status:")
	for _, r := range results {
		statusEmoji := ""
		switch r.status {
		case "live":
			statusEmoji = "🟢"
			live++
		case "scaled-down":
			statusEmoji = "🔴"
			scaledDown++
		case "skipped":
			statusEmoji = "⚪"
			skipped++
		case "error":
			statusEmoji = "❌"
			errorCount++
		}
		fmt.Printf("  %s %s: %s - %s\n", statusEmoji, r.name, r.status, r.message)
	}

	fmt.Println("\nSummary:")
	fmt.Printf("  Total review apps: %d\n", len(results))
	fmt.Printf("  Live (< 7 days old): %d\n", live)
	fmt.Printf("  Scaled down (≥ 7 days old): %d\n", scaledDown)
	fmt.Printf("  Skipped (no releases): %d\n", skipped)
	fmt.Printf("  Errors encountered: %d\n", errorCount)
	fmt.Println(strings.Repeat("=", 70))

	// Write to GitHub Actions summary if available
	summaryFile := os.Getenv("GITHUB_STEP_SUMMARY")
	if summaryFile != "" {
		writeGitHubSummary(summaryFile, results, live, scaledDown, skipped, errorCount)
	}
}

func writeGitHubSummary(summaryFile string, results []appStatus, live, scaledDown, skipped, errorCount int) {
	f, err := os.OpenFile(summaryFile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		fmt.Printf("Warning: Could not write to GitHub summary: %v\n", err)
		return
	}
	defer f.Close()

	// Write markdown summary
	f.WriteString("## 🧹 Review App Cleanup Summary\n\n")

	// Stats table
	f.WriteString("### Statistics\n\n")
	f.WriteString("| Metric | Count |\n")
	f.WriteString("|--------|-------|\n")
	f.WriteString(fmt.Sprintf("| **Total Review Apps** | %d |\n", len(results)))
	f.WriteString(fmt.Sprintf("| 🟢 Live (< 7 days) | %d |\n", live))
	f.WriteString(fmt.Sprintf("| 🔴 Scaled Down (≥ 7 days) | %d |\n", scaledDown))
	f.WriteString(fmt.Sprintf("| ⚪ Skipped (no releases) | %d |\n", skipped))
	f.WriteString(fmt.Sprintf("| ❌ Errors | %d |\n", errorCount))
	f.WriteString("\n")

	// Detailed app status
	if len(results) > 0 {
		f.WriteString("### Review Apps Status\n\n")
		f.WriteString("| Status | App Name | Details |\n")
		f.WriteString("|--------|----------|----------|\n")
		for _, r := range results {
			statusEmoji := ""
			switch r.status {
			case "live":
				statusEmoji = "🟢"
			case "scaled-down":
				statusEmoji = "🔴"
			case "skipped":
				statusEmoji = "⚪"
			case "error":
				statusEmoji = "❌"
			}
			f.WriteString(fmt.Sprintf("| %s %s | `%s` | %s |\n", statusEmoji, r.status, r.name, r.message))
		}
		f.WriteString("\n")
	}

	// Add footer
	if errorCount > 0 {
		f.WriteString("⚠️ **Warning:** Some errors were encountered during cleanup. See details above.\n")
	} else {
		f.WriteString("✅ **Success:** All review apps processed successfully.\n")
	}
}
