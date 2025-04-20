// Package main provides the entry point for the watchtower application.
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/httpoz/watchtower/pkg/notification"
	"github.com/httpoz/watchtower/pkg/packagemanager/apt"
	"github.com/httpoz/watchtower/pkg/riskassessor"
	"github.com/httpoz/watchtower/pkg/storage"
)

func main() {
	ctx := context.Background()

	packageManager := apt.NewManager()
	riskAssessor := riskassessor.NewOpenAIAssessor()
	storageManager, err := storage.DefaultStorage()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing storage: %v\n", err)
		os.Exit(1)
	}

	snapshotID := time.Now().Format("20060102") // snapshots are only created once a day

	snapshotFile, err := storageManager.CheckSnapshot(snapshotID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating snapshot: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Snapshot file path: %s\n", snapshotFile)

	fmt.Println("Collecting information about installed packages...")
	installedPackages, err := packageManager.GetInstalledPackages()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching installed packages: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Found %d installed packages\n", len(installedPackages))

	// Save snapshot of installed packages
	_, err = storageManager.WriteSnapshot(installedPackages, snapshotID, storage.SnapshotTypeInstalled)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing snapshot: %v\n", err)
		os.Exit(1)
	}

	// Get upgradable packages
	fmt.Println("Checking for available updates...")
	upgradableMap := packageManager.GetUpgradablePackages()
	fmt.Printf("Found %d packages with available updates\n", len(upgradableMap))

	upgradablePackages := packageManager.EnrichWithUpgradeInfo(installedPackages, upgradableMap)

	// Score the risk of upgradable packages
	fmt.Println("Assessing update risks...")
	scoredPackages, err := riskAssessor.ScorePackageRisks(ctx, upgradablePackages)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error assessing risks: %v\n", err)
		os.Exit(1)
	}

	// Save snapshot of upgradable packages with risk assessment
	_, err = storageManager.WriteSnapshot(scoredPackages, snapshotID, storage.SnapshotTypeUpdates)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing updates snapshot: %v\n", err)
		os.Exit(1)
	}

	// Generate and save summary
	fmt.Println("Generating update summary...")
	summaryPath, err := storageManager.WriteSummaryIfMissing(snapshotID, scoredPackages)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing summary: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Summary written to: %s\n", summaryPath)

	// Update low-risk packages
	fmt.Println("Updating safe packages...")
	err = packageManager.UpdatePackages(scoredPackages)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error updating packages: %v\n", err)
		os.Exit(1)
	}

	// Send notification if TELEGRAM_BOT_TOKEN is set
	if os.Getenv("WATCHTOWER_TELEGRAM_BOT_KEY") != "" {
		fmt.Println("Sending notification...")
		notifier, err := notification.DefaultTelegramNotifier()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error initializing notifier: %v\n", err)
		} else {
			err = notifier.SendSummary(summaryPath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error sending notification: %v\n", err)
			} else {
				fmt.Println("Notification sent successfully")
			}
		}
	}

	fmt.Println("All operations completed successfully.")
}
