package webview

/*
#cgo CFLAGS: -I${SRCDIR}/libs/webview/include
#cgo CXXFLAGS: -I${SRCDIR}/libs/webview/include -DWEBVIEW_STATIC

#cgo linux openbsd freebsd netbsd CXXFLAGS: -DWEBVIEW_GTK -std=c++11
#cgo linux openbsd freebsd netbsd LDFLAGS: -ldl
#cgo darwin CXXFLAGS: -DWEBVIEW_COCOA -std=c++11
#cgo darwin LDFLAGS: -framework WebKit -ldl

#cgo windows CXXFLAGS: -DWEBVIEW_EDGE -std=c++14 -I${SRCDIR}/libs/mswebview2/include
#cgo windows LDFLAGS: -static -ladvapi32 -lole32 -lshell32 -lshlwapi -luser32 -lversion

#include "webview.h"

#include <stdlib.h>
#include <stdint.h>

typedef webview_error_t (*dispatch_fn)(webview_t w, void (*fn)(webview_t w, void *arg), void *arg);

webview_error_t CgoWebViewDispatch(webview_t w, uintptr_t arg);
webview_error_t CgoWebViewBind(webview_t w, const char *name, uintptr_t arg);
webview_error_t CgoWebViewUnbind(webview_t w, const char *name);
*/
import "C"
import (
	"encoding/json"
	"errors"
	"fmt"
	_ "github.com/samcharles93/webview_go/libs/mswebview2"
	_ "github.com/samcharles93/webview_go/libs/mswebview2/include"
	_ "github.com/samcharles93/webview_go/libs/webview"
	_ "github.com/samcharles93/webview_go/libs/webview/include"
	"reflect"
	"runtime"
	"runtime/cgo"
	"strings"
	"sync"
	"unsafe"
)

func init() {
	// Ensure that main.main is called from the main thread
	runtime.LockOSThread()
}

// Hints are used to configure window sizing and resizing
type Hint int

const (
	// Width and height are default size
	HintNone Hint = C.WEBVIEW_HINT_NONE

	// Window size can not be changed by a user
	HintFixed Hint = C.WEBVIEW_HINT_FIXED

	// Width and height are minimum bounds
	HintMin Hint = C.WEBVIEW_HINT_MIN

	// Width and height are maximum bounds
	HintMax Hint = C.WEBVIEW_HINT_MAX
)

// NativeHandleKind specifies the kind of native handle to retrieve.
type NativeHandleKind int

const (
	// Top-level window. GtkWindow pointer (GTK), NSWindow pointer (Cocoa) or HWND (Win32).
	NativeHandleKindUIWindow NativeHandleKind = C.WEBVIEW_NATIVE_HANDLE_KIND_UI_WINDOW
	// Browser widget. GtkWidget pointer (GTK), NSView pointer (Cocoa) or HWND (Win32).
	NativeHandleKindUIWidget NativeHandleKind = C.WEBVIEW_NATIVE_HANDLE_KIND_UI_WIDGET
	// Browser controller. WebKitWebView pointer (GTK), WKWebView pointer (Cocoa) or ICoreWebView2Controller pointer (Win32/Edge).
	NativeHandleKindBrowserController NativeHandleKind = C.WEBVIEW_NATIVE_HANDLE_KIND_BROWSER_CONTROLLER
)

// VersionInfo holds the library's version information.
type VersionInfo struct {
	Major         int
	Minor         int
	Patch         int
	VersionNumber string
	PreRelease    string
	BuildMetadata string
}

// ErrorCode is a typed wrapper around the upstream webview error codes.
type ErrorCode int

const (
	ErrorCodeOK                ErrorCode = C.WEBVIEW_ERROR_OK
	ErrorCodeUnspecified       ErrorCode = C.WEBVIEW_ERROR_UNSPECIFIED
	ErrorCodeMissingDependency ErrorCode = C.WEBVIEW_ERROR_MISSING_DEPENDENCY
	ErrorCodeCanceled          ErrorCode = C.WEBVIEW_ERROR_CANCELED
	ErrorCodeInvalidState      ErrorCode = C.WEBVIEW_ERROR_INVALID_STATE
	ErrorCodeInvalidArgument   ErrorCode = C.WEBVIEW_ERROR_INVALID_ARGUMENT
	ErrorCodeDuplicate         ErrorCode = C.WEBVIEW_ERROR_DUPLICATE
	ErrorCodeNotFound          ErrorCode = C.WEBVIEW_ERROR_NOT_FOUND
)

