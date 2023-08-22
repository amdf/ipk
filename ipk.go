package ipk

import "time"

//Device - интерфейс устройств, составных частей ФПС-3
type Device interface {
	Open(ok bool)
	Close()
	Active() bool
}

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

const VendorRequestInput = 0xC0
const VendorRequestOutput = 0x40
