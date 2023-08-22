package ipk

import (
	"bytes"
	"encoding/binary"
	"errors"
	"math"
	"time"
)

const anlErrorNoConnection = `Нет соединения с ФАС-3`
const anlErrorWrongParam = `Неверный параметр функции`
const anlErrorNoDevice = `FreqDevice == nil`
const analogCount = 14

// 14 ЦАПов
const (
	DAC1 = iota
	DAC2
	DAC3
	DAC4
	DAC5
	DAC6
	DAC7
	DAC8
	DAC9
	DAC10
	DAC11
	DAC12
	DAC13
	DAC14
)

type analogDeviceData struct {
	analog [analogCount]uint16
	freq   [freqCount]uint16
	binary [1]uint16
}

func (data *analogDeviceData) Size() int {
	return 2 * (len(data.analog) + len(data.freq) + len(data.binary))
}

// преобразует в массив big endian для отправки на микроконтроллер
func (data *analogDeviceData) toBytes() []byte {
	if nil == data {
		return nil
	}
	buf := new(bytes.Buffer)

	for _, v := range data.analog {
		binary.Write(buf, binary.BigEndian, v)
	}
	for _, v := range data.freq {
		binary.Write(buf, binary.BigEndian, v)
	}
	for _, v := range data.binary {
		binary.Write(buf, binary.BigEndian, v)
	}
	return buf.Bytes()
}

func (data *analogDeviceData) setFromBytes(inbuf []byte) bool {
	// должно быть достаточное количество байт чтобы заполнить структуру
	if nil == inbuf || nil == data || len(inbuf) < data.Size() {
		return false
	}
	//конвертация значений, пришедщих из микроконтроллера
	inbufIndex := 0
	for i := range data.analog {
		data.analog[i] = binary.BigEndian.Uint16(inbuf[inbufIndex:])
		inbufIndex += 2
	}
	for i := range data.freq {
		data.freq[i] = binary.BigEndian.Uint16(inbuf[inbufIndex:])
		inbufIndex += 2
	}
	for i := range data.binary {
		data.binary[i] = binary.BigEndian.Uint16(inbuf[inbufIndex:])
		inbufIndex += 2
	}

	return true
}

///////////////////////////////////////////////////////////////

// GetProductID С помощью этой функции можно узнать, какой вариант ФАС
// используется, старый (12 бит) или новый (16 бит)
func (dev *AnalogDevice) GetProductID() uint16 {
	if nil != dev {
		return dev.idProductVariant
	}
	return 0
}

///////////////////////////////////////////////////////////////

// setDAC задаёт значение на один из каналов ЦАП ФАС-3.
// Параметр ch - номер канала. Значение от ipk.DAC1 до ipk.DAC14.
// Параметр val - значение для вывода на ЦАП. Значение следует получить
// с помощью одной из функций: MilliAmperToDAC, AtToDAC, KiloPascalToDAC
func (dev *AnalogDevice) setDAC(ch uint8, val uint16) (err error) {
	if nil == dev {
		err = errors.New("setDAC():" + anlErrorNoDevice)
		return
	}
	if ch >= analogCount {
		err = errors.New("setDAC():" + anlErrorWrongParam)
		return
	}

	// делаем несколько попыток, так как обращение по USB иногда приводит к ошибке device not functioning

	ok := false
	timeout := false
	t := time.Now()
	for !ok && !timeout {
		var as analogDeviceData
		err = dev.getDataUSB(&as)
		if nil == err {
			as.analog[ch] = val
			err = dev.setDataUSB(&as)
		}
		ok = (nil == err)
		timeout = time.Since(t) >= maxDelayUSB
	}

	if ok && timeout {
		err = errors.New("setDAC():" + errUsbTimeout)
	}

	return
}

// getOutputDAC позволяет узнать, какое значение установлено в данный момент
// на одном из каналов ЦАП ФАС-3.
// Параметр ch - номер канала. Значение от ipk.DAC1 до ipk.DAC14.
func (dev *AnalogDevice) getOutputDAC(ch uint8) (val uint16, err error) {
	if nil == dev {
		err = errors.New("getOutputDAC():" + anlErrorNoDevice)
		return
	}
	if ch >= analogCount {
		err = errors.New("getOutputDAC():" + anlErrorWrongParam)
		return
	}
	var as analogDeviceData
	err = dev.getDataUSB(&as)
	if nil == err {
		val = as.analog[ch]
	}
	return
}

// MilliAmperToDAC переводит миллиамеры в значение для ЦАП.
// milliAmper - миллиамперы.
// maxDAC - максимально допустимое значение для ЦАП.
// maxMilliAmper - максимальное значение миллиампер, соответствующее значению maxDAC.
func MilliAmperToDAC(milliAmper float64, maxDAC uint16, maxMilliAmper uint16) uint16 {
	if 0 == maxMilliAmper {
		return maxMilliAmper
	}
	fdac := (milliAmper / float64(maxMilliAmper)) * float64(maxDAC)

	return uint16(math.RoundToEven(fdac))
}

