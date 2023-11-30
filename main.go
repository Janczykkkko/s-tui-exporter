package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	proxTemp = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "proxhost temp",
		Help: "xd",
	})
)

func getPackagePower() (float64, error) {
	// Command to execute locally
	cmd := exec.Command("s-tui", "-j")

	// Run the command
	output, err := cmd.CombinedOutput()
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

	log.Printf("Command execution completed successfully")

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
		log.Fatal(r.Run(":8081"))
	}()

	select {}
}
