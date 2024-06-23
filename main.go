package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

var (
	backupStatus = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "backup_file_status",
		Help: "Indicates if the backup file was created (1) or not (0)",
	})
	lastBackupTimestamp = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "last_backup_file_timestamp",
		Help: "Timestamp of the last backup file",
	})
	lastBackupFileName = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "last_backup_file_name",
		Help: "Name of the last backup file",
	}, []string{"filename"})
	lastBackupSize = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "last_backup_size",
		Help: "Size of last backup file",
	}, []string{"size"})
	backupFolderSize = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "backup_folder_size",
		Help: "Size of backups folder",
	}, []string{"folder_size"})
)

func init() {
	prometheus.MustRegister(backupStatus)
	prometheus.MustRegister(lastBackupTimestamp)
	prometheus.MustRegister(lastBackupFileName)
	prometheus.MustRegister(lastBackupSize)
	prometheus.MustRegister(backupFolderSize)
}

func checkBackupFiles() {
	backupDirectory := "/home/danielnasc/c/bash/eficiencia-scripts/database-backups/backups"
	var latestFile string
	var latestModTime time.Time
	var totalFolderSize int64

	err := filepath.Walk(backupDirectory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			totalFolderSize += info.Size()
			if info.ModTime().After(latestModTime) {
				latestModTime = info.ModTime()
				latestFile = path
			}
		}
		return nil
	})

	if err != nil {
		log.Error("Error walking the path: ", err)
		backupStatus.Set(0)
		return
	}

	if latestFile == "" {
		backupStatus.Set(0)
	} else {
		backupStatus.Set(1)
		lastBackupTimestamp.Set(float64(latestModTime.Unix()))
		lastBackupFileName.With(prometheus.Labels{"filename": latestFile}).Set(1)

		info, err := os.Stat(latestFile)
		if err != nil {
			log.Error("Error getting file info: ", err)
		} else {
			lastBackupSize.With(prometheus.Labels{"size": formatBytes(info.Size())}).Set(float64(info.Size()))
		}
	}

	backupFolderSize.With(prometheus.Labels{"folder_size": formatBytes(totalFolderSize)}).Set(float64(totalFolderSize))
}

func formatBytes(bytes int64) string {
	const unit = 1000
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "kMGTPE"[exp])
}

func main() {
	go func() {
		for {
			checkBackupFiles()
			time.Sleep(2 * time.Second)
		}
	}()

	http.Handle("/metrics", promhttp.Handler())
	log.Info("Beginning to serve on port :8085")
	log.Fatal(http.ListenAndServe(":8085", nil))
}
