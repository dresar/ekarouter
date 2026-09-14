package cli

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type ServerStatus struct {
	Running bool
	Port    int
	Host    string
	PID     int
	Uptime  string
	Latency time.Duration
}

func GetPIDFilePath() string {
	return filepath.Join("data", "ekarouter.pid")
}

func ReadPID() int {
	data, err := os.ReadFile(GetPIDFilePath())
	if err != nil {
		return 0
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return 0
	}
	return pid
}

func WritePID(pid int) error {
	_ = os.MkdirAll("data", 0755)
	return os.WriteFile(GetPIDFilePath(), []byte(strconv.Itoa(pid)), 0644)
}

func RemovePID() {
	_ = os.Remove(GetPIDFilePath())
}

func ProbeGateway(host string, port int) *ServerStatus {
	if host == "" || host == "0.0.0.0" {
		host = "127.0.0.1"
	}
	if port <= 0 {
		port = 8999
	}

	url := fmt.Sprintf("http://%s:%d/health", host, port)
	client := http.Client{Timeout: 1500 * time.Millisecond}

	start := time.Now()
	resp, err := client.Get(url)
	latency := time.Since(start)

	if err != nil || resp.StatusCode != http.StatusOK {
		return &ServerStatus{
			Running: false,
			Host:    host,
			Port:    port,
			PID:     ReadPID(),
		}
	}
	defer resp.Body.Close()

	var payload struct {
		Status string `json:"status"`
		Uptime string `json:"uptime"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&payload)

	pid := ReadPID()

	return &ServerStatus{
		Running: true,
		Host:    host,
		Port:    port,
		PID:     pid,
		Uptime:  payload.Uptime,
		Latency: latency,
	}
}

func StartGatewayDaemon(dbPath, migDir string, port int) (*ServerStatus, error) {
	status := ProbeGateway("127.0.0.1", port)
	if status.Running {
		return status, fmt.Errorf("gateway is already running on port %d (PID: %d)", port, status.PID)
	}

	exe, err := os.Executable()
	if err != nil {
		exe = "./ekarouter.exe"
	}

	args := []string{
		"--db", dbPath,
		"--migrations", migDir,
		"--port", strconv.Itoa(port),
	}

	cmd := exec.Command(exe, args...)
	cmd.Dir, _ = os.Getwd()

	logFile, err := os.OpenFile(filepath.Join("data", "gateway.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err == nil {
		cmd.Stdout = logFile
		cmd.Stderr = logFile
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("spawn gateway process: %w", err)
	}

	pid := cmd.Process.Pid
	_ = WritePID(pid)

	for i := 0; i < 15; i++ {
		time.Sleep(200 * time.Millisecond)
		s := ProbeGateway("127.0.0.1", port)
		if s.Running {
			s.PID = pid
			return s, nil
		}
	}

	return nil, fmt.Errorf("gateway process started (PID: %d) but health check failed on port %d", pid, port)
}

func StopGatewayDaemon(port int) error {
	pid := ReadPID()
	if pid > 0 {
		cmd := exec.Command("taskkill", "/PID", strconv.Itoa(pid), "/F")
		_ = cmd.Run()
		RemovePID()
	}

	time.Sleep(500 * time.Millisecond)
	status := ProbeGateway("127.0.0.1", port)
	if status.Running {
		return fmt.Errorf("gateway is still responding on port %d after stop signal", port)
	}

	return nil
}
