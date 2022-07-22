package ipk

import (
	"errors"
)

const maxADC = 0x3FF //максимальное значение 12-битного АЦП

const frqErrorADCNoData = `Нет данных АЦП ФЧС-3`
const frqErrorADCNotEnabled = `На ФЧС-3 не включен режим АЦП`
const frqErrorADCDat1 = `Неисправен АЦП ФЧС-3 (вход ДАТ 1)`
const frqErrorADCDat2 = `Неисправен АЦП ФЧС-3 (вход ДАТ 2)`
const frqErrorADCRef = `Неисправен АЦП ФЧС-3 (эталонное значение)`

//EnableADC включает режим АЦП на ФЧС-3.
//enableADC - если true, то ФЧС-3 работает в режиме АЦП (функции задании частоты не работают),
//если false, то задание частоты работает, а АЦП нет.
func (dev *FreqDevice) EnableADC(enableADC bool) (err error) {
	if nil == dev {
		err = errors.New("FreqDevice.EnableADC():" + frqErrorNoDevice)
		return
	}

	if !dev.opened() {
		err = errors.New("FreqDevice.EnableADC():" + frqErrorNoConnection)
		return
	}

	var adcEnabled byte
	if enableADC {
		adcEnabled = 1
	}

	err = dev.deviceIoControl(VendorRequestOutput, 0xB2, []byte{adcEnabled}, 1)

	if nil != err {
		err = errors.New("FreqDevice.setEnableADC():" + err.Error())
	}
	return
}

//isADCEnabled возвращает true если включен режим АЦП.
func (dev *FreqDevice) isADCEnabled() (enabled bool, err error) {
	if nil == dev {
		err = errors.New("FreqDevice.IsADCEnabled():" + frqErrorNoDevice)
		return
	}

	if !dev.opened() {
		err = errors.New("FreqDevice.IsADCEnabled():" + frqErrorNoConnection)
		return
	}

	adcEnabled := make([]byte, 1)

	err = dev.deviceIoControl(VendorRequestInput, 0xB2, adcEnabled, 1)

	if nil != err {
		err = errors.New("FreqDevice.IsADCEnabled():" + err.Error())
		return
	}

	if adcEnabled[0] > 0 {
		enabled = true
	}

	return
}

//UpdateADC получает данные АЦП. Предварительно надо включить режим АЦП функцией SetEnableADC.
//Полученные данные доступны в поле ADC переменной типа FreqDevice.
//Функция расчитана на то, что её будут регулярно вызывать для обновления данных.
func (dev *FreqDevice) UpdateADC() (err error) {
	if nil == dev {
		err = errors.New("FreqDevice.UpdateADC():" + frqErrorNoDevice)
		return
	}

	if !dev.opened() {
		err = errors.New("FreqDevice.UpdateADC():" + frqErrorNoConnection)
		return
	}

	bdat := make([]byte, dataADCsize)

	err = dev.deviceIoControl(VendorRequestInput, 0xB1, bdat, len(bdat))

	if nil != err {
		err = errors.New("FreqDevice.UpdateADC():" + err.Error())
		return
	}

	dev.ADC.setFromBytes(bdat)
	dev.ADCModeEnabled, err = dev.isADCEnabled()

	return
}

//GetDat1ADC возвращает фильтрованное значение 12-битного АЦП MAX148 со входа ДАТ 1 ФЧС-3.
//Возвращаемое значение лежит в пределах от 0 до 0x3FF.
//Для работы этой функции предварительно надо включить режим АЦП функцией SetEnableADC(),
//а также где-то на фоне должна периодически вызываться UpdateADC().
func (dev *FreqDevice) GetDat1ADC() (dat1 uint16, err error) {
	if nil == dev {
		err = errors.New("FreqDevice.GetDat1ADC():" + frqErrorNoDevice)
		return
	}
	if 0 == dev.ADC.DivisorVal {
		err = errors.New(frqErrorADCNoData)
		return
	}
	if !dev.ADCModeEnabled {
		err = errors.New(frqErrorADCNotEnabled)
		return
	}

	rawdat1 := uint16((dev.ADC.Dat1 / uint32(dev.ADC.DivisorVal)) & 0xFFFF)

	// если значения превышают разрядность АЦП, значит АЦП неисправен
	if rawdat1 > maxADC {
		dat1 = maxADC
		err = errors.New(frqErrorADCDat1)
		return
	}

	dat1 = rawdat1

	return
}

//GetDat2ADC возвращает фильтрованное значение 12-битного АЦП MAX148 со входа ДАТ 2 ФЧС-3.
//Возвращаемое значение лежит в пределах от 0 до 0x3FF.
//Для работы этой функции предварительно надо включить режим АЦП функцией SetEnableADC(),
//а также где-то на фоне должна периодически вызываться UpdateADC().
func (dev *FreqDevice) GetDat2ADC() (dat2 uint16, err error) {
	if nil == dev {
		err = errors.New("FreqDevice.GetDat2ADC():" + frqErrorNoDevice)
		return
	}
	if 0 == dev.ADC.DivisorVal {
		err = errors.New(frqErrorADCNoData)
		return
	}
	if !dev.ADCModeEnabled {
		err = errors.New(frqErrorADCNotEnabled)
		return
	}

	rawdat2 := uint16((dev.ADC.Dat2 / uint32(dev.ADC.DivisorVal)) & 0xFFFF)

	// если значения превышают разрядность АЦП, значит АЦП неисправен
	if rawdat2 > maxADC {
		dat2 = maxADC
		err = errors.New(frqErrorADCDat2)
		return
	}

	dat2 = rawdat2

	return
}

