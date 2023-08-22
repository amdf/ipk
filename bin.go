package ipk

import (
	"encoding/binary"
	"errors"
	"time"
)

const binErrorNoConnection = `Нет соединения с ФДС-3`
const binErrorWrongParam = `Неверный параметр функции`
const binErrorNoDevice = `BinaryDevice == nil`

// Варианты кодирования сигнала ИФ
const (
	IFDisable     = 0
	IFRedYellow16 = 1
	IFYellow16    = 2
	IFGreen16     = 3
	IFRedYellow19 = 4
	IFYellow19    = 5
	IFGreen19     = 6
	IFEnable      = 7
	IFMax         = 8
)

type binaryData struct {
	data [8]byte
}

func (data *binaryData) Uint64() (val uint64) {
	if data != nil {
		val = binary.LittleEndian.Uint64(data.data[0:])
	}
	return
}

func (data *binaryData) SetUint64(val uint64) {
	if data != nil {
		binary.LittleEndian.PutUint64(data.data[0:], val)
	}
}

/////////////////////////////////////////////////////////////////////

//Set10V позволяет установить значение val
//на один из 10 В выходных сигналов на ФДС-3.
//num это номер выхода,
//num может принимать значение от 0 до 7
func (dev *BinaryDevice) Set10V(num uint, val bool) (err error) {
	if nil == dev {
		err = errors.New("Set10V():" + binErrorNoDevice)
		return
	}
	ok := false
	timeout := false
	t := time.Now()
	for !ok && !timeout {
		var bindata binaryData
		bindata, err = dev.getDataUSB()
		if nil == err {
			if val { //сброс бита когда true, потому что так было в SRS_BIN2_Set (SrsBin2.cpp, srs2.dll)
				bindata.data[0] &^= 1 << num
			} else {
				bindata.data[0] |= 1 << num
			}
			err = dev.setDataUSB(bindata)
		}
		ok = (nil == err)
		timeout = time.Since(t) >= maxDelayUSB
	}
	if ok && timeout {
		err = errors.New("Set10V():" + errUsbTimeout)
	}
	return
}

//UintSet10V позволяет установить значение val
//сразу на все 10 В выходные сигналы ФДС-3.
//Младший бит val соответствует первому 10 В выходу ФДС-3
func (dev *BinaryDevice) UintSet10V(val uint8) (err error) {
	if nil == dev {
		err = errors.New("UintSet10V():" + binErrorNoDevice)
		return
	}
	ok := false
	timeout := false
	t := time.Now()
	for !ok && !timeout {
		var bindata binaryData
		bindata, err = dev.getDataUSB()
		if nil == err {
			//инверсия потому что так было в SRS_BIN2_Set (SrsBin2.cpp, srs2.dll)
			bindata.data[0] = ^val
			err = dev.setDataUSB(bindata)
		}
		ok = (nil == err)
		timeout = time.Since(t) >= maxDelayUSB
	}
	if ok && timeout {
		err = errors.New("UintSet10V():" + errUsbTimeout)
	}
	return
}

//UintGetOutput10V возвращает все значения, установленные в данный момент
//на выходах 10 В в виде одного числа.
//Младший бит val соответствует первому 10 В выходу ФДС-3
func (dev *BinaryDevice) UintGetOutput10V() (val uint8, err error) {
	if nil == dev {
		err = errors.New("UintGetOutput10V():" + binErrorNoDevice)
		return
	}
	ok := false
	timeout := false
	t := time.Now()
	for !ok && !timeout {
		var bindata binaryData
		bindata, err = dev.getDataUSB()
		if nil == err {
			val = bindata.data[0]
			//инверсия потому что так было в SRS_BIN2_Set (SrsBin2.cpp, srs2.dll)
			val = ^val
		}
		ok = (nil == err)
		timeout = time.Since(t) >= maxDelayUSB
	}
	if ok && timeout {
		err = errors.New("UintGetOutput10V():" + errUsbTimeout)
	}

	return
}

