package ipk

import (
	"errors"
	"time"
)

const magicK = float64(0xFFFFFFFF)
const magicClock = float64(12000000) //12 МГц частота микроконтроллера

//SetHz устанавливает значение в герцах обоих генераторов частоты
func (dev *FreqDevice) SetHz(freqHz1, freqHz2 float64) (err error) {
	if nil == dev {
		err = errors.New("FreqDevice.SetHz():" + frqErrorNoDevice)
		return
	}
	if (freqHz1 < 0) || (freqHz2 < 0) {
		err = errors.New("FreqDevice.SetHz():" + frqErrorWrongParam)
		return
	}

	Freq1 := (magicK * freqHz1 / magicClock) * 4
	Freq2 := (magicK * freqHz2 / magicClock) * 4
	ok := false
	timeout := false
	t := time.Now()
	for !ok && !timeout {
		err = dev.setFreqUSB(uint32(Freq1), uint32(Freq2))
		ok = (nil == err)
		timeout = time.Since(t) >= maxDelayUSB
	}
	if ok && timeout {
		err = errors.New("SetHz():" + errUsbTimeout)
	}

	return
}

//GetOutputHz получить значение частоты в герцах обоих генераторов частоты
func (dev *FreqDevice) GetOutputHz() (hz1, hz2 float64, err error) {
	if nil == dev {
		err = errors.New("FreqDevice.GetOutputHz():" + frqErrorNoDevice)
		return
	}

	hz1 = (float64(dev.freqdata.freq1) * magicClock) / (magicK * 4)
	hz2 = (float64(dev.freqdata.freq2) * magicClock) / (magicK * 4)
	return
}

//SetDeltaHz устанавливает значение ускорения в герцах обоих генераторов частоты.
//Значения могут быть как положительными, так и отрицательными.
func (dev *FreqDevice) SetDeltaHz(deltaHz1, deltaHz2 float64) (err error) {
	if nil == dev {
		err = errors.New("FreqDevice.SetDeltaHz():" + frqErrorNoDevice)
		return
	}

	delta1 := magicK * deltaHz1 * 4 / magicClock
	delta2 := magicK * deltaHz2 * 4 / magicClock
	ok := false
	timeout := false
	t := time.Now()
	for !ok && !timeout {
		err = dev.setDeltaUSB(int32(delta1), int32(delta2))
		ok = (nil == err)
		timeout = time.Since(t) >= maxDelayUSB
	}
	if ok && timeout {
		err = errors.New("SetDeltaHz():" + errUsbTimeout)
	}
	return
}

//GetDeltaHz получить значение частоты в герцах обоих генераторов частоты
func (dev *FreqDevice) GetDeltaHz() (hz1, hz2 float64, err error) {
	if nil == dev {
		err = errors.New("FreqDevice.GetDeltaHz():" + frqErrorNoDevice)
		return
	}

	hz1 = (float64(dev.freqdata.freq1delta) * magicClock / magicK) / 4
	hz2 = (float64(dev.freqdata.freq2delta) * magicClock / magicK) / 4
	return
}
