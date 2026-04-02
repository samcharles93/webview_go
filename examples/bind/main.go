package main

import (
	"fmt"
	"log"
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
			font-family: system-ui, -apple-system, sans-serif;
			color-scheme: light dark;
			background-color: Canvas;
			color: CanvasText;
		}
		body {
			display: flex;
			flex-direction: column;
			align-items: center;
			justify-content: center;
			height: 100vh;
			margin: 0;
			text-align: center;
		}
		.container {
			padding: 2rem;
			border-radius: 12px;
			background: color-mix(in srgb, CanvasText, transparent 95%);
			box-shadow: 0 4px 6px -1px rgb(0 0 0 / 0.1);
		}
		button {
			appearance: none;
			background-color: #007aff;
			color: white;
			border: none;
			padding: 8px 16px;
			border-radius: 6px;
			font-size: 1rem;
			cursor: pointer;
			transition: opacity 0.2s;
		}
		button:hover { opacity: 0.8; }
		button:active { opacity: 0.6; }
		#count { font-weight: bold; font-size: 1.5rem; margin: 1rem 0; }
		.version { margin-top: 2rem; font-size: 0.8rem; opacity: 0.6; }
		#uptime { font-size: 0.9rem; margin-top: 0.5rem; color: #007aff; }
	</style>
</head>
<body>
	<div class="container">
		<h1>Webview Go</h1>
		<div id="count">0</div>
		<button id="increment">Increment Counter</button>
		<div id="uptime">Uptime: 0s</div>
		<div class="version" id="version"></div>
	</div>

	<script>
		const countEl = document.getElementById('count');
		const uptimeEl = document.getElementById('uptime');
		const versionEl = document.getElementById('version');
		const btn = document.getElementById('increment');

		// Call the Go 'getVersion' binding
		window.getVersion().then(v => {
			versionEl.textContent = 'v' + v.VersionNumber + ' (' + v.Major + '.' + v.Minor + '.' + v.Patch + ')';
		});

		btn.addEventListener('click', () => {
			// Call the Go 'increment' binding
			window.increment().then(newCount => {
				countEl.textContent = newCount;
			}).catch(err => {
				alert("Error: " + err);
			});
		});

		// This will be called from Go via Dispatch/Eval
		window.updateUptime = (seconds) => {
			uptimeEl.textContent = 'Uptime: ' + seconds + 's';
		};
	</script>
</body>
</html>
`

func main() {
	var count int
	start := time.Now()

	w, err := webview.NewWithOptions(webview.Options{
		Debug:  true,
		Title:  "Modern Bind Example",
		Width:  480,
		Height: 420,
		Hint:   webview.HintNone,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer w.Destroy()

	// Bind the version info
	if err := w.Bind("getVersion", func() webview.VersionInfo {
		return webview.Version()
	}); err != nil {
		log.Fatal(err)
	}

	// Bind an increment function
	if err := w.Bind("increment", func() (int, error) {
		count++
		if count > 10 {
			// Example of error handling
			return count, fmt.Errorf("count is too high!")
		}
		return count, nil
	}); err != nil {
		log.Fatal(err)
	}

	// Start a background goroutine to update uptime
	go func() {
		for {
			time.Sleep(time.Second)
			uptime := int(time.Since(start).Seconds())
			_ = w.DispatchCall("window.updateUptime", uptime)
		}
	}()

	if err := w.SetHtml(html); err != nil {
		log.Fatal(err)
	}
	if err := w.Run(); err != nil {
		log.Fatal(err)
	}
}
