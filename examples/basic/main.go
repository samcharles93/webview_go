package main

import (
	"log"

	webview "github.com/samcharles93/webview_go"
)

func main() {
	w, err := webview.NewWithOptions(webview.Options{
		Title:  "Basic Example",
		Width:  480,
		Height: 320,
		Hint:   webview.HintNone,
		HTML: `
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
	`,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer w.Destroy()

	if err := w.Run(); err != nil {
		log.Fatal(err)
	}
}
