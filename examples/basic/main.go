package main

import webview "github.com/samcharles93/webview_go"

func main() {
	w := webview.New(false)
	defer w.Destroy()
	w.SetTitle("Basic Example")
	w.SetSize(480, 320, webview.HintNone)
	w.SetHtml(`
		<!DOCTYPE html>
		<html>
		<head>
			<style>
				body {
					font-family: system-ui, -apple-system, sans-serif;
					display: flex;
					align-items: center;
					justify-content: center;
					height: 100vh;
					margin: 0;
					background: #f0f2f5;
				}
				h1 { color: #1c1e21; }
			</style>
		</head>
		<body>
			<h1>Thanks for using webview_go!</h1>
		</body>
		</html>
	`)
	w.Run()
}
