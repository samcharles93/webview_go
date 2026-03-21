//go:build (linux || openbsd || freebsd || netbsd) && webkitgtk_41 && !webkitgtk_60

package webview

/*
#cgo linux openbsd freebsd netbsd pkg-config: gtk+-3.0 webkit2gtk-4.1
*/
import "C"
