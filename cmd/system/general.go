package system

import (
	"os"
	"syscall"
	"unsafe"
)

// `ExitGracefully` will execute the supplied function before exiting the program.
func ExitGracefully(callOnExit func()) {
	callOnExit() // Execute the passed function.
	os.Exit(0)   // Exit the program with a successful exit code.
}

// `Win32MessageBox` calls the Win32 API MessageBoxW function.
// This is used for previewing the MessageBox that will be sent to the client on the server side.
func Win32MessageBox(hWindow uintptr, caption, title string, flags uintptr) int {
	// Convert the passed caption and title strings to syscall pointers.
	captionPtr := syscall.StringToUTF16Ptr(caption)
	titlePtr := syscall.StringToUTF16Ptr(title)

	ret, _, _ := syscall.NewLazyDLL("user32.dll").NewProc("MessageBoxW").Call(
		hWindow,
		uintptr(unsafe.Pointer(captionPtr)),
		uintptr(unsafe.Pointer(titlePtr)),
		flags)

	return int(ret)
}
