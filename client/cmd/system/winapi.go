package system

import (
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

// `MessageBox` represents the information needed to display a message box using the Win32 API.
type MessageBox struct {
	Delay    int32   `json:"delay"`
	Amount   int32   `json:"amount"`
	Title    string  `json:"title"`
	Content  string  `json:"content"`
	Buttons  uintptr `json:"buttons,omitempty"`
	Icon     uintptr `json:"icon,omitempty"`
	Default  uintptr `json:"default,omitempty"`
	Modality uintptr `json:"modality,omitempty"`
}

// User32 DLL and its functions.
var (
	User32DLL   = syscall.NewLazyDLL("user32.dll")
	MessageBoxW = User32DLL.NewProc("MessageBoxW")
	BlockInput  = User32DLL.NewProc("BlockInput")
)

// Kernel32 DLL and its functions.
var (
	Kernel32DLL        = syscall.NewLazyDLL("kernel32.dll")
	SetThreadExecution = Kernel32DLL.NewProc("SetThreadExecutionState")
)

// `IsCurrentUserPrivileged` returns whether the current user has admin privileges.
func IsCurrentUserPrivileged() bool {
	// Call check with current user token handle.
	return IsUserPrivileged(TK_CURRENT_USER)
}

// `IsUserPrivileged` returns whether the passed user (Windows token handle) has admin privileges.
func IsUserPrivileged(token windows.Token) (result bool) {
	var sid *windows.SID

	// Read the MSDN documentation for this function to understand how it works.
	// https://docs.microsoft.com/en-us/windows/desktop/api/securitybaseapi/nf-securitybaseapi-checktokenmembership
	err := windows.AllocateAndInitializeSid(
		&windows.SECURITY_NT_AUTHORITY,
		2,
		windows.SECURITY_BUILTIN_DOMAIN_RID,
		windows.DOMAIN_ALIAS_RID_ADMINS,
		0, 0, 0, 0, 0, 0,
		&sid)
	if err != nil {
		return false
	}

	// Check if the passed account Token handle is member of SID.
	result, err = token.IsMember(sid)
	if err != nil {
		return false
	}

	return result || token.IsElevated()
}

// `ToggleInputLocker` calls the Win32 API BlockInput function to enable/disable any input calls
// from the connected keyboard and mouse.
//
// If the inputs are blocked, a call is made to the Win32 API SetThreadExecutionState
// function that stops the screen from sleeping, otherwise it is set to the default state.
func ToggleInputLocker(block bool) error {
	// Toggles all input events from the mouse and keyboard (blocks or unblocks).
	if block {
		// Block keyboard and mouse input calls.
		ret, _, err := BlockInput.Call(1)
		if ret == 0 {
			return err
		}

		// Make sure the system will not go into sleep mode by disabling movement requirements.
		SetThreadExecution.Call(uintptr(ES_CONTINUOUS | ES_SYSTEMREQUIRED | ES_DISPLAYREQUIRED))
	} else {
		// Allow keyboard and mouse input calls.
		ret, _, err := BlockInput.Call(0)
		if ret == 0 {
			return err
		}

		// Reset the ThreadExecution requirements to its default.
		// Therefore allowing the system to go back into sleep mode naturally.
		SetThreadExecution.Call(uintptr(ES_CONTINUOUS))
	}

	return nil
}

// `Win32MessageBox` calls the Win32 API MessageBoxW function to display a message box dialog
// with the supplied caption, title, and flags.
func Win32MessageBox(hWindow uintptr, caption, title string, flags uintptr) int {
	// Convert the passed caption and title strings to syscall pointers.
	captionPtr := syscall.StringToUTF16Ptr(caption)
	titlePtr := syscall.StringToUTF16Ptr(title)

	// Call the MessageBoxW function to display the message box dialog popup.
	ret, _, _ := MessageBoxW.Call(
		hWindow,
		uintptr(unsafe.Pointer(captionPtr)),
		uintptr(unsafe.Pointer(titlePtr)),
		flags)

	// The return value from MessageBoxW indicates which button the user clicked.
	return int(ret)
}

// `MessageBox.Show` displays formats and passes the current MessageBox values to the Win32MessageBox function.
func (msgbox *MessageBox) Show() {
	// Run the display loop in a goroutine so that the client can continue recieving commands.
	go func() {
		// Call Win32MessageBox MessageBox.Amount of times.
		for i := 0; i < int(msgbox.Amount); i++ {
			// We're not handling the MessageBoxW return value in this case, as it is only
			// returned after the user has closed the dialog popup, which means the client cannot further
			// recieve commands until it is closed unless we run it in a goroutine and discard the return value.
			go Win32MessageBox(0, msgbox.Content, msgbox.Title, msgbox.Buttons|msgbox.Icon|MB_TOPMOST)

			// The display loop needs to sleep so that the dialog popup doesn't spawn in the same position
			// as the previous one. It sleeps for MessageBox.Delay Milliseconds before moving on.
			time.Sleep(time.Millisecond * time.Duration(msgbox.Delay))
		}
	}()
}

// Win32 API Account Token handles.
const (
	TK_CURRENT_USER windows.Token = windows.Token(0)
)

// Win32 API MessageBoxW Button flags.
const (
	MB_OK                uintptr = 0x00000000
	MB_OKCANCEL          uintptr = 0x00000001
	MB_ABORTRETRYIGNORE  uintptr = 0x00000002
	MB_YESNOCANCEL       uintptr = 0x00000003
	MB_YESNO             uintptr = 0x00000004
	MB_RETRYCANCEL       uintptr = 0x00000005
	MB_CANCELTRYCONTINUE uintptr = 0x00000006
)

// Win32 API MessageBoxW Icon flags.
const (
	MB_ICONNULL        uintptr = 0x00000000
	MB_ICONERROR       uintptr = 0x00000010
	MB_ICONQUESTION    uintptr = 0x00000020
	MB_ICONWARNING     uintptr = 0x00000030
	MB_ICONINFORMATION uintptr = 0x00000040
)

// Win32 API MessageBoxW Alignment flags
const (
	MB_TOPMOST uintptr = 0x00040000
)

// Win32 API BlockInput flags.
const (
	ES_CONTINUOUS      uintptr = 0x80000000
	ES_SYSTEMREQUIRED  uintptr = 0x00000001
	ES_DISPLAYREQUIRED uintptr = 0x00000002
)
