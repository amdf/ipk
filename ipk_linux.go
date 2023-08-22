package ipk

import (
	"github.com/gotmc/libusb"
)

// IPK все три устройства в одной структуре для удобства
type IPK struct {
	AnalogDev *AnalogDevice
	BinDev    *BinaryDevice
	FreqDev   *FreqDevice
}

/*
не используем структуру, потому что неизвестно как Go упакует её в памяти
type DeviceDescriptor struct {
	Length              uint8
	DescriptorType      byte
	USBSpecification    uint16
	DeviceClass         byte
	DeviceSubClass      byte
	DeviceProtocol      byte
	MaxPacketSize0      uint8
	VendorID            uint16
	ProductID           uint16
	DeviceReleaseNumber uint16
	ManufacturerIndex   uint8
	ProductIndex        uint8
	SerialNumberIndex   uint8
	NumConfigurations   uint8
}*/

// USBOpen соединяет приложение с устройством по USB.
// Возвращает хэндл устройства, если устройство подключено и его удалось открыть.
func USBOpen(product uint16) (handle *libusb.DeviceHandle, ok bool) {
	usb, err1 := libusb.NewContext()
	if err1 != nil {
		return
	}

	_, h, err2 := usb.OpenDeviceWithVendorProduct(IDVendorElmeh, product)
	if err2 != nil {
		return
	}

	handle = h
	ok = true

	return
}
