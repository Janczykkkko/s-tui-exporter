package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/crypto/ssh"
)

var (
	proxTemp = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "proxhost temp",
		Help: "xd",
	})
)

func getPackagePower() (float64, error) {
	// Retrieve environment variables
	hostname := os.Getenv("REMOTE_HOST")
	username := os.Getenv("REMOTE_USERNAME")
	password := os.Getenv("REMOTE_PASSWORD")

	// Check if variables are set
	if hostname == "" || username == "" || password == "" {
		log.Fatal("Required environment variables not set!")
	}

	// SSH configuration
	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	// SSH connection
	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:22", hostname), config)
	if err != nil {
		return 0, fmt.Errorf("failed to dial: %v", err)
	}
	defer client.Close()

	// Create a session
	session, err := client.NewSession()
	if err != nil {
		return 0, fmt.Errorf("failed to create session: %v", err)
	}
	defer session.Close()

	// Command to execute
	cmd := "s-tui -j"

	// Run the command
	output, err := session.CombinedOutput(cmd)
	if err != nil {
		return 0, fmt.Errorf("failed to run command: %v", err)
	}

	// Parsing JSON to extract Package Power
	var data map[string]interface{}
	err = json.Unmarshal(output, &data)
	if err != nil {
		return 0, fmt.Errorf("failed to unmarshal JSON: %v", err)
	}

	// Extracting Package Power metric
	powerVal, ok := data["Power"].(map[string]interface{})["package-0,0"]
	if !ok {
		return 0, fmt.Errorf("package Power not found in JSON")
	}

	power, err := strconv.ParseFloat(powerVal.(string), 64)
	if err != nil {
		return 0, fmt.Errorf("failed to convert power value: %v", err)
	}

	log.Printf("Request to %s completed succesfully", hostname)

	return power, nil
}

func main() {
	r := gin.Default()

	r.GET("/metrics", func(c *gin.Context) {
		power, err := getPackagePower()
		if err != nil {
			c.String(http.StatusInternalServerError, "Error getting power metric")
			return
		}
		proxTemp.Set(float64(power))
		c.String(http.StatusOK, "prox_power %f", power)
	})

	go func() {
		log.Fatal(r.Run(":8080"))
	}()

	select {}
}
