package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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
		fmt.Printf("failed to run command: %v", err)
	}

	var status SystemStatus
	if err := json.Unmarshal([]byte(output), &status); err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Convert your dynamic maps to Prometheus metrics
	frequencyMetric := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "s_tui_frequency_data",
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
		Name: "s_tui_temperature_data",
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
	powerMetric := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "s_tui_power_data",
		Help: "Power data from the system",
	}, []string{"sensor"})
	for key, value := range status.Power {
		val, err := strconv.ParseFloat(value, 64)
		if err != nil {
			fmt.Println("Error parsing power value:", err)
			continue
		}
		powerMetric.With(prometheus.Labels{"sensor": key}).Set(val)
	}

	utilMetric := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "s_tui_utilisation_data",
		Help: "Utilisation data from the system",
	}, []string{"core"})
	for key, value := range status.Util {
		val, err := strconv.ParseFloat(value, 64)
		if err != nil {
			fmt.Println("Error parsing power value:", err)
			continue
		}
		utilMetric.With(prometheus.Labels{"core": key}).Set(val)
	}

	// Register the metrics with the Prometheus collector
	prometheus.MustRegister(frequencyMetric)
	prometheus.MustRegister(tempMetric)
	prometheus.MustRegister(powerMetric)
	prometheus.MustRegister(utilMetric)
	// Serve the metrics
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":8081", nil)
}
