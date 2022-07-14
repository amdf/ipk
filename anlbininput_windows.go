package ipk

import "errors"

//UintGetBinaryInput позволяет получить данные с двоичных входов ФАС-3 в виде одного числа.
func (dev *AnalogDevice) UintGetBinaryInput() (val uint16, err error) {
	if nil == dev {
		err = errors.New("UintGetBinaryInput():" + anlErrorWrongParam)
		return
	}
	var as analogDeviceData
	err = dev.getDataUSB(&as)
	if nil == err {
		val = as.binary[0]
		return
	}
	return
}

//GetBinaryInputVal позволяет получить значение отдельного двоичного входа ФАС-3.
//num - номер двоичного входа, от 0 до 15.
func (dev *AnalogDevice) GetBinaryInputVal(num uint16) (val bool, err error) {
	if nil == dev || num >= 16 {
		err = errors.New("GetBinaryInputVal():" + anlErrorWrongParam)
		return
	}
	val = false
	var uintval uint16
	uintval, err = dev.UintGetBinaryInput()
	if nil == err {
		if 0 != (uintval & (1 << num)) {
			val = true
			return
		}
	}
	return
}
