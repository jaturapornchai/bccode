package handlers

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"smlcloudplatform/internal/goapi/logger"

	"github.com/labstack/echo/v4"
)

// Deploy state
var (
	deployMu      sync.Mutex
	deployRunning bool
	lastDeployLog []string
	lastDeployAt  time.Time
	lastDeployOK  bool
)

// getDeployToken คืนค่า deploy token จาก env
func getDeployToken() string {
	token := os.Getenv("DEPLOY_TOKEN")
	if token == "" {
		return "" // ไม่มี token = ปิดฟีเจอร์ deploy
	}
	return token
}

// validateDeployToken ตรวจสอบ token
func validateDeployToken(c echo.Context) error {
	expectedToken := getDeployToken()
	if expectedToken == "" {
		return fmt.Errorf("ฟีเจอร์ deploy ถูกปิด — ตั้ง DEPLOY_TOKEN ใน environment")
	}

	token := c.QueryParam("token")
	if token == "" {
		token = c.Request().Header.Get("X-Deploy-Token")
	}
	if token != expectedToken {
		return fmt.Errorf("token ไม่ถูกต้อง")
	}
	return nil
}

// DeployBackendHandler — POST /api/deploy/backend?token=xxx
// รัน: git pull → docker compose build goapi mainapi → docker compose up -d goapi mainapi
func DeployBackendHandler(c echo.Context) error {
	if err := validateDeployToken(c); err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
	}

	deployMu.Lock()
	if deployRunning {
		deployMu.Unlock()
		return c.JSON(http.StatusConflict, map[string]interface{}{
			"success": false,
			"error":   "กำลัง deploy อยู่ กรุณารอสักครู่",
		})
	}
	deployRunning = true
	deployMu.Unlock()

	// รัน async — ตอบกลับทันที
	go func() {
		defer func() {
			deployMu.Lock()
			deployRunning = false
			deployMu.Unlock()
		}()

		repoDir := os.Getenv("DEPLOY_REPO_DIR")
		if repoDir == "" {
			repoDir = "/opt/bcaiclouderp"
		}

		composeFile := os.Getenv("DEPLOY_COMPOSE_FILE")
		if composeFile == "" {
			composeFile = "deploy/docker-compose.dev-vps.yml"
		}

		commands := []struct {
			name string
			cmd  string
			args []string
			dir  string
		}{
			{
				name: "git pull",
				cmd:  "git",
				args: []string{"pull", "origin", "develop"},
				dir:  repoDir,
			},
			{
				name: "docker compose build",
				cmd:  "docker",
				args: []string{"compose", "-f", composeFile, "build", "goapi", "mainapi"},
				dir:  repoDir,
			},
			{
				name: "docker compose up -d",
				cmd:  "docker",
				args: []string{"compose", "-f", composeFile, "up", "-d", "goapi", "mainapi"},
				dir:  repoDir,
			},
		}

		var logs []string
		startTime := time.Now()
		logs = append(logs, fmt.Sprintf("[%s] เริ่ม deploy backend...", startTime.Format("15:04:05")))
		logger.Info("🚀 เริ่ม deploy backend...")

		success := true
		for _, c := range commands {
			stepStart := time.Now()
			logs = append(logs, fmt.Sprintf("[%s] รัน: %s", stepStart.Format("15:04:05"), c.name))
			logger.Info("📦 Deploy: %s", c.name)

			output, err := runCommand(c.cmd, c.args, c.dir)
			elapsed := time.Since(stepStart).Seconds()

			if err != nil {
				errMsg := fmt.Sprintf("[%s] ❌ %s ล้มเหลว (%.1fs): %v\n%s",
					time.Now().Format("15:04:05"), c.name, elapsed, err, output)
				logs = append(logs, errMsg)
				logger.Error("❌ Deploy %s ล้มเหลว: %v", c.name, err)
				success = false
				break
			}

			okMsg := fmt.Sprintf("[%s] ✅ %s สำเร็จ (%.1fs)", time.Now().Format("15:04:05"), c.name, elapsed)
			logs = append(logs, okMsg)
			if output != "" {
				// เก็บแค่ 10 บรรทัดสุดท้าย
				lines := strings.Split(strings.TrimSpace(output), "\n")
				if len(lines) > 10 {
					lines = lines[len(lines)-10:]
				}
				logs = append(logs, strings.Join(lines, "\n"))
			}
		}

		totalElapsed := time.Since(startTime).Seconds()
		if success {
			logs = append(logs, fmt.Sprintf("[%s] 🎉 Deploy backend สำเร็จ (%.1fs)", time.Now().Format("15:04:05"), totalElapsed))
			logger.Success("🎉 Deploy backend สำเร็จ (%.1fs)", totalElapsed)
		} else {
			logs = append(logs, fmt.Sprintf("[%s] 💥 Deploy backend ล้มเหลว (%.1fs)", time.Now().Format("15:04:05"), totalElapsed))
			logger.Error("💥 Deploy backend ล้มเหลว (%.1fs)", totalElapsed)
		}

		deployMu.Lock()
		lastDeployLog = logs
		lastDeployAt = time.Now()
		lastDeployOK = success
		deployMu.Unlock()
	}()

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "เริ่ม deploy backend แล้ว — ใช้ /api/deploy/status เพื่อดูสถานะ",
	})
}

