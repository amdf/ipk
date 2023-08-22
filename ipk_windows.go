package ipk

import (
	"fmt"

	"golang.org/x/sys/windows"
)

//IPK все три устройства в одной структуре для удобства
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

const (
	devicedescriptorsize = 18 /* размер структуры DeviceDescriptor */
	ezprefix             = `\\.\ezusb-`
	ezusbioctlindex      = 0x0800
	filedeviceunknown    = uint32(0x00000022)
	methodbuffered       = 0
	methodindirect       = 1
	methodoutdirect      = 2
	methodneithre        = 3
	fileanyaccess        = 0
)

func ctlcode(devicetype, function, method, access uint32) uint32 {
	return (devicetype << 16) | (access << 14) | (function << 2) | method
}

//IoctlEZUSBGetDeviceDescriptor возвращает ioctl команду для запроса дескриптора
func IoctlEZUSBGetDeviceDescriptor() uint32 {
	return ctlcode(filedeviceunknown, ezusbioctlindex+1, methodbuffered, fileanyaccess)
}

//IoctlEZUSBVendorOrClassRequest возвращает ioctl команду для выполнения запросов к USB-устройству
func IoctlEZUSBVendorOrClassRequest() uint32 {
	return ctlcode(filedeviceunknown, ezusbioctlindex+22, methodindirect, fileanyaccess)
}

//GetVendorProduct делает запрос к EZ-USB и возвращает идентификаторы производителя и продукта
//Если запрос не удался, возвращает нули
func GetVendorProduct(handle windows.Handle) (vendorID, productID uint16) {
	if windows.InvalidHandle == handle {
		return
	}

	var bytesReturned uint32
	deviceDesc := make([]byte, devicedescriptorsize)
	iDesc := IoctlEZUSBGetDeviceDescriptor() //0x00222004

	ioerr := windows.DeviceIoControl(handle,
		iDesc, nil, 0, &deviceDesc[0], uint32(devicedescriptorsize),
		&bytesReturned, nil)
	if nil == ioerr {
		vendorID = (uint16(deviceDesc[9]) << 8) + uint16(deviceDesc[8])
		productID = (uint16(deviceDesc[11]) << 8) + uint16(deviceDesc[10])
	}

	return
}

//USBOpen соединяет приложение с устройством по USB.
//Возвращает хэндл устройства, если устройство подключено и его удалось открыть.
func USBOpen(product uint16) (handle windows.Handle, ok bool) {
	for i := 0; i < 10; i++ {
		ename := fmt.Sprintf("%s%d", ezprefix, i)
		ezusbname, _ := windows.UTF16PtrFromString(ename)
		var cferr error
		handle, cferr = windows.CreateFile(ezusbname,
			windows.GENERIC_WRITE,
			windows.FILE_SHARE_WRITE,
			nil,
			windows.OPEN_EXISTING,
			windows.FILE_ATTRIBUTE_NORMAL, 0)

		if cferr == nil {

			vendorID, productID := GetVendorProduct(handle)
			if IDVendorElmeh == vendorID && product == productID {
				ok = true
				return
			}

			windows.CloseHandle(handle)
		}
	}
	ok = false
	return
}

/*
не используем структуру, потому что неизвестно как Go упакует её в памяти
type vendorOrClassRequestControl struct {
	// transfer direction (0=host to device, 1=device to host)
	direction uint8

	// request type (1=class, 2=vendor)
	requestType uint8

	// recipient (0=device,1=interface,2=endpoint,3=other)
	recepient uint8
	//
	// see the USB Specification for an explanation of the
	// following paramaters.
	//
	requestTypeReservedBits uint8
	request                 uint8
	value                   uint16
	index                   uint16
}
*/

//MakeVendorOrClassRequestControlStruct делает структуру в виде массива байт
func MakeVendorOrClassRequestControlStruct(direction, requestType, recepient, request uint8) []byte {
	buf := make([]byte, 10)
	buf[0] = direction
	buf[1] = requestType
	buf[2] = recepient
	buf[4] = request
	return buf
}