func (c ErrorCode) String() string {
	switch c {
	case ErrorCodeOK:
		return "ok"
	case ErrorCodeMissingDependency:
		return "missing dependency"
	case ErrorCodeCanceled:
		return "operation canceled"
	case ErrorCodeInvalidState:
		return "invalid state"
	case ErrorCodeInvalidArgument:
		return "invalid argument"
	case ErrorCodeDuplicate:
		return "duplicate"
	case ErrorCodeNotFound:
		return "not found"
	case ErrorCodeUnspecified:
		return "unspecified error"
	default:
		return fmt.Sprintf("unknown error (%d)", int(c))
	}
}

// Error is a typed error returned by library operations.
type Error struct {
	Op     string
	Code   ErrorCode
	Detail string
}

func (e *Error) Error() string {
	if e == nil {
		return "<nil>"
	}
	msg := e.Code.String()
	if e.Detail != "" {
		msg += ": " + e.Detail
	}
	if e.Op != "" {
		return e.Op + ": " + msg
	}
	return msg
}

func (e *Error) Is(target error) bool {
	t, ok := target.(*Error)
	return ok && e.Code == t.Code
}

var (
	ErrUnspecified       = &Error{Code: ErrorCodeUnspecified}
	ErrMissingDependency = &Error{Code: ErrorCodeMissingDependency}
	ErrCanceled          = &Error{Code: ErrorCodeCanceled}
	ErrInvalidState      = &Error{Code: ErrorCodeInvalidState}
	ErrInvalidArgument   = &Error{Code: ErrorCodeInvalidArgument}
	ErrDuplicate         = &Error{Code: ErrorCodeDuplicate}
	ErrNotFound          = &Error{Code: ErrorCodeNotFound}
)

// Options configures webview construction and common initial setup.
type Options struct {
	Debug       bool
	Window      unsafe.Pointer
	Title       string
	Width       int
	Height      int
	Hint        Hint
	HTML        string
	URL         string
	InitScripts []string
}

type WebView interface {

	// Run runs the main loop until it's terminated. After this function exits -
	// you must destroy the webview.
	Run() error

	// Terminate stops the main loop. It is safe to call this function from
	// a background thread.
	Terminate() error

	// Dispatch posts a function to be executed on the main thread. You normally
	// do not need to call this function, unless you want to tweak the native
	// window.
	Dispatch(f func()) error

	// Destroy destroys a webview and closes the native window.
	Destroy() error

	// Window returns a native window handle pointer. When using GTK backend the
	// pointer is GtkWindow pointer, when using Cocoa backend the pointer is
	// NSWindow pointer, when using Win32 backend the pointer is HWND pointer.
	// Deprecated: Use NativeHandle(NativeHandleKindUIWindow) instead.
	Window() unsafe.Pointer

	// NativeHandle returns a native handle of the given kind.
	NativeHandle(kind NativeHandleKind) unsafe.Pointer

	// SetTitle updates the title of the native window. Must be called from the UI
	// thread.
	SetTitle(title string) error

	// SetSize updates native window size. See Hint constants.
	SetSize(w int, h int, hint Hint) error

	// Navigate navigates webview to the given URL. URL may be a properly encoded data.
	// URI. Examples:
	// w.Navigate("https://github.com/webview/webview")
	// w.Navigate("data:text/html,%3Ch1%3EHello%3C%2Fh1%3E")
	// w.Navigate("data:text/html;base64,PGgxPkhlbGxvPC9oMT4=")
	Navigate(url string) error

	// SetHtml sets the webview HTML directly.
	// Example: w.SetHtml(w, "<h1>Hello</h1>");
	SetHtml(html string) error

	// Init injects JavaScript code at the initialization of the new page. Every
	// time the webview will open a the new page - this initialization code will
	// be executed. It is guaranteed that code is executed before window.onload.
	Init(js string) error

	// Eval evaluates arbitrary JavaScript code. Evaluation happens asynchronously,
	// also the result of the expression is ignored. Use RPC bindings if you want
	// to receive notifications about the results of the evaluation.
	Eval(js string) error

	// Call evaluates a JavaScript function call built from JSON-marshaled arguments.
	// function is a JavaScript expression resolving to a callable value, such as
	// "window.updateStats" or "globalThis.app.refresh".
	Call(function string, args ...any) error

	// DispatchCall posts a JavaScript function call to the UI thread.
	DispatchCall(function string, args ...any) error

	// Bind binds a callback function so that it will appear under the given name
	// as a global JavaScript function. Internally it uses webview_init().
	// Callback receives a request string and a user-provided argument pointer.
	// Request string is a JSON array of all the arguments passed to the
	// JavaScript function.
	//
	// f must be a function
	// f must return either value and error or just error
	Bind(name string, f any) error

	// Removes a callback that was previously set by Bind.
	Unbind(name string) error
}

