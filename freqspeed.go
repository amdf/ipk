package ipk

import (
	"errors"
	"math"
	"time"
)

const frqErrorSpeedNotInitialized = `структура Speed не инициализирована`

//Speed тип для работы с заданием скорости и получением пройденного пути на ФЧС-3
type Speed struct {
	dev      *FreqDevice
	teeth    uint32
	diameter uint32
}

//Init устанавливает параметры для расчёта скорости и пути.
//dev - устройство ФЧС-3
//teeth - количество зубьев датчика скорости (например, 42)
//diameter - диаметр бандажа в мм (например, 1350 или 600)
func (sp *Speed) Init(dev *FreqDevice, teeth, diameter uint32) (err error) {
	if nil == sp || nil == dev || 0 == teeth || 0 == diameter {
		err = errors.New("Speed.setFreqUSB():" + frqErrorWrongParam)
		return
	}
	sp.dev = dev
	sp.teeth = teeth
	sp.diameter = diameter
	return
}

func (sp *Speed) initialized() bool {
	if (nil == sp) || (nil == sp.dev) {
		return false
	}
	return (0 != sp.teeth) && (0 != sp.diameter)
}

//SetLimitWay устанавливает заданный предельный путь перемещения ( в метрах ).
//Если задать этот путь, направление и скорость, то ФЧС-3 "проедет" этот путь,
//а затем остановится (скорость и ускорение станут равны 0).
//Пройденный путь можно будет получить с помощью функции GetWay
func (sp *Speed) SetLimitWay(meters uint32) (err error) {
	if !sp.initialized() {
		err = errors.New("Speed.SetLimitWay():" + frqErrorSpeedNotInitialized)
		return
	}
	c := (float64(meters) * 1000 * float64(sp.teeth)) / (math.Pi * float64(sp.diameter))
	count := uint32(c)
	ok := false
	timeout := false
	t := time.Now()
	for !ok && !timeout {
		err = sp.dev.setLimitWayUSB(count)
		ok = (nil == err)
		timeout = time.Since(t) >= maxDelayUSB
	}
	if ok && timeout {
		err = errors.New("Speed.SetLimitWay():" + errUsbTimeout)
	}

	return
}

//GetLimitWay получить заданный предельный путь перемещения (в метрах)
func (sp *Speed) GetLimitWay() (meters uint32, err error) {
	if !sp.initialized() {
		err = errors.New("Speed.GetLimitWay():" + frqErrorSpeedNotInitialized)
		return
	}
	c := (float64(sp.dev.freqdata.limitWay1) * math.Pi * float64(sp.diameter)) / (1000 * float64(sp.teeth))
	meters = uint32(math.Ceil(c))
	return
}

//GetOutputSpeed получить скорость (в км/ч) обоих генераторов частоты
func (sp *Speed) GetOutputSpeed() (kmh1, kmh2 float64, err error) {
	if !sp.initialized() {
		err = errors.New("Speed.GetOutputSpeed():" + frqErrorSpeedNotInitialized)
		return
	}

	Fout1 := (float64(sp.dev.freqdata.freq1) * magicClock) / (magicK * 4)

	kmh1 = (((Fout1 * math.Pi * float64(sp.diameter)) / float64(sp.teeth)) * 3600) / 1000000

	Fout2 := (float64(sp.dev.freqdata.freq2) * magicClock) / (magicK * 4)

	kmh2 = (((Fout2 * math.Pi * float64(sp.diameter)) / float64(sp.teeth)) * 3600) / 1000000

	return
}

//SetMotion устанавливает направление движения.
//direction - допустимые параметры: MotionOnward (вперёд), MotionBackwards (назад)
func (sp *Speed) SetMotion(direction uint8) (err error) {
	if !sp.initialized() {
		err = errors.New("Speed.SetMotion():" + frqErrorSpeedNotInitialized)
		return
	}

	switch direction {
	case MotionBackwards, MotionOnward:
		ok := false
		timeout := false
		t := time.Now()
		for !ok && !timeout {
			err = sp.dev.setMotionUSB(direction)
			ok = (nil == err)
			timeout = time.Since(t) >= maxDelayUSB
		}
		if ok && timeout {
			err = errors.New("Speed.SetMotion():" + errUsbTimeout)
		}
	default:
		err = errors.New("Speed.SetMotion():" + frqErrorWrongParam)
	}

	return
}

