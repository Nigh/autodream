//go:build windows

package dxgi

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	displayDeviceActive = 0x00000001
	enumCurrentSettings = 0xFFFFFFFF
	cchDeviceName       = 32
	cchFormName         = 32
)

var (
	user32                   = windows.NewLazySystemDLL("user32.dll")
	shcore                   = windows.NewLazySystemDLL("shcore.dll")
	procMonitorFromRect      = user32.NewProc("MonitorFromRect")
	procGetDpiForMonitor     = shcore.NewProc("GetDpiForMonitor")
	procEnumDisplayDevicesW  = user32.NewProc("EnumDisplayDevicesW")
	procEnumDisplaySettingsW = user32.NewProc("EnumDisplaySettingsW")
)

type displayDeviceW struct {
	Size         uint32
	DeviceName   [cchDeviceName]uint16
	DeviceString [128]uint16
	StateFlags   uint32
	DeviceID     [128]uint16
	DeviceKey    [128]uint16
}

// DEVMODEW — only PelsWidth/Height are read; layout matches Win32.
type devModeW struct {
	DeviceName       [cchDeviceName]uint16
	SpecVersion      uint16
	DriverVersion    uint16
	Size             uint16
	DriverExtra      uint16
	Fields           uint32
	Orientation      int16
	PaperSize        int16
	PaperLength      int16
	PaperWidth       int16
	Scale            int16
	Copies           int16
	DefaultSource    int16
	PrintQuality     int16
	Color            int16
	Duplex           int16
	YResolution      int16
	TTOption         int16
	Collate          int16
	FormName         [cchFormName]uint16
	LogPixels        uint16
	BitsPerPel       uint32
	PelsWidth        uint32
	PelsHeight       uint32
	DisplayFlags     uint32
	DisplayFrequency uint32
	ICMMethod        uint32
	ICMIntent        uint32
	MediaType        uint32
	DitherType       uint32
	Reserved1        uint32
	Reserved2        uint32
	PanningWidth     uint32
	PanningHeight    uint32
}

// monitorDPI returns effective DPI for the monitor covering rect, or 96 on failure.
func monitorDPI(left, top, right, bottom int) uint32 {
	r := windows.Rect{
		Left: int32(left), Top: int32(top),
		Right: int32(right), Bottom: int32(bottom),
	}
	// MONITOR_DEFAULTTONEAREST = 2
	hmon, _, _ := procMonitorFromRect.Call(uintptr(unsafe.Pointer(&r)), 2)
	if hmon == 0 {
		return 96
	}
	var dpiX, dpiY uint32
	// MDT_EFFECTIVE_DPI = 0
	hr, _, _ := procGetDpiForMonitor.Call(hmon, 0, uintptr(unsafe.Pointer(&dpiX)), uintptr(unsafe.Pointer(&dpiY)))
	if hr != 0 || dpiX == 0 {
		return 96
	}
	return dpiX
}

// displaySize returns the current mode of the Nth active display (physical pixels).
// Used when DXGI GetSize returns 0 after open or a DPI/mode switch.
func displaySize(outputIndex uint) (int, int, bool) {
	var device displayDeviceW
	device.Size = uint32(unsafe.Sizeof(device))
	var active uint32
	for i := uint32(0); ; i++ {
		r, _, _ := procEnumDisplayDevicesW.Call(0, uintptr(i), uintptr(unsafe.Pointer(&device)), 0)
		if r == 0 {
			break
		}
		if device.StateFlags&displayDeviceActive == 0 {
			continue
		}
		if active != uint32(outputIndex) {
			active++
			continue
		}
		var mode devModeW
		mode.Size = uint16(unsafe.Sizeof(mode))
		r, _, _ = procEnumDisplaySettingsW.Call(
			uintptr(unsafe.Pointer(&device.DeviceName[0])),
			uintptr(enumCurrentSettings),
			uintptr(unsafe.Pointer(&mode)),
		)
		if r == 0 || !validSize(int(mode.PelsWidth), int(mode.PelsHeight)) {
			return 0, 0, false
		}
		return int(mode.PelsWidth), int(mode.PelsHeight), true
	}
	return 0, 0, false
}
