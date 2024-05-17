package main

import (
	"fmt"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Definir métricas
var (
	backupSize = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "backup_size_bytes",
		Help: "Size of the backup in bytes.",
	})
	diskUsage = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "disk_usage_bytes",
		Help: "Disk usage in bytes.",
	})
	percentageUse = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "disk_usage_percentage",
		Help: "Percentage of disk used.",
	})
	currentDate = prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "current_date_timestamp",
		Help: "Current date as a Unix timestamp.",
	}, func() float64 {
		return float64(time.Now().Unix())
	})
)

func init() {
	// Registrar as métricas no registrador padrão
	prometheus.MustRegister(backupSize)
	prometheus.MustRegister(diskUsage)
	prometheus.MustRegister(percentageUse)
	prometheus.MustRegister(currentDate)
}

func parseSize(sizeStr string) (float64, error) {
	var multiplier float64
	sizeStr = strings.TrimSpace(sizeStr)
	switch {
	case strings.HasSuffix(sizeStr, "G"):
		multiplier = 1024 * 1024 * 1024
		sizeStr = strings.TrimSuffix(sizeStr, "G")
	case strings.HasSuffix(sizeStr, "M"):
		multiplier = 1024 * 1024
		sizeStr = strings.TrimSuffix(sizeStr, "M")
	case strings.HasSuffix(sizeStr, "K"):
		multiplier = 1024
		sizeStr = strings.TrimSuffix(sizeStr, "K")
	default:
		multiplier = 1
	}

	sizeValue, err := strconv.ParseFloat(sizeStr, 64)
	if err != nil {
		return 0, err
	}

	return sizeValue * multiplier, nil
}

func getBackupSize() float64 {
	cmd := exec.Command("du", "-s", "/home/a/Área de Trabalho/backup")
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("Error getting backup size:", err)
		return 0
	}
	backupSizeStr := strings.Split(string(output), "\t")[0]
	backupSize, err := strconv.Atoi(backupSizeStr)
	if err != nil {
		fmt.Println("Error converting backup size to integer:", err)
		return 0
	}
	backupSizeGb := float64(backupSize)
	return backupSizeGb / (1024.0 * 1024.0 * 1024.0)
}

func getDiskUsage() (float64, error) {
	cmd := exec.Command("df", "-h", "/")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return 0, fmt.Errorf("error getting disk usage: %v", err)
	}
	lines := strings.Split(string(output), "\n")
	if len(lines) < 2 {
		return 0, fmt.Errorf("unexpected output format")
	}
	usageLine := strings.Fields(lines[1])[2]
	usageValue, err := parseSize(usageLine)
	if err != nil {
		return 0, fmt.Errorf("error parsing disk usage: %v", err)
	}
	return usageValue, nil
}

func getDiskUsagePercentage() float64 {
	return 50.0 // Exemplo: 50%
}

func updateMetrics() {
	backupSize.Set(getBackupSize())
	percentageUse.Set(getDiskUsagePercentage())
	usage, err := getDiskUsage()
	if err != nil {
		fmt.Println(err)
	} else {
		diskUsage.Set(usage / (1024.0 * 1024.0 * 1024.0))
	}

}

func main() {
	// Atualiza as métricas periodicamente (por exemplo, a cada 10 segundos)
	go func() {
		for {
			updateMetrics()
			time.Sleep(10 * time.Second)
		}
	}()

	// Endpoint para expor as métricas
	http.Handle("/metrics", promhttp.Handler())

	fmt.Println("Listening on :8081")
	http.ListenAndServe(":8081", nil)
}
