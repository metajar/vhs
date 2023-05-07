package main

import (
	"context"
	"fmt"
	expect "github.com/google/goexpect"
	"golang.org/x/crypto/ssh"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"
	"vhs/pkg/vhs/server"
)

func main() {
	cl := server.NewVhsServiceProtobufClient("http://127.0.0.1:8080", &http.Client{})
	// Replace with your device's IP, username, and password
	ip := "192.168.88.3"
	username := "grpc"
	password := "53cret"

	// List of commands to execute
	commands := []string{
		"term len 0",
		"show interfaces",
		"show version",
		"show run",
	}

	// SSH client configuration
	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	// Connect to the network device
	client, err := ssh.Dial("tcp", ip+":22", config)
	if err != nil {
		log.Fatalf("Failed to connect: %s", err)
	}
	defer client.Close()

	// Create a new expect session
	session, _, err := expect.SpawnSSH(client, time.Second*10, expect.Verbose(false))
	if err != nil {
		log.Fatalf("Failed to create expect session: %s", err)
	}
	defer session.Close()

	// Run commands and print output
	var combinedOutput string
	for _, cmd := range commands {
		output, err := runCommand(session, cmd)
		if err != nil {
			log.Fatalf("Failed to run command: %s", err)
		}
		output = redactSensitiveData(output) // Redact sensitive data
		combinedOutput += output + "\n"
	}
	backup, err := cl.Backup(context.Background(), &server.BackupRequest{Device: &server.Device{
		Host:    "br01.jared01",
		Payload: []byte(combinedOutput),
	}})
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(backup.Success, "with", backup.Status)
}

func runCommand(session *expect.GExpect, cmd string) (string, error) {

	err := session.Send(cmd + "\n\n")
	if err != nil {
		return "", err
	}
	_, _, err = session.Expect(regexp.MustCompile(`#`), time.Second*5)
	if err != nil {
		return "", err
	}

	// Send the command
	err = session.Send(cmd + "\n\n")
	if err != nil {
		return "", err
	}

	// Wait for the output
	output, _, err := session.Expect(regexp.MustCompile(`#`), time.Second*30)
	if err != nil {
		return "", err
	}

	// Format the output
	output = strings.TrimPrefix(output, cmd+"\r\n")
	output = strings.TrimSpace(output)
	lines := strings.Split(output, "\n")
	var filteredLines []string
	for i, line := range lines {
		if i == 0 {
			continue
		}
		if !strings.Contains(line, "RP/0/RP0/CPU0:XR-732#") {
			filteredLines = append(filteredLines, line)
		}
	}
	output = strings.Join(filteredLines, "\n")

	return fmt.Sprintf("++++++++++++++++++++++++++++++++++++++++++++++\n%s\n++++++++++++++++++++++++++++++++++++++++++++++\n%s", cmd, output), nil
}

func redactSensitiveData(output string) string {
	redactionPatterns := []string{
		`(?i)(password|secret)(\s+\d+)?\s+\S+`, // matches "password <password>", "secret <secret>", and case-insensitive variations
		// Add more regular expressions to redact other sensitive data
	}

	for _, pattern := range redactionPatterns {
		re := regexp.MustCompile(pattern)
		output = re.ReplaceAllStringFunc(output, func(match string) string {
			parts := strings.SplitN(match, " ", 3)
			if len(parts) == 3 {
				return parts[0] + " " + parts[1] + " REDACTED"
			} else if len(parts) == 2 {
				return parts[0] + " REDACTED"
			}
			return "REDACTED"
		})
	}

	return output
}
