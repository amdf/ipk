package ipk

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
)

const (
	ipkUpdatePrepare = iota
	ipkUpdateWrite
	ipkUpdateFinish
	ipkUpdateReboot
)

//UpdateIPK данные для обновления плат ФПС-3 на основе STM32
type UpdateIPK struct {
	code          uint8
	uFlashAddress uint32
	uWord         uint32
}

// преобразует в массив big endian для отправки на микроконтроллер
func (data *UpdateIPK) toBytes() []byte {
	if nil == data {
		return nil
	}
	buf := new(bytes.Buffer)

	binary.Write(buf, binary.BigEndian, data.code)
	binary.Write(buf, binary.BigEndian, data.uFlashAddress)
	binary.Write(buf, binary.BigEndian, data.uWord)

	return buf.Bytes()
}

/* В старой ревизии платы через эту структуру осуществлялась отладка АЦП.
В новой ревизии (STM32) через эту структуру возвращается версия прошивки.
Это сделано затем, чтобы задействовать обработчик запроса в старой ревизии,
иначе старая плата не смогла бы ответить на неизвестный ей запрос версии.
*/
type debugADC struct {
	rawData     [2]uint32
	convData    [2]uint32
	numInterval [2]uint16
}

const debugADCsize = 20

func (data *debugADC) setFromBytes(inbuf []byte) bool {
	// должно быть достаточное количество байт чтобы заполнить структуру
	if nil == inbuf || nil == data || len(inbuf) < 20 {
		return false
	}

	data.rawData[0] = binary.BigEndian.Uint32(inbuf[0:])
	data.rawData[1] = binary.BigEndian.Uint32(inbuf[4:])
	data.convData[0] = binary.BigEndian.Uint32(inbuf[8:])
	data.convData[1] = binary.BigEndian.Uint32(inbuf[12:])
	data.numInterval[0] = binary.BigEndian.Uint16(inbuf[16:])
	data.numInterval[1] = binary.BigEndian.Uint16(inbuf[18:])

	return true
}

func (dev *FreqDevice) sendUpdateCommand(bytes []byte) (err error) {
	if nil == dev {
		err = errors.New("FreqDevice.sendUpdateCommand():" + frqErrorNoDevice)
		return
	}

	if !dev.opened() {
		err = errors.New("FreqDevice.sendUpdateCommand():" + frqErrorNoConnection)
		return
	}

	if 0 == len(bytes) {
		err = errors.New("FreqDevice.sendUpdateCommand():" + frqErrorWrongParam)
		return
	}

	err = dev.deviceIoControl(VendorRequestOutput, 0xB4, bytes, len(bytes))

	if nil != err {
		err = errors.New("FreqDevice.sendUpdateCommand():" + err.Error())
	}

	return
}

//PrepareUpdate подготавливает запись обновления прошивки
func (dev *FreqDevice) PrepareUpdate() (err error) {
	if nil == dev {
		err = errors.New("FreqDevice.PrepareUpdate():" + frqErrorNoDevice)
		return
	}

	dataout := UpdateIPK{code: ipkUpdatePrepare}

	err = dev.sendUpdateCommand(dataout.toBytes())

	return
}

//WriteUpdate записывает слово по адресу (пишет прошивку в память)
func (dev *FreqDevice) WriteUpdate(uFlashAddress uint32, uWord uint32) (err error) {
	if nil == dev {
		err = errors.New("FreqDevice.WriteUpdate():" + frqErrorNoDevice)
		return
	}

	dataout := UpdateIPK{code: ipkUpdateWrite}
	dataout.uFlashAddress = uFlashAddress
	dataout.uWord = uWord

	err = dev.sendUpdateCommand(dataout.toBytes())

	return
}

//FinishUpdate завершает запись обновления прошивки
func (dev *FreqDevice) FinishUpdate() (err error) {
	if nil == dev {
		err = errors.New("FreqDevice.FinishUpdate():" + frqErrorNoDevice)
		return
	}

	dataout := UpdateIPK{code: ipkUpdateFinish}

	err = dev.sendUpdateCommand(dataout.toBytes())

	return
}

//RestartToAnotherBank отправляет команду устройству перезагрузиться с другого банка памяти
func (dev *FreqDevice) RestartToAnotherBank() (err error) {
	if nil == dev {
		err = errors.New("FreqDevice.RestartToAnotherBank():" + frqErrorNoDevice)
		return
	}

	dataout := UpdateIPK{code: ipkUpdateReboot}

	err = dev.sendUpdateCommand(dataout.toBytes())

	return
}

//GetVersionString возвращает версию прошивки ФЧС-3 (например, 1.0.0) в виде строки
//Старая версия платы - 0.0.0
func (dev *FreqDevice) GetVersionString() (version string, err error) {
	if nil == dev {
		err = errors.New("FreqDevice.GetVersion():" + frqErrorNoDevice)
		return
	}
	major, minor, patch, err2 := dev.GetVersion()
	err = err2

	version = fmt.Sprintf("%d.%d.%d", major, minor, patch)
	return
}

//GetVersion возвращает версию прошивки ФЧС-3 (например, 1.0.0)
//Старая версия платы - 0.0.0
func (dev *FreqDevice) GetVersion() (major, minor, patch uint32, err error) {
	if nil == dev {
		err = errors.New("FreqDevice.GetVersion():" + frqErrorNoDevice)
		return
	}

	if !dev.opened() {
		err = errors.New("FreqDevice.GetVersion():" + frqErrorNoConnection)
		return
	}

	bdat := make([]byte, debugADCsize)

	err = dev.deviceIoControl(VendorRequestInput, 0xB4, bdat, len(bdat))

	if nil != err {
		err = errors.New("FreqDevice.GetVersion():" + err.Error())
		return
	}

	//определяем номер версии прошивки

	var data debugADC

	if data.setFromBytes(bdat) {
		const verSignature = 0xDEADC0DE
		if verSignature == data.rawData[0] {
			major = data.rawData[1]
			minor = data.convData[0]
			patch = data.convData[1]
		}
	} else {
		err = errors.New("FreqDevice.GetVersion():" + "wrong data from device")
		return
	}

	return
}
