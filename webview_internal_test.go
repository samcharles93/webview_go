package webview

import (
	"errors"
	"testing"
)

func TestErrorIs(t *testing.T) {
	err := newError("set title", ErrorCodeInvalidState, "webview is destroyed")
	if !errors.Is(err, ErrInvalidState) {
		t.Fatalf("expected errors.Is(..., ErrInvalidState) to match, got %v", err)
	}
}

func TestValidateOptions(t *testing.T) {
	tests := []struct {
		name    string
		opts    Options
		wantErr error
	}{
		{
			name:    "html and url conflict",
			opts:    Options{HTML: "<p>x</p>", URL: "https://example.com"},
			wantErr: ErrInvalidArgument,
		},
		{
			name:    "width without height",
			opts:    Options{Width: 640},
			wantErr: ErrInvalidArgument,
		},
		{
			name:    "height without width",
			opts:    Options{Height: 480},
			wantErr: ErrInvalidArgument,
		},
		{
			name: "valid empty options",
			opts: Options{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validateOptions(tc.opts)
			if tc.wantErr == nil {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("expected %v, got %v", tc.wantErr, err)
			}
		})
	}
}

func TestMarshalJS(t *testing.T) {
	js, err := MarshalJS(map[string]any{"count": 2, "label": "ok"})
	if err != nil {
		t.Fatalf("MarshalJS returned error: %v", err)
	}
	if js != `{"count":2,"label":"ok"}` && js != `{"label":"ok","count":2}` {
		t.Fatalf("unexpected JSON: %s", js)
	}
}

func TestBuildJSCall(t *testing.T) {
	js, err := buildJSCall("window.updateStats", map[string]any{"count": 2}, []int{1, 2})
	if err != nil {
		t.Fatalf("buildJSCall returned error: %v", err)
	}
	if js != `(window.updateStats)({"count":2}, [1,2])` &&
		js != `(window.updateStats)({"count":2}, [1, 2])` {
		t.Fatalf("unexpected JS call: %s", js)
	}
}

func TestBuildJSCallEmptyFunction(t *testing.T) {
	_, err := buildJSCall("   ", 1)
	if !errors.Is(err, ErrInvalidArgument) {
		t.Fatalf("expected invalid argument error, got %v", err)
	}
}

func TestMakeBindingFunc(t *testing.T) {
	binding, err := makeBindingFunc(func(a int, b string) (string, error) {
		return b + "-" + string(rune('0'+a)), nil
	})
	if err != nil {
		t.Fatalf("makeBindingFunc returned error: %v", err)
	}

	res, err := binding("1", `[2,"ok"]`)
	if err != nil {
		t.Fatalf("binding returned error: %v", err)
	}
	if got := res.(string); got != "ok-2" {
		t.Fatalf("unexpected binding result: %q", got)
	}
}

func TestMakeBindingFuncVariadic(t *testing.T) {
	binding, err := makeBindingFunc(func(prefix string, values ...int) int {
		sum := len(prefix)
		for _, v := range values {
			sum += v
		}
		return sum
	})
	if err != nil {
		t.Fatalf("makeBindingFunc returned error: %v", err)
	}

	res, err := binding("1", `["x",1,2,3]`)
	if err != nil {
		t.Fatalf("binding returned error: %v", err)
	}
	if got := res.(int); got != 7 {
		t.Fatalf("unexpected variadic result: %d", got)
	}
}

func TestMakeBindingFuncErrors(t *testing.T) {
	t.Run("non function", func(t *testing.T) {
		_, err := makeBindingFunc(123)
		if err == nil {
			t.Fatal("expected error for non-function binding")
		}
	})

	t.Run("too many returns", func(t *testing.T) {
		_, err := makeBindingFunc(func() (int, int, error) { return 0, 0, nil })
		if err == nil {
			t.Fatal("expected error for too many return values")
		}
	})

	t.Run("bad args", func(t *testing.T) {
		binding, err := makeBindingFunc(func(v int) {})
		if err != nil {
			t.Fatalf("unexpected setup error: %v", err)
		}
		if _, err := binding("1", `["wrong"]`); err == nil {
			t.Fatal("expected argument unmarshal error")
		}
	})

	t.Run("second return must be error", func(t *testing.T) {
		binding, err := makeBindingFunc(func() (int, string) { return 1, "bad" })
		if err != nil {
			t.Fatalf("unexpected setup error: %v", err)
		}
		if _, err := binding("1", `[]`); err == nil {
			t.Fatal("expected invalid second return type error")
		}
	})

	t.Run("error only", func(t *testing.T) {
		binding, err := makeBindingFunc(func() error { return errors.New("boom") })
		if err != nil {
			t.Fatalf("unexpected setup error: %v", err)
		}
		if _, err := binding("1", `[]`); err == nil || err.Error() != "boom" {
			t.Fatalf("expected binding error, got %v", err)
		}
	})
}
