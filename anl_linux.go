package ipk

import (
	"errors"
	"sync"

	"github.com/gotmc/libusb"
)

//AnalogDevice это тип для работы с ФАС-3
type AnalogDevice struct {
	Device
	handle           *libusb.DeviceHandle
	idProductVariant uint16
	mutexUSB         sync.Mutex
}

//Потокобезопасный обмен данными с микроконтроллером по USB.
func (dev *AnalogDevice) deviceIoControl(direction, request byte, bytes []byte, length int) (err error) {
	if nil == dev {
		err = errors.New("deviceIoControl():" + anlErrorNoDevice)
		return
	}
	dev.mutexUSB.Lock()
	switch direction {
	case VendorRequestOutput:
		_, err = dev.handle.ControlTransfer(VendorRequestOutput, request, 0, 0, bytes, length, int(maxDelayUSB.Milliseconds()))
	case VendorRequestInput:
		_, err = dev.handle.ControlTransfer(VendorRequestInput, request, 0, 0, bytes, length, int(maxDelayUSB.Milliseconds()))
	default:
		err = errors.New("unknown deviceIoControl transfer")
	}
	dev.mutexUSB.Unlock()
	return
}

//opened показывает открыто ли соединение с ФАС-3
func (dev *AnalogDevice) opened() bool {
	if nil == dev {
		return false
	}
	return dev.handle != nil
}

/////////////////////ИНТЕРФЕЙСНЫЕ ФУНКЦИИ/////////////////////

//Close закрыть соединение с ФАС-3
func (dev *AnalogDevice) Close() {
	if dev == nil {
		return
	}
	dev.handle.Close()

	dev.handle = nil
}

//Active показывает активно ли соединение с ФАС-3
func (dev *AnalogDevice) Active() (ok bool) {
	if dev == nil {
		return
	}
	if dev.opened() {
		ok = true
	}
	return
}
