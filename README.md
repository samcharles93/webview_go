# webview_go

[![Go Reference](https://pkg.go.dev/badge/github.com/samcharles93/webview_go.svg)](https://pkg.go.dev/github.com/samcharles93/webview_go)
[![Go Report Card](https://goreportcard.com/badge/github.com/samcharles93/webview_go)](https://goreportcard.com/report/github.com/samcharles93/webview_go)

Go language binding for the [webview library][webview].

> [!NOTE]
> Versions <= 0.1.1 are available in the [old repository][old-repo].

### Getting Started

See [Go package documentation][go-docs] for the Go API documentation, or simply read the source code.

Start with creating a new directory structure for your project.

```sh
mkdir my-project && cd my-project
```

Create a new Go module.

```sh
go mod init example.com/app
```

Save one of the example programs into your project directory.

```sh
curl -sSLo main.go "https://raw.githubusercontent.com/samcharles93/webview_go/main/examples/basic/main.go"
```

Install dependencies.

```sh
go get github.com/samcharles93/webview_go
```

Create a window with the recommended constructor.

```go
w, err := webview.NewWithOptions(webview.Options{
    Title:  "Hello",
    Width:  800,
    Height: 600,
    Hint:   webview.HintNone,
    HTML:   "<h1>Hello from Go</h1>",
})
if err != nil {
    log.Fatal(err)
}
defer w.Destroy()

if err := w.Run(); err != nil {
    log.Fatal(err)
}
```

Build the example.

```sh
go build
```

On Windows, add `-ldflags="-H windowsgui"` to build a GUI executable:

```sh
go build -ldflags="-H windowsgui"
```

### API highlights

- `NewWithOptions(...)` returns a typed error instead of failing with a nil `WebView`
- `MarshalJS(...)`, `Call(...)`, and `DispatchCall(...)` help avoid hand-built `Eval(fmt.Sprintf(...))`
- `errors.Is(err, webview.ErrMissingDependency)` and the other typed error sentinels work with constructor and runtime failures

### Linux and BSD dependencies

On Linux, OpenBSD, FreeBSD, and NetBSD, pick one supported GTK/WebKitGTK pair:

- Default: `gtk4` + `webkitgtk-6.0`
- Optional: `gtk+-3.0` + `webkit2gtk-4.1`

The default build targets `webkitgtk-6.0`. You can also select either backend explicitly with a build tag:

```sh
go build -tags webkitgtk_41
go build -tags webkitgtk_60
```

On Debian/Ubuntu the development packages are typically:

```sh
# Default backend
sudo apt-get install libgtk-4-dev libwebkitgtk-6.0-dev

# Optional legacy backend
sudo apt-get install libgtk-3-dev libwebkit2gtk-4.1-dev
```

### Windows requirements

Windows builds require a C++14-capable compiler and the Windows SDK headers
that the selected toolchain normally provides.

Supported toolchains:

- MSVC / Visual Studio 2022 or later
- GNU-family toolchains such as LLVM MinGW, MSYS2 MinGW-w64, or `zig cc`/`zig c++`

This module vendors the upstream `webview.h` amalgamation plus the WebView2 SDK
headers needed by the Windows backend, including the MinGW compatibility
`EventToken.h` shim required by some GNU-family toolchains.

Applications still require the Microsoft Edge WebView2 runtime on end-user
systems before Windows 11.

Example Windows GNU-family cross-compile from Linux:

```sh
GOOS=windows GOARCH=amd64 CGO_ENABLED=1 \
CC='zig cc -target x86_64-windows-gnu' \
CXX='zig c++ -target x86_64-windows-gnu' \
go build -ldflags="-H windowsgui"
```

### Vendored dependencies

Vendored third-party code is pinned in:

- `libs/webview/version.txt` for the upstream `webview` revision
- `libs/mswebview2/version.txt` for the Microsoft WebView2 SDK header version

When updating either dependency, refresh the vendored headers together and keep
the pinned version files in sync.

### Notes

Calling `Eval()` or `Dispatch()` before `Run()` does not work because the webview instance has only been configured and not yet started.

If a Linux build fails with `pkg-config` errors, install the GTK/WebKitGTK
development package pair that matches the backend you are building.

If a Windows build fails with missing WebView2 headers while using a GNU-family
toolchain, verify that cgo is compiling against this module's vendored headers
rather than a partial SDK installation.

[go-docs]: https://pkg.go.dev/github.com/samcharles93/webview_go
[old-repo]: https://github.com/webview/webview_go
[webview]: https://github.com/webview/webview
