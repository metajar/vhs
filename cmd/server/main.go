package main

import (
	"context"
	"log"
	"net/http"
	"time"
	"vhs/devices"
	"vhs/git"
	"vhs/pkg/vhs/server"
)

var deviceChan = make(chan devices.Device, 100) // Buffer size of 100, adjust as needed.

type VhsServer struct {
	VHS git.Git
}

func (v *VhsServer) Backup(ctx context.Context, request *server.BackupRequest) (*server.BackupResponse, error) {
	dev := request.GetDevice()
	deviceChan <- devices.NewDevice(dev.GetHost(), dev.GetPayload())
	return &server.BackupResponse{
		Success: true,
		Status:  200,
	}, nil
}

const repoURL = "git@github.com:metajar/testbackup.git"

func main() {
	g := git.NewGit("/tmp/vhs", "main") // Set the path to your repo and the branch name.
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
	v := VhsServer{VHS: g}
	go func() {
		for device := range deviceChan {
			if err := g.SaveDeviceConfiguration(device); err != nil {
				log.Printf("Failed to save configuration for device %s: %v\n", device.Name, err)
				continue
			}
			log.Printf("Saved configuration for device %s\n", device.Name)
		}
	}()
	go g.StartPeriodicPush(context.Background(), time.Second*10, time.Second*60)
	twirpHandler := server.NewVhsServiceServer(&v)
	mux := http.NewServeMux()
	mux.Handle(twirpHandler.PathPrefix(), twirpHandler)
	http.ListenAndServe(":8080", mux)
}
