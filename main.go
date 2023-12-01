package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/crypto/ssh"
)

type SystemStatus struct {
	Frequency map[string]string `json:"Frequency"`
	Temp      map[string]string `json:"Temp"`
	Util      map[string]string `json:"Util"`
	Power     map[string]string `json:"Power"`
}

func main() {
	cmd := exec.Command("s-tui", "-j")

	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Errorf("failed to run command: %v", err)
	}

	var status SystemStatus
	if err := json.Unmarshal([]byte(output), &status); err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Convert your dynamic maps to Prometheus metrics
	frequencyMetric := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "frequency_data",
		Help: "Frequency data from the system",
	}, []string{"core"})
	for key, value := range status.Frequency {
		val, err := strconv.ParseFloat(value, 64)
		if err != nil {
			fmt.Println("Error parsing frequency value:", err)
			continue
		}
		frequencyMetric.With(prometheus.Labels{"core": key}).Set(val)
	}

	tempMetric := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "temperature_data",
		Help: "Temperature data from the system",
	}, []string{"sensor"})
	for key, value := range status.Temp {
		val, err := strconv.ParseFloat(value, 64)
		if err != nil {
			fmt.Println("Error parsing temperature value:", err)
			continue
		}
		tempMetric.With(prometheus.Labels{"sensor": key}).Set(val)
	}

	// Register the metrics with the Prometheus collector
	prometheus.MustRegister(frequencyMetric)
	prometheus.MustRegister(tempMetric)

	// Serve the metrics
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":8080", nil)
}

func readPrivateKey(keyPath string) ssh.Signer {
	key, err := ioutil.ReadFile(keyPath)
	if err != nil {
		log.Fatalf("Unable to read private key: %v", err)
	}
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		log.Fatalf("Unable to parse private key: %v", err)
	}
	return signer
}