//GetMotion возвращает направление движения: MotionOnward (вперёд), MotionBackwards (назад)
func (sp *Speed) GetMotion() (direction uint8, err error) {
	direction = MotionUnknown
	if !sp.initialized() {
		err = errors.New("Speed.GetMotion():" + frqErrorSpeedNotInitialized)
		return
	}

	direction = sp.dev.freqdata.motion
	return
}

//SetSpeed устанавливает скорость (в км/ч) обоих генераторов
func (sp *Speed) SetSpeed(kmh1, kmh2 float64) (err error) {
	if !sp.initialized() {
		err = errors.New("Speed.SetSpeed():" + frqErrorSpeedNotInitialized)
		return
	}
	if (kmh1 < 0) || (kmh2 < 0) {
		return errors.New("Speed.SetSpeed():" + frqErrorWrongParam)
	}

	s1 := kmh1 * 1000 * 1000
	s1 = s1 / 3600
	z1 := float64(sp.teeth)
	d1 := float64(sp.diameter)
	Fout1 := (s1 * z1) / (math.Pi * d1)

	s2 := kmh2 * 1000 * 1000
	s2 = s2 / 3600
	z2 := float64(sp.teeth)
	d2 := float64(sp.diameter)
	Fout2 := (s2 * z2) / (math.Pi * d2)

	setFreq1 := (magicK * Fout1 / magicClock) * 4
	setFreq2 := (magicK * Fout2 / magicClock) * 4

	ok := false
	timeout := false
	t := time.Now()
	for !ok && !timeout {
		err = sp.dev.setFreqUSB(uint32(setFreq1), uint32(setFreq2))
		ok = (nil == err)
		timeout = time.Since(t) >= maxDelayUSB
	}
	if ok && timeout {
		err = errors.New("Speed.SetSpeed():" + errUsbTimeout)
	}

	return
}

//SetAcceleration устанавливает ускорение (в 0,01 м/с²) обоих генераторов
func (sp *Speed) SetAcceleration(accel1, accel2 float64) (err error) {
	if !sp.initialized() {
		err = errors.New("Speed.SetAcceleration():" + frqErrorSpeedNotInitialized)
		return
	}

	s1 := accel1 * 10
	z1 := float64(sp.teeth)
	d1 := float64(sp.diameter)
	Fout1 := (s1 * z1) / (math.Pi * d1) // частота в герцах

	s2 := accel2 * 10
	z2 := float64(sp.teeth)
	d2 := float64(sp.diameter)
	Fout2 := (s2 * z2) / (math.Pi * d2)

	setFreq1 := (magicK * Fout1 * 4) / magicClock
	setFreq2 := (magicK * Fout2 * 4) / magicClock

	ok := false
	timeout := false
	t := time.Now()
	for !ok && !timeout {
		err = sp.dev.setDeltaUSB(int32(setFreq1), int32(setFreq2))
		ok = (nil == err)
		timeout = time.Since(t) >= maxDelayUSB
	}
	if ok && timeout {
		err = errors.New("Speed.SetAcceleration():" + errUsbTimeout)
	}

	return
}

//GetOutputAcceleration получить ускорение (в 0,01 м/с²) обоих генераторов частоты
func (sp *Speed) GetOutputAcceleration() (accel1, accel2 float64, err error) {
	if !sp.initialized() {
		err = errors.New("Speed.GetOutputAcceleration():" + frqErrorSpeedNotInitialized)
		return
	}

	Delta1 := float64(sp.dev.freqdata.freq1delta)
	Delta2 := float64(sp.dev.freqdata.freq2delta)

	d := float64(sp.diameter)
	z := float64(sp.teeth)

	Fout1 := (Delta1 * magicClock) / (magicK * 4)
	Fout2 := (Delta2 * magicClock) / (magicK * 4)

	v1 := (Fout1 * math.Pi * d) / z
	v2 := (Fout2 * math.Pi * d) / z

	accel1 = math.RoundToEven(v1 / 10)
	accel2 = math.RoundToEven(v2 / 10)

	return
}

//GetWay получает пройденный путь в метрах с обоих генераторов
func (sp *Speed) GetWay() (way1, way2 uint32, err error) {
	if !sp.initialized() {
		err = errors.New("Speed.GetWay():" + frqErrorSpeedNotInitialized)
		return
	}
	d := float64(sp.diameter)
	n := float64(sp.teeth)
	count1 := float64(sp.dev.freqdata.way1count)
	count2 := float64(sp.dev.freqdata.way2count)
	s1 := ((math.Pi * d * count1) / n) / 1000
	s2 := ((math.Pi * d * count2) / n) / 1000

	way1 = uint32(math.Ceil(s1))
	way2 = uint32(math.Ceil(s2))

	return
}
