package ipk

import (
	"errors"
	"sync"

	"golang.org/x/sys/windows"
)

//FreqDevice это тип для работы с ФЧС-3
type FreqDevice struct {
	Device
	handle         windows.Handle
	ADC            DataADC
	ADCModeEnabled bool
	freqdata       dataFreq
	Teeth          uint32
	Diameter       uint32
	mutexUSB       sync.Mutex
}

func (dev *FreqDevice) deviceIoControl(direction, request byte, bytes []byte, length int) (err error) {
	if nil == dev {
		err = errors.New("deviceIoControl():" + frqErrorNoDevice)
		return
	}
	ioControlCode := IoctlEZUSBVendorOrClassRequest()
	var vcrq []byte
	var bytesReturned uint32

	switch direction {
	case VendorRequestOutput:
		vcrq = MakeVendorOrClassRequestControlStruct(0, 2, 0, request)
	case VendorRequestInput:
		vcrq = MakeVendorOrClassRequestControlStruct(1, 2, 0, request)
	default:
		err = errors.New("unknown deviceIoControl transfer")
	}
	if err == nil {
		dev.mutexUSB.Lock()
		err = windows.DeviceIoControl(dev.handle, ioControlCode, &vcrq[0], uint32(len(vcrq)), &bytes[0], uint32(length), &bytesReturned, nil)
		dev.mutexUSB.Unlock()
	}
	return
}

//opened показывает открыто ли соединение с ФЧС-3
func (dev *FreqDevice) opened() bool {
	if nil == dev {
		return false
	}
	return dev.handle != windows.InvalidHandle
}

/////////////////////ИНТЕРФЕЙСНЫЕ ФУНКЦИИ/////////////////////

//Close закрыть соединение с ФчС-3
func (dev *FreqDevice) Close() {
	if dev == nil {
		return
	}
	windows.CloseHandle(dev.handle)
	dev.handle = windows.InvalidHandle
}

//Open соединиться с ФЧС-3
func (dev *FreqDevice) Open() (ok bool) {
	dev.handle, ok = USBOpen(IDProductFRQ)
	return
}

//Active показывает активно ли соединение с ФЧС-3
func (dev *FreqDevice) Active() (ok bool) {
	if dev == nil {
		return
	}
	if dev.opened() {
		vendorID, productID := GetVendorProduct(dev.handle)
		if IDVendorElmeh == vendorID && IDProductFRQ == productID {
			ok = true
			return
		}
	}
	return
}
