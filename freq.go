package ipk

import (
	"bytes"
	"encoding/binary"
	"errors"
)

//направление движения
const (
	MotionUnknown   = 0xFF // неизвестно (ошибка)
	MotionOnward    = 1    // вперёд
	MotionBackwards = 0    // назад
)

const frqErrorNoConnection = `Нет соединения с ФАС-3`
const frqErrorWrongParam = `Неверный параметр функции`
const frqErrorNoDevice = `FreqDevice == nil`

const dataADCsize = 14

//DataADC Значения АЦП с ФЧС-3
type DataADC struct {
	Dat1         uint32
	Dat2         uint32
	ReferenceVal uint32
	DivisorVal   uint16
}

func (data *DataADC) setFromBytes(inbuf []byte) bool {
	// должно быть достаточное количество байт чтобы заполнить структуру
	if nil == inbuf || nil == data || len(inbuf) < dataADCsize {
		return false
	}

	//конвертация значений, пришедщих из микроконтроллера
	data.Dat1 = binary.BigEndian.Uint32(inbuf[0:])
	data.Dat2 = binary.BigEndian.Uint32(inbuf[4:])
	data.ReferenceVal = binary.BigEndian.Uint32(inbuf[8:])
	data.DivisorVal = binary.BigEndian.Uint16(inbuf[12:])

	return true
}

const dataFreqSize = 34

//dataFreq Данные для задания частоты на ФЧС-3
type dataFreq struct {
	cmd uint8 // команда изменения

	freq1      uint32 // новое значение частоты для 1 генератора
	freq1delta int32  // приращение частоты в секунду для 1 генератора
	way1count  uint32 // счёчик пути для 1 генератора
	limitWay1  uint32 // путь перемещения 1 генератора

	freq2      uint32 // новое значение частоты для 2 генератора
	freq2delta int32  // приращение частоты в секунду для 2 генератора
	way2count  uint32 // счёчик пути для 2 генератора
	limitWay2  uint32 // путь перемещения 2 генератора

	motion uint8 // направление движения (1 - вперёд; 0 - назад;)
}

// преобразует в массив big endian для отправки на микроконтроллер
func (data *dataFreq) toBytes() []byte {
	if nil == data {
		return nil
	}
	buf := new(bytes.Buffer)

	binary.Write(buf, binary.BigEndian, data.cmd)
	binary.Write(buf, binary.BigEndian, data.freq1)
	binary.Write(buf, binary.BigEndian, data.freq1delta)
	binary.Write(buf, binary.BigEndian, data.way1count)
	binary.Write(buf, binary.BigEndian, data.limitWay1)
	binary.Write(buf, binary.BigEndian, data.freq2)
	binary.Write(buf, binary.BigEndian, data.freq2delta)
	binary.Write(buf, binary.BigEndian, data.way2count)
	binary.Write(buf, binary.BigEndian, data.limitWay2)
	binary.Write(buf, binary.BigEndian, data.motion)

	return buf.Bytes()
}

func (data *dataFreq) setFromBytes(inbuf []byte) bool {
	// должно быть достаточное количество байт чтобы заполнить структуру
	if nil == inbuf || nil == data || len(inbuf) < dataFreqSize {
		return false
	}

	//конвертация значений, пришедщих из микроконтроллера
	data.cmd = inbuf[0]
	data.freq1 = binary.BigEndian.Uint32(inbuf[1:])
	data.freq1delta = int32(binary.BigEndian.Uint32(inbuf[4+1:]))
	data.way1count = binary.BigEndian.Uint32(inbuf[8+1:])
	data.limitWay1 = binary.BigEndian.Uint32(inbuf[12+1:])
	data.freq2 = binary.BigEndian.Uint32(inbuf[16+1:])
	data.freq2delta = int32(binary.BigEndian.Uint32(inbuf[20+1:]))
	data.way2count = binary.BigEndian.Uint32(inbuf[24+1:])
	data.limitWay2 = binary.BigEndian.Uint32(inbuf[28+1:])
	data.motion = inbuf[33]

	return true
}

//UpdateFreqDataUSB получает значения, связанные с частотой, по USB из ФЧС-3.
//Нужно регулярно вызывать эту функцию, чтобы значения обновлялись.
func (dev *FreqDevice) UpdateFreqDataUSB() (err error) {

	if nil == dev {
		err = errors.New("FreqDevice.getFreqDataUSB():" + frqErrorNoDevice)
		return
	}

	if !dev.opened() {
		err = errors.New("FreqDevice.getFreqDataUSB():" + frqErrorNoConnection)
		return
	}

	freqbytes := make([]byte, dataFreqSize)

	err = dev.deviceIoControl(VendorRequestInput, 0xB0, freqbytes, len(freqbytes))

	if nil != err {
		err = errors.New("FreqDevice.getFreqDataUSB():" + err.Error())
		return
	}

	dev.freqdata.setFromBytes(freqbytes)
	dev.ADCModeEnabled, err = dev.isADCEnabled()

	return
}