type webview struct {
	w        C.webview_t
	bindings map[string]cgo.Handle
	mu       sync.Mutex
}

type bindingFunc func(id, req string) (any, error)

type bindingInfo struct {
	w *webview
	f bindingFunc
}

func boolToInt(b bool) C.int {
	if b {
		return 1
	}
	return 0
}

func newError(op string, code ErrorCode, detail string) error {
	if code == ErrorCodeOK {
		return nil
	}
	return &Error{Op: op, Code: code, Detail: detail}
}

func webviewError(op string, err C.webview_error_t) error {
	if err == C.WEBVIEW_ERROR_OK {
		return nil
	}
	return newError(op, ErrorCode(err), "")
}

func validateOptions(opts Options) error {
	if opts.HTML != "" && opts.URL != "" {
		return newError("new with options", ErrorCodeInvalidArgument, "html and url cannot both be set")
	}
	if (opts.Width == 0) != (opts.Height == 0) {
		return newError("new with options", ErrorCodeInvalidArgument, "width and height must both be set")
	}
	return nil
}

// MarshalJS marshals a Go value into a JavaScript literal using JSON encoding.
func MarshalJS(v any) (string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func buildJSCall(function string, args ...any) (string, error) {
	function = strings.TrimSpace(function)
	if function == "" {
		return "", newError("build js call", ErrorCodeInvalidArgument, "function cannot be empty")
	}
	argv := make([]string, 0, len(args))
	for _, arg := range args {
		js, err := MarshalJS(arg)
		if err != nil {
			return "", err
		}
		argv = append(argv, js)
	}
	return fmt.Sprintf("(%s)(%s)", function, strings.Join(argv, ", ")), nil
}

func makeBindingFunc(f any) (bindingFunc, error) {
	v := reflect.ValueOf(f)
	if v.Kind() != reflect.Func {
		return nil, errors.New("only functions can be bound")
	}
	if n := v.Type().NumOut(); n > 2 {
		return nil, errors.New("function may only return a value or a value+error")
	}

	return func(id, req string) (any, error) {
		raw := []json.RawMessage{}
		if err := json.Unmarshal([]byte(req), &raw); err != nil {
			return nil, err
		}

		isVariadic := v.Type().IsVariadic()
		numIn := v.Type().NumIn()
		if (isVariadic && len(raw) < numIn-1) || (!isVariadic && len(raw) != numIn) {
			return nil, errors.New("function arguments mismatch")
		}
		args := []reflect.Value{}
		for i := range raw {
			var arg reflect.Value
			if isVariadic && i >= numIn-1 {
				arg = reflect.New(v.Type().In(numIn - 1).Elem())
			} else {
				arg = reflect.New(v.Type().In(i))
			}
			if err := json.Unmarshal(raw[i], arg.Interface()); err != nil {
				return nil, err
			}
			args = append(args, arg.Elem())
		}
		errorType := reflect.TypeOf((*error)(nil)).Elem()
		res := v.Call(args)
		switch len(res) {
		case 0:
			return nil, nil
		case 1:
			if res[0].Type().Implements(errorType) {
				if res[0].Interface() != nil {
					return nil, res[0].Interface().(error)
				}
				return nil, nil
			}
			return res[0].Interface(), nil
		case 2:
			if !res[1].Type().Implements(errorType) {
				return nil, errors.New("second return value must be an error")
			}
			if res[1].Interface() == nil {
				return res[0].Interface(), nil
			}
			return res[0].Interface(), res[1].Interface().(error)
		default:
			return nil, errors.New("unexpected number of return values")
		}
	}, nil
}

func (w *webview) cHandle(op string) (C.webview_t, error) {
	if w == nil {
		return nil, newError(op, ErrorCodeInvalidState, "webview is nil")
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.w == nil || w.bindings == nil {
		return nil, newError(op, ErrorCodeInvalidState, "webview is destroyed")
	}
	return w.w, nil
}

func (w *webview) applyOptions(opts Options) error {
	if opts.Title != "" {
		if err := w.SetTitle(opts.Title); err != nil {
			return err
		}
	}
	if opts.Width != 0 {
		if err := w.SetSize(opts.Width, opts.Height, opts.Hint); err != nil {
			return err
		}
	}
	for _, js := range opts.InitScripts {
		if err := w.Init(js); err != nil {
			return err
		}
	}
	switch {
	case opts.HTML != "":
		return w.SetHtml(opts.HTML)
	case opts.URL != "":
		return w.Navigate(opts.URL)
	default:
		return nil
	}
}

// NewWithOptions creates a new webview and applies common initial configuration.
func NewWithOptions(opts Options) (WebView, error) {
	if err := validateOptions(opts); err != nil {
		return nil, err
	}
	w := &webview{
		bindings: make(map[string]cgo.Handle),
	}
	w.w = C.webview_create(boolToInt(opts.Debug), opts.Window)
	if w.w == nil {
		return nil, newError("create", ErrorCodeUnspecified, "webview_create returned nil")
	}
	if err := w.applyOptions(opts); err != nil {
		_ = C.webview_destroy(w.w)
		w.w = nil
		w.bindings = nil
		return nil, err
	}
	return w, nil
}

// New calls NewWindow to create a new window and a new webview instance. If debug
// is non-zero - developer tools will be enabled (if the platform supports them).
func New(debug bool) WebView {
	w, _ := NewWithOptions(Options{Debug: debug})
	return w
}

// NewWindow creates a new webview instance. If debug is non-zero - developer
// tools will be enabled (if the platform supports them). Window parameter can be
// a pointer to the native window handle. If it's non-null - then child WebView is
// embedded into the given parent window. Otherwise a new window is created.
// Depending on the platform, a GtkWindow, NSWindow or HWND pointer can be passed
// here.
func NewWindow(debug bool, window unsafe.Pointer) WebView {
	w, _ := NewWithOptions(Options{Debug: debug, Window: window})
	return w
}

// Version returns the library's version information.
func Version() VersionInfo {
	info := C.webview_version()
	return VersionInfo{
		Major:         int(info.version.major),
		Minor:         int(info.version.minor),
		Patch:         int(info.version.patch),
		VersionNumber: C.GoString(&info.version_number[0]),
		PreRelease:    C.GoString(&info.pre_release[0]),
		BuildMetadata: C.GoString(&info.build_metadata[0]),
	}
}

func (w *webview) Destroy() error {
	w.mu.Lock()
	if w.w == nil || w.bindings == nil {
		w.mu.Unlock()
		return newError("destroy", ErrorCodeInvalidState, "webview is destroyed")
	}
	handle := w.w
	bindings := w.bindings
	w.mu.Unlock()

	if err := webviewError("destroy", C.webview_destroy(handle)); err != nil {
		return err
	}

	w.mu.Lock()
	for _, h := range bindings {
		h.Delete()
	}
	w.bindings = nil
	w.w = nil
	w.mu.Unlock()
	return nil
}

func (w *webview) Run() error {
	handle, err := w.cHandle("run")
	if err != nil {
		return err
	}
	return webviewError("run", C.webview_run(handle))
}

func (w *webview) Terminate() error {
	handle, err := w.cHandle("terminate")
	if err != nil {
		return err
	}
	return webviewError("terminate", C.webview_terminate(handle))
}

func (w *webview) Window() unsafe.Pointer {
	handle, err := w.cHandle("window")
	if err != nil {
		return nil
	}
	return C.webview_get_window(handle)
}

func (w *webview) NativeHandle(kind NativeHandleKind) unsafe.Pointer {
	handle, err := w.cHandle("native handle")
	if err != nil {
		return nil
	}
	return C.webview_get_native_handle(handle, C.webview_native_handle_kind_t(kind))
}

func (w *webview) Navigate(url string) error {
	handle, err := w.cHandle("navigate")
	if err != nil {
		return err
	}
	s := C.CString(url)
	defer C.free(unsafe.Pointer(s))
	return webviewError("navigate", C.webview_navigate(handle, s))
}

func (w *webview) SetHtml(html string) error {
	handle, err := w.cHandle("set html")
	if err != nil {
		return err
	}
	s := C.CString(html)
	defer C.free(unsafe.Pointer(s))
	return webviewError("set html", C.webview_set_html(handle, s))
}

func (w *webview) SetTitle(title string) error {
	handle, err := w.cHandle("set title")
	if err != nil {
		return err
	}
	s := C.CString(title)
	defer C.free(unsafe.Pointer(s))
	return webviewError("set title", C.webview_set_title(handle, s))
}

func (w *webview) SetSize(width int, height int, hint Hint) error {
	handle, err := w.cHandle("set size")
	if err != nil {
		return err
	}
	return webviewError("set size", C.webview_set_size(handle, C.int(width), C.int(height), C.webview_hint_t(hint)))
}

func (w *webview) Init(js string) error {
	handle, err := w.cHandle("init")
	if err != nil {
		return err
	}
	s := C.CString(js)
	defer C.free(unsafe.Pointer(s))
	return webviewError("init", C.webview_init(handle, s))
}

func (w *webview) Eval(js string) error {
	handle, err := w.cHandle("eval")
	if err != nil {
		return err
	}
	s := C.CString(js)
	defer C.free(unsafe.Pointer(s))
	return webviewError("eval", C.webview_eval(handle, s))
}

func (w *webview) Call(function string, args ...any) error {
	js, err := buildJSCall(function, args...)
	if err != nil {
		return err
	}
	return w.Eval(js)
}

func (w *webview) DispatchCall(function string, args ...any) error {
	js, err := buildJSCall(function, args...)
	if err != nil {
		return err
	}
	return w.Dispatch(func() {
		_ = w.Eval(js)
	})
}

func (w *webview) Dispatch(f func()) error {
	handle, err := w.cHandle("dispatch")
	if err != nil {
		return err
	}
	h := cgo.NewHandle(f)
	if err := webviewError("dispatch", C.CgoWebViewDispatch(handle, C.uintptr_t(h))); err != nil {
		h.Delete()
		return err
	}
	return nil
}

//export _webviewDispatchGoCallback
func _webviewDispatchGoCallback(h C.uintptr_t) {
	handle := cgo.Handle(h)
	f := handle.Value().(func())
	handle.Delete()
	f()
}

//export _webviewBindingGoCallback
func _webviewBindingGoCallback(id *C.char, req *C.char, h C.uintptr_t) {
	handle := cgo.Handle(h)
	info := handle.Value().(*bindingInfo)

	jsString := func(v any) string { b, _ := json.Marshal(v); return string(b) }
	status, result := 0, ""
	if res, err := info.f(C.GoString(id), C.GoString(req)); err != nil {
		status = -1
		result = jsString(err.Error())
	} else if b, err := json.Marshal(res); err != nil {
		status = -1
		result = jsString(err.Error())
	} else {
		status = 0
		result = string(b)
	}
	s := C.CString(result)
	defer C.free(unsafe.Pointer(s))
	if handle, err := info.w.cHandle("return"); err == nil {
		C.webview_return(handle, id, C.int(status), s)
	}
}

func (w *webview) Bind(name string, f any) error {
	handle, err := w.cHandle("bind")
	if err != nil {
		return err
	}
	binding, err := makeBindingFunc(f)
	if err != nil {
		return err
	}

	h := cgo.NewHandle(&bindingInfo{w: w, f: binding})
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	var oldHandle cgo.Handle
	var hasOld bool

	w.mu.Lock()
	if current, ok := w.bindings[name]; ok {
		oldHandle = current
		hasOld = true
	}
	w.mu.Unlock()

	if hasOld {
		if err := webviewError("unbind", C.CgoWebViewUnbind(handle, cname)); err != nil {
			h.Delete()
			return err
		}
	}
	if err := webviewError("bind", C.CgoWebViewBind(handle, cname, C.uintptr_t(h))); err != nil {
		if hasOld {
			_ = C.CgoWebViewBind(handle, cname, C.uintptr_t(oldHandle))
		}
		h.Delete()
		return err
	}

	w.mu.Lock()
	w.bindings[name] = h
	w.mu.Unlock()
	if hasOld {
		oldHandle.Delete()
	}
	return nil
}

func (w *webview) Unbind(name string) error {
	handle, err := w.cHandle("unbind")
	if err != nil {
		return err
	}
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	err = webviewError("unbind", C.CgoWebViewUnbind(handle, cname))
	if err != nil && !errors.Is(err, ErrNotFound) {
		return err
	}

	w.mu.Lock()
	if h, ok := w.bindings[name]; ok {
		h.Delete()
		delete(w.bindings, name)
	}
	w.mu.Unlock()
	return err
}