// DeployFrontendHandler — POST /api/deploy/frontend?token=xxx
// รัน: git pull → copy web files
func DeployFrontendHandler(c echo.Context) error {
	if err := validateDeployToken(c); err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
	}

	deployMu.Lock()
	if deployRunning {
		deployMu.Unlock()
		return c.JSON(http.StatusConflict, map[string]interface{}{
			"success": false,
			"error":   "กำลัง deploy อยู่ กรุณารอสักครู่",
		})
	}
	deployRunning = true
	deployMu.Unlock()

	go func() {
		defer func() {
			deployMu.Lock()
			deployRunning = false
			deployMu.Unlock()
		}()

		repoDir := os.Getenv("DEPLOY_REPO_DIR")
		if repoDir == "" {
			repoDir = "/opt/bcaiclouderp"
		}

		webSrcDir := repoDir + "/bcmerchant/build/web"
		webDestDir := os.Getenv("DEPLOY_WEB_DIR")
		if webDestDir == "" {
			webDestDir = repoDir + "/deploy/web/bcmerchant"
		}

		commands := []struct {
			name string
			cmd  string
			args []string
			dir  string
		}{
			{
				name: "git pull",
				cmd:  "git",
				args: []string{"pull", "origin", "develop"},
				dir:  repoDir,
			},
			{
				name: "copy web files",
				cmd:  "sh",
				args: []string{"-c", fmt.Sprintf("cp -r %s/* %s/", webSrcDir, webDestDir)},
				dir:  repoDir,
			},
		}

		var logs []string
		startTime := time.Now()
		logs = append(logs, fmt.Sprintf("[%s] เริ่ม deploy frontend...", startTime.Format("15:04:05")))
		logger.Info("🚀 เริ่ม deploy frontend...")

		success := true
		for _, c := range commands {
			stepStart := time.Now()
			logs = append(logs, fmt.Sprintf("[%s] รัน: %s", stepStart.Format("15:04:05"), c.name))

			output, err := runCommand(c.cmd, c.args, c.dir)
			elapsed := time.Since(stepStart).Seconds()

			if err != nil {
				logs = append(logs, fmt.Sprintf("[%s] ❌ %s ล้มเหลว (%.1fs): %v\n%s",
					time.Now().Format("15:04:05"), c.name, elapsed, err, output))
				logger.Error("❌ Deploy frontend %s ล้มเหลว: %v", c.name, err)
				success = false
				break
			}

			logs = append(logs, fmt.Sprintf("[%s] ✅ %s สำเร็จ (%.1fs)", time.Now().Format("15:04:05"), c.name, elapsed))
		}

		totalElapsed := time.Since(startTime).Seconds()
		if success {
			logs = append(logs, fmt.Sprintf("[%s] 🎉 Deploy frontend สำเร็จ (%.1fs)", time.Now().Format("15:04:05"), totalElapsed))
			logger.Success("🎉 Deploy frontend สำเร็จ (%.1fs)", totalElapsed)
		} else {
			logs = append(logs, fmt.Sprintf("[%s] 💥 Deploy frontend ล้มเหลว (%.1fs)", time.Now().Format("15:04:05"), totalElapsed))
		}

		deployMu.Lock()
		lastDeployLog = logs
		lastDeployAt = time.Now()
		lastDeployOK = success
		deployMu.Unlock()
	}()

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "เริ่ม deploy frontend แล้ว — ใช้ /api/deploy/status เพื่อดูสถานะ",
	})
}

// DeployStatusHandler — GET /api/deploy/status?token=xxx
func DeployStatusHandler(c echo.Context) error {
	if err := validateDeployToken(c); err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
	}

	deployMu.Lock()
	running := deployRunning
	logs := make([]string, len(lastDeployLog))
	copy(logs, lastDeployLog)
	at := lastDeployAt
	ok := lastDeployOK
	deployMu.Unlock()

	status := "idle"
	if running {
		status = "running"
	} else if !at.IsZero() {
		if ok {
			status = "success"
		} else {
			status = "failed"
		}
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success":        true,
		"deploy_status":  status,
		"deploy_running": running,
		"last_deploy_at": at,
		"last_deploy_ok": ok,
		"logs":           logs,
	})
}

// runCommand รันคำสั่งและคืน output
func runCommand(command string, args []string, dir string) (string, error) {
	cmd := exec.Command(command, args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")

	// รวม stdout + stderr
	var output strings.Builder

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", fmt.Errorf("สร้าง stdout pipe ล้มเหลว: %w", err)
	}

	cmd.Stderr = cmd.Stdout // รวม stderr ไปที่ stdout

	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("เริ่มคำสั่งล้มเหลว: %w", err)
	}

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()
		output.WriteString(line + "\n")
		logger.Info("[deploy] %s", line)
	}

	if err := cmd.Wait(); err != nil {
		return output.String(), fmt.Errorf("คำสั่งล้มเหลว: %w", err)
	}

	return output.String(), nil
}
