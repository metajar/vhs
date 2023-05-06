package main

import (
	"context"
	"log"
	"time"
	"vhs/devices"
	"vhs/git"
)

const repoURL = "git@github.com:metajar/testbackup.git"

func main() {
	// Create a channel for devices.
	deviceChan := make(chan devices.Device, 100) // Buffer size of 100, adjust as needed.

	// Initialize the Git object.
	g := git.NewGit("/tmp", "main") // Set the path to your repo and the branch name.

	// Clone the Git repository.
	if err := g.Clone(repoURL); err != nil {
		log.Fatalf("Failed to clone repository: %v\n", err)
	}

	// Set up the upstream branch for the local branch.
	if err := g.SetUpstreamBranch(); err != nil {
		log.Fatalf("Failed to set upstream branch: %v\n", err)
	}

	// Pull the latest changes before starting the application.
	if err := g.Pull(); err != nil {
		log.Fatalf("Failed to pull changes: %v\n", err)
	}

	// Launch a goroutine to process devices.
	go func() {
		for device := range deviceChan {
			if err := g.SaveDeviceConfiguration(device); err != nil {
				log.Printf("Failed to save configuration for device %s: %v\n", device.Name, err)
				continue
			}
			log.Printf("Saved configuration for device %s\n", device.Name)
		}
	}()

	// Here you can send devices to the channel. This is just an example.
	deviceChan <- devices.NewDevice("co01.test01", []byte("device configuration\nWow\nNice\nSeriously!"))
	//deviceChan <- devices.NewDevice("la01.test01", []byte("another device configuration\nWow"))

	// Close the channel when you're done sending devices.
	close(deviceChan)

	go g.StartPeriodicPush(context.Background(), time.Second*10)

	// Keep the main function from returning, since our goroutines are running in the background.
	select {}
}