//GetRefValADC возвращает эталонное фильтрованное значение 12-битного АЦП MAX148.
//Наличие этого значения подтверждает работоспособность ЦАПа в целом.
//Возвращаемое значение лежит в пределах от 0 до 0x3FF.
//Для работы этой функции предварительно надо включить режим АЦП функцией SetEnableADC(),
//а также где-то на фоне должна периодически вызываться UpdateADC().
func (dev *FreqDevice) GetRefValADC() (refval uint16, err error) {
	if nil == dev {
		err = errors.New("FreqDevice.GetRefValADC():" + frqErrorNoDevice)
		return
	}
	if 0 == dev.ADC.DivisorVal {
		err = errors.New(frqErrorADCNoData)
		return
	}
	if !dev.ADCModeEnabled {
		err = errors.New(frqErrorADCNotEnabled)
		return
	}

	rawRefVal := uint16((dev.ADC.ReferenceVal / uint32(dev.ADC.DivisorVal)) & 0xFFFF)

	// если значения превышают разрядность АЦП, значит АЦП неисправен
	if rawRefVal > maxADC {
		refval = maxADC
		err = errors.New(frqErrorADCRef)
		return
	}

	refval = rawRefVal

	return
}

//convertDACToMilliAmper переводит значение из АЦП ФЧС-3 в миллиамперы
// r - сопротивление резистора?
// maxU - максимальное напряжение?
// значения r и maxU:
// dat1 487, 2500;
// dat2 121, 2500;
// ref 487, 2500;
func convertDACToMilliAmper(nBitData uint32, divisor, r, maxU uint16) float64 {
	if (0 == divisor) || (0 == r) {
		return 0
	}
	data := float64(nBitData)
	koef := float64(maxU) / float64(r)
	return (data * koef) / (1023.0 * float64(divisor))
}

//GetDat1MilliAmper возвращает значение со входа ДАТ 1 ФЧС-3 в миллиамперах.
//Для работы этой функции предварительно надо включить режим АЦП функцией SetEnableADC(),
//а также где-то на фоне должна периодически вызываться UpdateADC().
func (dev *FreqDevice) GetDat1MilliAmper() (ma float64, err error) {
	if nil == dev {
		err = errors.New("FreqDevice.GetDat1ADC():" + frqErrorNoDevice)
		return
	}
	if 0 == dev.ADC.DivisorVal {
		err = errors.New(frqErrorADCNoData)
		return
	}
	if !dev.ADCModeEnabled {
		err = errors.New(frqErrorADCNotEnabled)
		return
	}
	rawdat1 := uint16((dev.ADC.Dat1 / uint32(dev.ADC.DivisorVal)) & 0xFFFF)

	// если значения превышают разрядность АЦП, значит АЦП неисправен
	if rawdat1 > maxADC {
		err = errors.New(frqErrorADCDat1)
		return
	}

	ma = convertDACToMilliAmper(dev.ADC.Dat1, dev.ADC.DivisorVal, 487, 2500)

	return
}

//GetDat2MilliAmper возвращает значение со входа ДАТ 2 ФЧС-3 в миллиамперах.
//Для работы этой функции предварительно надо включить режим АЦП функцией SetEnableADC(),
//а также где-то на фоне должна периодически вызываться UpdateADC().
func (dev *FreqDevice) GetDat2MilliAmper() (ma float64, err error) {
	if nil == dev {
		err = errors.New("FreqDevice.GetDat2ADC():" + frqErrorNoDevice)
		return
	}
	if 0 == dev.ADC.DivisorVal {
		err = errors.New(frqErrorADCNoData)
		return
	}
	if !dev.ADCModeEnabled {
		err = errors.New(frqErrorADCNotEnabled)
		return
	}
	rawdat2 := uint16((dev.ADC.Dat2 / uint32(dev.ADC.DivisorVal)) & 0xFFFF)

	// если значения превышают разрядность АЦП, значит АЦП неисправен
	if rawdat2 > maxADC {
		err = errors.New(frqErrorADCDat2)
		return
	}

	ma = convertDACToMilliAmper(dev.ADC.Dat2, dev.ADC.DivisorVal, 121, 2500)

	return
}

//GetRefValMilliAmper возвращает эталонное значение АЦП ФЧС-3 в миллиамперах.
//Для работы этой функции предварительно надо включить режим АЦП функцией SetEnableADC(),
//а также где-то на фоне должна периодически вызываться UpdateADC().
func (dev *FreqDevice) GetRefValMilliAmper() (ma float64, err error) {
	if nil == dev {
		err = errors.New("FreqDevice.GetRefValADC():" + frqErrorNoDevice)
		return
	}
	if 0 == dev.ADC.DivisorVal {
		err = errors.New(frqErrorADCNoData)
		return
	}
	if !dev.ADCModeEnabled {
		err = errors.New(frqErrorADCNotEnabled)
		return
	}
	rawRefVal := uint16((dev.ADC.ReferenceVal / uint32(dev.ADC.DivisorVal)) & 0xFFFF)

	// если значения превышают разрядность АЦП, значит АЦП неисправен
	if rawRefVal > maxADC {

		err = errors.New(frqErrorADCRef)
		return
	}

	ma = convertDACToMilliAmper(dev.ADC.ReferenceVal, dev.ADC.DivisorVal, 487, 2500)

	return
}
