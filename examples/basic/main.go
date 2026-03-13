package main

import (
	"log"

	webview "github.com/samcharles93/webview_go"
)

func main() {
	w := webview.New(false)
	defer w.Destroy()

	if err := w.SetSize(480, 320, webview.HintNone); err != nil {
		log.Fatal(err)
	}
	if err := w.SetTitle("Basic Example"); err != nil {
		log.Fatal(err)
	}
	if err := w.SetHtml(`
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
	`); err != nil {
		log.Fatal(err)
	}

	if err := w.Run(); err != nil {
		log.Fatal(err)
	}
}
