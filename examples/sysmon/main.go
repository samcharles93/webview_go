package main

import (
	"bufio"
	"log"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	webview "github.com/samcharles93/webview_go"
)

const html = `
<!DOCTYPE html>
<html>
<head>
	<meta charset="utf-8">
	<style>
		:root {
			font-family: 'Segoe UI', system-ui, -apple-system, sans-serif;
			color-scheme: light dark;
			--primary: #007aff;
			--bg: Canvas;
			--text: CanvasText;
			--card-bg: color-mix(in srgb, CanvasText, transparent 95%);
		}
		body {
			background-color: var(--bg);
			color: var(--text);
			margin: 0;
			padding: 20px;
			display: flex;
			flex-direction: column;
			gap: 20px;
		}
		.header {
			display: flex;
			justify-content: space-between;
			align-items: center;
		}
		.card {
			background: var(--card-bg);
			padding: 20px;
			border-radius: 12px;
			box-shadow: 0 4px 6px -1px rgb(0 0 0 / 0.1);
		}
		.stat-row {
			display: flex;
			justify-content: space-between;
			margin-bottom: 10px;
			font-weight: 500;
		}
		.progress-bar {
			height: 8px;
			background: color-mix(in srgb, var(--primary), transparent 80%);
			border-radius: 4px;
			overflow: hidden;
		}
		.progress-fill {
			height: 100%;
			background: var(--primary);
			width: 0%;
			transition: width 0.5s ease-out;
		}
		.footer {
			margin-top: auto;
			font-size: 0.8rem;
			opacity: 0.5;
			text-align: center;
		}
		h2 { margin: 0 0 15px 0; font-size: 1.1rem; opacity: 0.8; }
	</style>
</head>
<body>
	<div class="header">
		<h1>System Monitor</h1>
		<div id="os-tag" style="font-size: 0.8rem; padding: 4px 8px; background: var(--primary); color: white; border-radius: 4px;"></div>
	</div>

	<div class="card">
		<h2>CPU Usage</h2>
		<div class="stat-row">
			<span>Load</span>
			<span id="cpu-text">0%</span>
		</div>
		<div class="progress-bar">
			<div id="cpu-fill" class="progress-fill"></div>
		</div>
	</div>

	<div class="card">
		<h2>Memory Usage (Go Heap)</h2>
		<div class="stat-row">
			<span>Allocated</span>
			<span id="mem-text">0 MB</span>
		</div>
		<div class="progress-bar">
			<div id="mem-fill" class="progress-fill"></div>
		</div>
	</div>

	<div class="footer" id="version-info"></div>

	<script>
		const cpuText = document.getElementById('cpu-text');
		const cpuFill = document.getElementById('cpu-fill');
		const memText = document.getElementById('mem-text');
		const memFill = document.getElementById('mem-fill');
		const osTag = document.getElementById('os-tag');
		const versionInfo = document.getElementById('version-info');

		// Initialize OS and Version info
		window.getInitData().then(data => {
			osTag.textContent = data.os;
			versionInfo.textContent = 'webview_go v' + data.version.VersionNumber;
		});

		// Global function called from Go via Dispatch
		window.updateStats = (stats) => {
			cpuText.textContent = stats.cpu + '%';
			cpuFill.style.width = stats.cpu + '%';
			
			memText.textContent = stats.memAlloc + ' MB';
			// Scale memory bar to a reasonable 100MB for the example
			const memPercent = Math.min((stats.memAlloc / 100) * 100, 100);
			memFill.style.width = memPercent + '%';
		};
	</script>
</body>
</html>
`

type Stats struct {
	CPU      int     `json:"cpu"`
	MemAlloc float64 `json:"memAlloc"`
}

type InitData struct {
	OS      string              `json:"os"`
	Version webview.VersionInfo `json:"version"`
}

func main() {
	w, err := webview.NewWithOptions(webview.Options{
		Debug:  true,
		Title:  "System Monitor",
		Width:  400,
		Height: 500,
		Hint:   webview.HintMin,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer w.Destroy()

	if err := w.Bind("getInitData", func() InitData {
		return InitData{
			OS:      runtime.GOOS,
			Version: webview.Version(),
		}
	}); err != nil {
		log.Fatal(err)
	}

	// Background ticker to fetch and dispatch stats
	go func() {
		var lastIdle, lastTotal uint64
		for {
			time.Sleep(time.Second)

			// Get CPU usage (Linux specific, fallback for others)
			cpuUsage := 0
			if runtime.GOOS == "linux" {
				idle, total := getLinuxCPU()
				if lastTotal > 0 && total > lastTotal {
					diffIdle := idle - lastIdle
					diffTotal := total - lastTotal
					cpuUsage = int(100 * (diffTotal - diffIdle) / diffTotal)
				}
				lastIdle, lastTotal = idle, total
			} else {
				// Mock CPU for non-linux in this example
				cpuUsage = 15
			}

			// Get Memory usage (Cross-platform)
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			memMB := float64(m.Alloc) / 1024 / 1024

			stats := Stats{
				CPU:      cpuUsage,
				MemAlloc: float64(int(memMB*100)) / 100, // Round to 2 decimals
			}

			_ = w.DispatchCall("window.updateStats", stats)
		}
	}()

	if err := w.SetHtml(html); err != nil {
		log.Fatal(err)
	}
	if err := w.Run(); err != nil {
		log.Fatal(err)
	}
}

func getLinuxCPU() (idle, total uint64) {
	f, err := os.Open("/proc/stat")
	if err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	if scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 5 {
			return
		}
		// Sum all fields for total
		for i := 1; i < len(fields); i++ {
			val, _ := strconv.ParseUint(fields[i], 10, 64)
			total += val
			if i == 4 { // index 4 is idle
				idle = val
			}
		}
	}
	return
}
