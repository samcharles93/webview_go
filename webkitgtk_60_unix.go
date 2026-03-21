//go:build (linux || openbsd || freebsd || netbsd) && (webkitgtk_60 || (!webkitgtk_41 && !webkitgtk_60))

package webview

/*
#cgo linux openbsd freebsd netbsd pkg-config: gtk4 webkitgtk-6.0
*/
import "C"
