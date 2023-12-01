package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type SystemStatus struct {
	Frequency map[string]string `json:"Frequency"`
	Temp      map[string]string `json:"Temp"`
	Util      map[string]string `json:"Util"`
	Power     map[string]string `json:"Power"`
}

func updateMetrics() {
	for {
		cmd := exec.Command("s-tui", "-j")
		output, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Printf("failed to run command: %v", err)
			continue
		}

		var status SystemStatus
		if err := json.Unmarshal([]byte(output), &status); err != nil {
			fmt.Println("Error:", err)
			continue
		}

		updateMetric(frequencyMetric, status.Frequency, "core")
		updateMetric(tempMetric, status.Temp, "sensor")
		updateMetric(powerMetric, status.Power, "sensor")
		updateMetric(utilMetric, status.Util, "core")

		time.Sleep(30 * time.Second)
	}
}

func updateMetric(metric *prometheus.GaugeVec, data map[string]string, labelName string) {
	for key, value := range data {
		val, err := strconv.ParseFloat(value, 64)
		if err != nil {
			fmt.Printf("Error parsing %s value: %v\n", labelName, err)
			continue
		}
		metric.With(prometheus.Labels{labelName: key}).Set(val)
	}
}

var (
	frequencyMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "s_tui_frequency_data",
		Help: "Frequency data from the system",
	}, []string{"core"})

	tempMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "s_tui_temperature_data",
		Help: "Temperature data from the system",
	}, []string{"sensor"})

	powerMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "s_tui_power_data",
		Help: "Power data from the system",
	}, []string{"sensor"})

	utilMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "s_tui_utilisation_data",
		Help: "Utilisation data from the system",
	}, []string{"core"})
)

func main() {
	prometheus.MustRegister(frequencyMetric)
	prometheus.MustRegister(tempMetric)
	prometheus.MustRegister(powerMetric)
	prometheus.MustRegister(utilMetric)

	go updateMetrics()

	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":8081", nil)
}