//UintSet50V позволяет установить значение val
//сразу на все 50 В выходные сигналы ФДС-3.
//Младший бит val соответствует первому 50 В выходу ФДС-3.
//Побочный эффект: вызов этой функции установит 28-й выход,
//который отвечает за сигнал ИФ. Но через 10 мс его состояние
//снова поменяется микроконтроллером на то, которое предусмотрено.
//Так что теоретически вызов этой функции может затронуть кодирование ИФ(?).
func (dev *BinaryDevice) UintSet50V(val uint64) (err error) {
	if nil == dev {
		err = errors.New("UintSet50V():" + binErrorNoDevice)
		return
	}
	ok := false
	timeout := false
	t := time.Now()
	for !ok && !timeout {
		var bindata binaryData
		bindata, err = dev.getDataUSB()
		if nil == err {
			save10v := uint64(bindata.data[0])     // сохраняем 10 В сигналы (не хотим их менять)
			setval := ^val                         //инверсия потому что так было в SRS_BIN2_Set (SrsBin2.cpp, srs2.dll)
			bindata.SetUint64(setval<<8 | save10v) // новые значения 50 В + старые значения 10 В
			err = dev.setDataUSB(bindata)
		}
		ok = (nil == err)
		timeout = time.Since(t) >= maxDelayUSB
	}
	if ok && timeout {
		err = errors.New("UintSet50V():" + errUsbTimeout)
	}

	return
}

//UintGetOutput50V возвращает все значения, установленные в данный момент
//на выходах 50 В в виде одного числа.
func (dev *BinaryDevice) UintGetOutput50V() (val uint64, err error) {
	if nil == dev {
		err = errors.New("UintGetOutput50V():" + binErrorNoDevice)
		return
	}
	ok := false
	timeout := false
	t := time.Now()
	for !ok && !timeout {
		var bindata binaryData
		bindata, err = dev.getDataUSB()
		if err == nil {
			val = bindata.Uint64() >> 8
			val = ^val //инверсия потому что так было в SRS_BIN2_Set (SrsBin2.cpp, srs2.dll)
		}
		ok = (nil == err)
		timeout = time.Since(t) >= maxDelayUSB
	}
	if ok && timeout {
		err = errors.New("UintGetOutput50V():" + errUsbTimeout)
	}
	return
}

//Set50V позволяет установить значение val
//на один из 50 В выходных сигналов на ФДС-3.
//num это номер выхода, нумерация выходов начинается с нуля.
//num может принимать значение от 0 до 35
//(кроме 28, вместо этого выхода - сигнал ИФ, который обрабатывается отдельно).
func (dev *BinaryDevice) Set50V(num uint, val bool) (err error) {
	if nil == dev {
		err = errors.New("Set50V():" + binErrorNoDevice)
		return
	}
	if num == 28 { // не позволяем менять сигнал ИФ
		return
	}

	ok := false
	timeout := false
	t := time.Now()
	for !ok && !timeout {
		var bindata binaryData
		bindata, err = dev.getDataUSB()
		if nil == err {
			inum := num + 8 // нумерация 50 В сигналов начианется с 8 бита
			ibs := bindata.Uint64()
			if val { //сброс бита когда true, потому что так было в SRS_BIN2_Set (SrsBin2.cpp, srs2.dll)
				ibs &^= uint64(1) << inum
			} else {
				ibs |= uint64(1) << inum
			}
			bindata.SetUint64(ibs)
			err = dev.setDataUSB(bindata)
		}
		ok = (nil == err)
		timeout = time.Since(t) >= maxDelayUSB
	}

	if ok && timeout {
		err = errors.New("Set50V():" + errUsbTimeout)
	}

	return
}

