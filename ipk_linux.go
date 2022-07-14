package ipk

import (
	"time"

	"github.com/gotmc/libusb"
)

//Device - интерфейс устройств, составных частей ФПС-3
type Device interface {
	Open(ok bool)
	Close()
	Active() bool
}

//IPK все три устройства в одной структуре для удобства
type IPK struct {
	//TODO: вернуть
	// AnalogDev *AnalogDevice
	// BinDev    *BinaryDevice
	// FreqDev   *FreqDevice
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

//UBS-идентификаторы оборудования ИПК-3
const (
	IDVendorElmeh     = uint16(0x0547)
	IDProductANL12bit = 0x0892 // ФАС-3 12 бит ЦАП
	IDProductANL16bit = 0x0894 // ФАС-3 16 бит ЦАП
	IDProductBIN      = 0x0891 // ФДС-3
	IDProductFRQ      = 0x0893 // ФЧС-3
)

//Максимально допустимое время реакции во время обращения к оборудованию по USB.
const maxDelayUSB = 100 * time.Millisecond
const errUsbTimeout = `Слишком большое время отклика по USB`

const usbDirectionIn = byte(0xC0)
const usbDirectionOut = byte(0x40)

//USBOpen соединяет приложение с устройством по USB.
//Возвращает хэндл устройства, если устройство подключено и его удалось открыть.
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