//установить путь перемещения (в импульсах)
func (dev *FreqDevice) setLimitWayUSB(wayImpulseCount uint32) (err error) {
	if nil == dev {
		err = errors.New("FreqDevice.setLimitWayUSB():" + frqErrorNoDevice)
		return
	}

	if !dev.opened() {
		err = errors.New("FreqDevice.setLimitWayUSB():" + frqErrorNoConnection)
		return
	}

	var dataout dataFreq
	dataout.cmd = 6
	dataout.limitWay1 = wayImpulseCount
	dataout.limitWay2 = wayImpulseCount

	freqbytes := dataout.toBytes()

	err = dev.deviceIoControl(VendorRequestOutput, 0xB0, freqbytes, len(freqbytes))

	if nil != err {
		err = errors.New("FreqDevice.setCrsWayUSB():" + err.Error())
	}

	return
}

//установка нового значения обоих генераторов частоты
func (dev *FreqDevice) setFreqUSB(freq1, freq2 uint32) (err error) {
	if nil == dev {
		err = errors.New("FreqDevice.setFreqUSB():" + frqErrorNoDevice)
		return
	}

	if !dev.opened() {
		err = errors.New("FreqDevice.setFreqUSB():" + frqErrorNoConnection)
		return
	}

	var dataout dataFreq
	dataout.cmd = 2
	dataout.freq1 = freq1
	dataout.freq2 = freq2

	freqbytes := dataout.toBytes()

	err = dev.deviceIoControl(VendorRequestOutput, 0xB0, freqbytes, len(freqbytes))

	if nil != err {
		err = errors.New("FreqDevice.setFreqUSB():" + err.Error())
	}

	return
}

//установка нового значения обоих генераторов частоты
func (dev *FreqDevice) setDeltaUSB(delta1, delta2 int32) (err error) {
	if nil == dev {
		err = errors.New("FreqDevice.setDeltaUSB():" + frqErrorNoDevice)
		return
	}

	if !dev.opened() {
		err = errors.New("FreqDevice.setDeltaUSB():" + frqErrorNoConnection)
		return
	}

	var dataout dataFreq
	dataout.cmd = 1
	dataout.freq1delta = delta1
	dataout.freq2delta = delta2

	freqbytes := dataout.toBytes()

	err = dev.deviceIoControl(VendorRequestOutput, 0xB0, freqbytes, len(freqbytes))

	if nil != err {
		err = errors.New("FreqDevice.setDeltaUSB():" + err.Error())
	}

	return
}

func (dev *FreqDevice) setWayCountUSB(way1, way2 uint32) (err error) {
	if nil == dev {
		err = errors.New("FreqDevice.setDeltaUSB():" + frqErrorNoDevice)
		return
	}

	if !dev.opened() {
		err = errors.New("FreqDevice.setDeltaUSB():" + frqErrorNoConnection)
		return
	}

	var dataout dataFreq
	dataout.cmd = 3
	dataout.way1count = way1
	dataout.way2count = way2

	freqbytes := dataout.toBytes()

	err = dev.deviceIoControl(VendorRequestOutput, 0xB0, freqbytes, len(freqbytes))

	if nil != err {
		err = errors.New("FreqDevice.setDeltaUSB():" + err.Error())
	}

	return
}

func (dev *FreqDevice) setMotionUSB(direction uint8) (err error) {
	if nil == dev {
		err = errors.New("FreqDevice.setDeltaUSB():" + frqErrorNoDevice)
		return
	}

	if !dev.opened() {
		err = errors.New("FreqDevice.setDeltaUSB():" + frqErrorNoConnection)
		return
	}

	var dataout dataFreq
	dataout.cmd = 5
	switch direction {
	case MotionOnward:
		dataout.motion = 0 //было 1; решили поменять в мае 2022
	case MotionBackwards:
		dataout.motion = 1 //было 0; решили поменять в мае 2022
	default:
		err = errors.New("FreqDevice.setDeltaUSB():" + frqErrorWrongParam)
		return
	}

	freqbytes := dataout.toBytes()

	err = dev.deviceIoControl(VendorRequestOutput, 0xB0, freqbytes, len(freqbytes))

	if nil != err {
		err = errors.New("FreqDevice.setDeltaUSB():" + err.Error())
	}

	return
}
