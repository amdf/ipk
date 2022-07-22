package ipk

import "errors"

const freqCount = 4

//4 частотных выхода на ФАС-3
const (
	FREQ1 = iota
	FREQ2
	FREQ3
	FREQ4
)

// Константы для функции SetPredefinedFreq
const (
	AnlFreq200Hz = 3981
	AnlFreq500Hz = 3993
	AnlFreq1kHz  = 3997
	AnlFreq2kHz  = 3999
	AnlFreq4kHz  = 4000
)

//SetFreq выводит на выход ВЫХ.ЧС-БУС одно из заранее заданных значений частоты.
// Параметр ch - номер канала. Значение от ipk.FREQ1 до ipk.FREQ4
// Параметр predefinedVal - (см. константы ipk.AnlFreq).
func (dev *AnalogDevice) SetFreq(ch uint8, predefinedVal uint16) (err error) {
	if nil == dev || ch >= freqCount {
		err = errors.New("SetFreq():" + anlErrorWrongParam)
		return
	}
	var as analogDeviceData
	err = dev.getDataUSB(&as)
	if nil == err {
		as.freq[ch] = predefinedVal
		err = dev.setDataUSB(&as)
	}
	return
}

//GetOutputFreq позволяет узнать, какое значение частоты установлено в данный момент
//на одном из выходов ВЫХ.ЧС-БУС. Значение следует сравнивать с константами ipk.AnlFreq.
//Параметр ch - номер канала. Значение от ipk.FREQ1 до ipk.FREQ4
func (dev *AnalogDevice) GetOutputFreq(ch uint8) (val uint16, err error) {
	if nil == dev || ch >= freqCount {
		err = errors.New("GetOutputFreq():" + anlErrorWrongParam)
		return
	}
	var as analogDeviceData
	err = dev.getDataUSB(&as)
	if nil == err {
		val = as.freq[ch]
	}
	return
}
