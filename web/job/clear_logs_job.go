package job

import (
	"os"
	"path/filepath"

	"x-ui/logger"
	"x-ui/xray"
)

type ClearLogsJob struct{}

func NewClearLogsJob() *ClearLogsJob {
	return new(ClearLogsJob)
}

// ensureFileExists creates the necessary directories and file if they don't exist
func ensureFileExists(path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	file.Close()
	return nil
}

func (j *ClearLogsJob) Run() {
	unified := xray.GetIPLimitLogPath()
	if err := ensureFileExists(unified); err != nil {
		logger.Warning("Failed to ensure 3xipl log exists:", unified, "-", err)
	}
}
