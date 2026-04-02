package webview

import (
	"errors"
	"log"
	"os"
	"testing"
)

func TestUISmoke(t *testing.T) {
	if os.Getenv("WEBVIEW_GO_RUN_UI_TESTS") != "1" {
		t.Skip("set WEBVIEW_GO_RUN_UI_TESTS=1 to run UI smoke tests")
	}

	w, err := NewWithOptions(Options{
		Debug:  true,
		Title:  "Hello",
		Width:  480,
		Height: 320,
		Hint:   HintNone,
	})
	if err != nil {
		if errors.Is(err, ErrMissingDependency) {
			t.Skipf("missing runtime dependency: %v", err)
		}
		t.Fatalf("create webview: %v", err)
	}
	defer w.Destroy()

	if err := w.Bind("noop", func() string {
		log.Println("hello")
		return "hello"
	}); err != nil {
		t.Fatalf("bind noop: %v", err)
	}
	if err := w.Bind("add", func(a, b int) int {
		return a + b
	}); err != nil {
		t.Fatalf("bind add: %v", err)
	}
	if err := w.Bind("quit", func() error {
		return w.Terminate()
	}); err != nil {
		t.Fatalf("bind quit: %v", err)
	}

	if err := w.SetHtml(`<!doctype html>
		<html>
			<body>hello</body>
			<script>
				window.onload = function() {
					document.body.innerText = ` + "`hello, ${navigator.userAgent}`" + `;
					noop().then(function(res) {
						console.log('noop res', res);
						add(1, 2).then(function(res) {
							console.log('add res', res);
							quit();
						});
					});
				};
			</script>
		</html>
	`); err != nil {
		t.Fatalf("set html: %v", err)
	}
	if err := w.Run(); err != nil {
		t.Fatalf("run: %v", err)
	}
}