// ValueToMa переводит величину в заданный диапазон в мА
// val - требуемое значение величины.
// maxVal - максимальное значение величины.
// minMilliAmper, maxMilliAmper - диапазон мА (0-5, 4-20).
func ValueToMa(val, maxVal float64, minMilliAmper uint16, maxMilliAmper uint16) (milliAmper float64) {
	if (0 >= maxVal) || (0 > val) || (minMilliAmper > maxMilliAmper) {
		milliAmper = float64(maxMilliAmper)
		return
	}

	diap := float64(maxMilliAmper) - float64(minMilliAmper) // диапазон изменения миллиампер
	milliAmper = (val*diap)/float64(maxVal) + float64(minMilliAmper)
	return
}

// AtToDAC переводит давление ат в значение для ЦАП.
// Под "ат" подразумевается килограмм-сила на квадратный сантиметр (кгс/см²) также называемая технической атмосферой.
// at - требуемое значение давления.
// maxAt - максимальное значение давления.
// maxDAC - максимально допустимое значение для ЦАП.
// minMilliAmper, maxMilliAmper - диапазон мА (0-5, 4-20).
func AtToDAC(at, maxAt float64, maxDAC uint16, minMilliAmper uint16, maxMilliAmper uint16) uint16 {
	if (at < 0) || (maxAt <= 0) || (minMilliAmper > maxMilliAmper) || (at > maxAt) {
		return maxDAC
	}

	diap := float64(maxMilliAmper) - float64(minMilliAmper) // диапазон изменения миллиампер
	milliAmper := (at*diap)/float64(maxAt) + float64(minMilliAmper)

	return MilliAmperToDAC(milliAmper, maxDAC, maxMilliAmper)
}

// KiloPascalToDAC переводит давление кПа в значение для ЦАП.
// kiloPascal - требуемое значение давления.
// maxKiloPascal - максимальное значение давления.
// maxDAC - максимально допустимое значение для ЦАП.
// minMilliAmper, maxMilliAmper - диапазон мА (0-5, 4-20).
func KiloPascalToDAC(kiloPascal, maxKiloPascal float64, maxDAC uint16, minMilliAmper uint16, maxMilliAmper uint16) uint16 {
	if (kiloPascal < 0) || (maxKiloPascal <= 0) || (minMilliAmper > maxMilliAmper) || (kiloPascal > maxKiloPascal) {
		return maxDAC
	}

	diap := float64(maxMilliAmper) - float64(minMilliAmper) // диапазон изменения миллиампер
	milliAmper := (kiloPascal*diap)/float64(maxKiloPascal) + float64(minMilliAmper)

	return MilliAmperToDAC(milliAmper, maxDAC, maxMilliAmper)
}

///////////////////////////////////////////////////////////////

func (dev *AnalogDevice) getDataUSB(data *analogDeviceData) (err error) {
	if nil == data || nil == dev {
		err = errors.New("AnalogDevice.getDataUSB():" + anlErrorWrongParam)
		return
	}

	if !dev.opened() {
		err = errors.New("AnalogDevice.getDataUSB():" + anlErrorNoConnection)
		return
	}

	asbytes := make([]byte, data.Size())

	err = dev.deviceIoControl(VendorRequestInput, 0xB0, asbytes, len(asbytes))

	if nil != err {
		err = errors.New("AnalogDevice.getDataUSB():" + err.Error())
		return
	}

	data.setFromBytes(asbytes)

	return
}

func (dev *AnalogDevice) setDataUSB(data *analogDeviceData) (err error) {
	if nil == data || nil == dev {
		err = errors.New("AnalogDevice.setDataUSB():" + anlErrorWrongParam)
		return
	}

	if !dev.opened() {
		err = errors.New("AnalogDevice.setDataUSB():" + anlErrorNoConnection)
		return
	}

	asbytes := data.toBytes()

	err = dev.deviceIoControl(VendorRequestOutput, 0xB0, asbytes, len(asbytes))

	if nil != err {
		err = errors.New("AnalogDevice.setDataUSB():" + err.Error())
	}
	return
}

// Open соединиться с ФАС-3
func (dev *AnalogDevice) Open() (ok bool) {
	if dev == nil {
		return
	}
	dev.handle, ok = USBOpen(IDProductANL12bit)
	if ok {
		dev.idProductVariant = IDProductANL12bit
	} else {
		dev.handle, ok = USBOpen(IDProductANL16bit)
		if ok {
			dev.idProductVariant = IDProductANL16bit
		}
	}
	return
}