//GetOutputIF возвращает установленный в данный момент сигнал ИФ
func (dev *BinaryDevice) GetOutputIF() (state uint8, err error) {
	if nil == dev {
		err = errors.New("GetOutputIF():" + binErrorNoDevice)
		return
	}

	if !dev.opened() {
		err = errors.New("GetOutputIF():" + binErrorNoConnection)
		return
	}

	getstate := make([]byte, 1)

	err = dev.deviceIoControl(VendorRequestInput, 0xB1, getstate, 1)

	if nil != err {
		err = errors.New("GetOutputIF():" + err.Error())
		return
	}

	state = getstate[0]

	return
}

//SetIF устанавливает сигнал ИФ
func (dev *BinaryDevice) SetIF(state uint8) (err error) {
	if nil == dev {
		err = errors.New("SetIF():" + binErrorNoDevice)
		return
	}
	if state >= IFMax {
		err = errors.New("SetIF():" + binErrorWrongParam)
		return
	}

	if !dev.opened() {
		err = errors.New(binErrorNoConnection)
		return
	}

	err = dev.deviceIoControl(VendorRequestOutput, 0xB1, []byte{state}, 1)

	if nil != err {
		err = errors.New("SetIF():" + err.Error())
	}

	return
}

//GetOutputTURT возвращает состояние сигнала TURT
func (dev *BinaryDevice) GetOutputTURT() (val bool, err error) {
	if nil == dev {
		err = errors.New("GetOutputTURT():" + binErrorNoDevice)
		return
	}
	if !dev.opened() {
		err = errors.New("GetOutputTURT():" + binErrorNoConnection)
		return
	}

	state := make([]byte, 1)

	err = dev.deviceIoControl(VendorRequestInput, 0xB2, state, 1)

	if nil != err {
		err = errors.New("GetOutputTURT():" + err.Error())
		return
	}

	val = (state[0] != 0)

	return
}

//SetTURT устанавливает сигнал ИФ
func (dev *BinaryDevice) SetTURT(val bool) (err error) {
	if nil == dev {
		err = errors.New("SetTURT():" + binErrorNoDevice)
		return
	}
	if !dev.opened() {
		err = errors.New("SetTURT():" + binErrorNoConnection)
		return
	}

	var state uint8

	if val {
		state = 1
	}

	err = dev.deviceIoControl(VendorRequestOutput, 0xB2, []byte{state}, 1)

	if nil != err {
		err = errors.New("SetTURT():" + err.Error())
	}

	return
}

//////////////////////////////////////////////////////////////

func (dev *BinaryDevice) getDataUSB() (bindata binaryData, err error) {
	if nil == dev {
		err = errors.New("BinaryDevice.getDataUSB():" + binErrorNoDevice)
		return
	}

	if !dev.opened() {
		err = errors.New("BinaryDevice.getDataUSB():" + binErrorNoConnection)
		return
	}

	binaryDataSize := len(bindata.data)
	data := make([]byte, binaryDataSize)

	err = dev.deviceIoControl(VendorRequestInput, 0xB0, data, len(data))

	if nil != err {
		err = errors.New("BinaryDevice.getDataUSB():" + err.Error())
		return
	}

	for i := 0; i < binaryDataSize; i++ {
		bindata.data[i] = data[i]
	}

	return
}

func (dev *BinaryDevice) setDataUSB(bindata binaryData) (err error) {
	if nil == dev {
		err = errors.New("BinaryDevice.setDataUSB():" + binErrorNoDevice)
		return
	}

	if !dev.opened() {
		err = errors.New("BinaryDevice.setDataUSB():" + binErrorNoConnection)
		return
	}

	err = dev.deviceIoControl(VendorRequestOutput, 0xB0, bindata.data[:], len(bindata.data))

	if nil != err {
		err = errors.New("BinaryDevice.setDataUSB():" + err.Error())
	}
	return
}

//Open соединиться с ФДС-3
func (dev *BinaryDevice) Open() (ok bool) {
	if dev == nil {
		return
	}
	dev.handle, ok = USBOpen(IDProductBIN)
	return
}

//TODO: контролировать время обращения по USB для функций TURT и IF?
