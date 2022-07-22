package ipk

import (
	"errors"
	"sync"

	"github.com/gotmc/libusb"
)

//FreqDevice это тип для работы с ФЧС-3
type FreqDevice struct {
	Device
	handle         *libusb.DeviceHandle
	ADC            DataADC
	ADCModeEnabled bool
	freqdata       dataFreq
	Teeth          uint32
	Diameter       uint32
	mutexUSB       sync.Mutex
}

func (dev *FreqDevice) deviceIoControl(direction, request byte, bytes []byte, length int) (err error) {

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

//opened показывает открыто ли соединение с ФЧС-3
func (dev *FreqDevice) opened() bool {
	if nil == dev {
		return false
	}
	return dev.handle != nil
}

/////////////////////ИНТЕРФЕЙСНЫЕ ФУНКЦИИ/////////////////////

//Close закрыть соединение с ФчС-3
func (dev *FreqDevice) Close() {
	if dev == nil {
		return
	}
	dev.handle.Close()

	dev.handle = nil
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
		ok = true
	}
	return
}
