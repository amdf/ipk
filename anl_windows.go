package ipk

import (
	"errors"
	"sync"

	"golang.org/x/sys/windows"
)

//AnalogDevice это тип для работы с ФАС-3
type AnalogDevice struct {
	Device
	handle           windows.Handle
	idProductVariant uint16
	mutexUSB         sync.Mutex
}

//Потокобезопасный обмен данными с микроконтроллером по USB.
func (dev *AnalogDevice) deviceIoControl(direction, request byte, bytes []byte, length int) (err error) {
	if nil == dev {
		err = errors.New("deviceIoControl():" + anlErrorNoDevice)
		return
	}
	ioControlCode := IoctlEZUSBVendorOrClassRequest()
	var vcrq []byte
	var bytesReturned uint32

	switch direction {
	case VendorRequestOutput:
		vcrq = MakeVendorOrClassRequestControlStruct(0, 2, 0, 0xB0)
	case VendorRequestInput:
		vcrq = MakeVendorOrClassRequestControlStruct(1, 2, 0, 0xB0)
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

//Opened показывает открыто ли соединение с ФАС-3
func (dev *AnalogDevice) opened() bool {
	if nil == dev {
		return false
	}
	return dev.handle != windows.InvalidHandle
}

/////////////////////ИНТЕРФЕЙСНЫЕ ФУНКЦИИ/////////////////////

//Close закрыть соединение с ФАС-3
func (dev *AnalogDevice) Close() {
	if dev == nil {
		return
	}
	windows.CloseHandle(dev.handle)
	dev.handle = windows.InvalidHandle
}

//Active показывает активно ли соединение с ФАС-3
func (dev *AnalogDevice) Active() (ok bool) {
	if dev == nil {
		return
	}
	if dev != nil && dev.opened() {
		vendorID, productID := GetVendorProduct(dev.handle)
		if IDVendorElmeh == vendorID && dev.idProductVariant == productID {
			ok = true
			return
		}
	}
	return
}
