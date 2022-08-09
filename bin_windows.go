package ipk

import (
	"errors"
	"sync"

	"golang.org/x/sys/windows"
)

//BinaryDevice это тип для работы с ФДС-3
type BinaryDevice struct {
	Device
	handle   windows.Handle
	mutexUSB sync.Mutex
}

//Потокобезопасный обмен данными с микроконтроллером по USB.
func (dev *BinaryDevice) deviceIoControl(direction, request byte, bytes []byte, length int) (err error) {
	if nil == dev {
		err = errors.New("deviceIoControl():" + binErrorNoDevice)
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

//opened показывает открыто ли соединение с ФДС-3
func (dev *BinaryDevice) opened() bool {
	if nil == dev {
		return false
	}
	return dev.handle != windows.InvalidHandle
}

/////////////////////ИНТЕРФЕЙСНЫЕ ФУНКЦИИ/////////////////////

//Close закрыть соединение с ФДС-3
func (dev *BinaryDevice) Close() {
	if dev == nil {
		return
	}
	windows.CloseHandle(dev.handle)
	dev.handle = windows.InvalidHandle
}

//Active показывает активно ли соединение с ФДС-3
func (dev *BinaryDevice) Active() (ok bool) {
	if dev == nil {
		return
	}
	if dev.opened() {
		vendorID, productID := GetVendorProduct(dev.handle)
		if IDVendorElmeh == vendorID && IDProductBIN == productID {
			ok = true
			return
		}
	}
	return
}
