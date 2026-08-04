//go:build !windows

package gamepad

// Windows XInput gamepad executor lives in gamepad_windows.go.
// MVP supports rumble (vibrate) only; virtual button injection needs a driver.
