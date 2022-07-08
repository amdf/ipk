package ipk

import "errors"

const anlErrorIncorrectMaxVal = `Максимальное значение для канала слишком большое`
const anlErrorInternal = `Внутренняя ошибка`

// DAC представляет один отдельный канал ЦАП на ФАС-3. Позволяет задавать значение в мА.
type DAC struct {
	device        *AnalogDevice //ФАС-3
	numChannel    uint8         //канал цап (0-13)
	maxDAC        uint16        //максимальное значение канала ЦАП
	maxMilliAmper uint16        //максимально возможное мА на канале ЦАП (зависиот от номера канала)
}

//DACKiloPascal вывод на ЦАП в килопаскалях
const DACKiloPascal = 0

//DACAtmosphere вывод на ЦАП в кгс/см² (технических атмосферах)
const DACAtmosphere = 1

//PressureOutput служит для эмуляции датчика давления на одном из каналов ЦАП.
type PressureOutput struct {
	dac *DAC // канал ЦАП на ФАС-3
	//диапазон в мА, в котором будет изменяться устанавливаемое значение
	minMilliAmperConv uint16
	maxMilliAmperConv uint16

	outputType uint8   // тип для задания в других величинах.
	maxValue   float64 // максимальное значение этой величины. минимальное примем за 0
	value      float64 //текущее установленное значение величины. только для чтения
}

//Init инициализирует канал ЦАП для дальнейшей работы с ним.
//ipkanl - устройство ФАС-3.
//numChannel - номер канала ЦАП (от ipk.DAC1 до ipk.DAC14).
func (dac *DAC) Init(device *AnalogDevice, numChannel uint8) (err error) {
	if nil == dac || nil == device || numChannel >= analogCount {
		err = errors.New("DAC.Init():" + anlErrorWrongParam)
		return
	}

	dac.device = device

	dac.numChannel = numChannel

	switch device.GetProductID() {
	case IDProductANL12bit:
		if dac.numChannel < 7 {
			dac.maxMilliAmper = 10
		} else {
			dac.maxMilliAmper = 20
		}
		dac.maxDAC = 4095
	case IDProductANL16bit:
		dac.maxMilliAmper = 20
		dac.maxDAC = 0xFFFF
	default:
		err = errors.New("DAC.Init():" + anlErrorNoConnection)
		return
	}

	return
}

//Init инициализирует эмуляцию датчика давления.
//dac - канал ЦАП ФАС-3.
//outputType - в каких единицах будет задаваться давление (ipkapi.DACOutput)
//maxValue - максимальное значение в выбранных единицах.
func (pres *PressureOutput) Init(dac *DAC, outputType uint8, maxValue float64) (err error) {
	if nil == pres || nil == dac {
		err = errors.New("PressureOutput.Init():" + anlErrorWrongParam)
		return
	}
	pres.dac = dac

	if pres.dac.numChannel < 7 {
		pres.minMilliAmperConv, pres.maxMilliAmperConv = 0, 5
	} else {
		pres.minMilliAmperConv, pres.maxMilliAmperConv = 4, 20
	}

	switch outputType {
	case DACAtmosphere, DACKiloPascal:
		pres.outputType = outputType
		pres.maxValue = maxValue
	default:
		err = errors.New("PressureOutput.Init():" + anlErrorWrongParam)
		return
	}

	return
}

//SetMilliAmper устанавливает значение на выход канала ЦАП.
//Если значение выходит за установленный максимум, вернёт ошибку.
func (dac *DAC) SetMilliAmper(val float64) (err error) {
	if nil == dac || val > float64(dac.maxMilliAmper) {
		err = errors.New("DAC.Set():" + anlErrorWrongParam)
		return
	}

	dacval := MilliAmperToDAC(val, dac.maxDAC, dac.maxMilliAmper)
	err = dac.device.setDAC(dac.numChannel, dacval)

	return
}

//Set устанавливает значение давления на выход канала ЦАП.
//Если значение выходит за установленный максимум, вернёт ошибку.
func (pres *PressureOutput) Set(val float64) (err error) {
	if nil == pres || val > pres.maxValue {
		err = errors.New("PressureOutput.Set():" + anlErrorWrongParam)
		return
	}

	switch pres.outputType {
	case DACAtmosphere, DACKiloPascal:
		maVal := ValueToMa(val, pres.maxValue, pres.minMilliAmperConv, pres.maxMilliAmperConv)
		err = pres.dac.SetMilliAmper(maVal)
	default:
		err = errors.New("PressureOutput.Set():" + anlErrorInternal)
	}

	pres.value = val

	return
}

//GetVal возвращает текущее установленное значение величины.
func (pres *PressureOutput) GetVal() float64 {
	if nil == pres {
		return 0
	}
	return pres.value
}
