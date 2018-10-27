package decimal

// Abs computes ||d||.
func (d Decimal64) Abs() Decimal64 {
	return Decimal64{^neg64 & uint64(d.bits)}
}

// Add computes d + e
func (d Decimal64) Add(e Decimal64) Decimal64 {
	flavor1, sign1, exp1, significand1 := d.parts()
	flavor2, sign2, exp2, significand2 := e.parts()
	if flavor1 == flSNaN || flavor2 == flSNaN {
		return signalNaN64()
	}
	if flavor1 == flQNaN || flavor2 == flQNaN {
		return QNaN64
	}
	if flavor1 == flInf || flavor2 == flInf {
		if flavor1 != flInf {
			return e
		}
		if flavor2 != flInf || sign1 == sign2 {
			return d
		}
		return QNaN64
	}
	exp1, significand1, exp2, significand2 = matchScales(exp1, significand1, exp2, significand2)
	if sign1 == sign2 {
		significand := significand1 + significand2
		if significand >= decimal64Base {
			exp1++
			significand /= 10
		}
		return newFromParts(sign1, exp1, significand)
	}
	if significand1 > significand2 {
		return newFromParts(sign1, exp1, significand1-significand2)
	}
	return newFromParts(sign2, exp2, significand2-significand1)
}

// Cmp returns:
//
//   -2 if d or e is NaN
//   -1 if d <  e
//    0 if d == e (incl. -0 == 0, -Inf == -Inf, and +Inf == +Inf)
//   +1 if d >  e
//
func (d Decimal64) Cmp(e Decimal64) int {
	flavor1, _, _, _ := d.parts()
	flavor2, _, _, _ := e.parts()
	if flavor1 == flSNaN || flavor2 == flSNaN {
		signalNaN64()
		return 0
	}
	if flavor1 == flQNaN || flavor2 == flQNaN {
		return -2
	}
	if d == NegZero64 {
		d = Zero64
	}
	if e == NegZero64 {
		e = Zero64
	}
	if d == e {
		return 0
	}
	d = d.Sub(e)
	return 1 - 2*int(d.bits>>63)
}

// Mul computes d * e.
func (d Decimal64) Mul(e Decimal64) Decimal64 {
	flavor1, sign1, exp1, significand1 := d.parts()
	flavor2, sign2, exp2, significand2 := e.parts()

	if flavor1 == flSNaN || flavor2 == flSNaN {
		return signalNaN64()
	}
	if flavor1 == flQNaN || flavor2 == flQNaN {
		return QNaN64
	}

	sign := sign1 ^ sign2
	if d == Zero64 || d == NegZero64 || e == Zero64 || e == NegZero64 {
		return zeroes[sign]
	}
	if flavor1 == flInf || flavor2 == flInf {
		return infinities[sign]
	}

	exp := exp1 + exp2
	significand := umul64(significand1, significand2)
	for significand.hi > 0 || significand.lo >= decimal64Base {
		exp++
		significand = significand.divBy10()
	}

	return newFromParts(sign, exp, significand.lo)
}

// Neg computes -d.
func (d Decimal64) Neg() Decimal64 {
	return Decimal64{neg64 ^ uint64(d.bits)}
}

// Quo computes d / e.
func (d Decimal64) Quo(e Decimal64) Decimal64 {
	flavor1, sign1, exp1, significand1 := d.parts()
	flavor2, sign2, exp2, significand2 := e.parts()

	if flavor1 == flSNaN || flavor2 == flSNaN {
		return signalNaN64()
	}
	if flavor1 == flQNaN || flavor2 == flQNaN {
		return QNaN64
	}

	sign := sign1 ^ sign2
	if d == Zero64 || d == NegZero64 {
		if e == Zero64 || e == NegZero64 {
			return QNaN64
		}
		return zeroes[sign]
	}
	if flavor1 == flInf {
		if flavor2 == flInf {
			return QNaN64
		}
		return infinities[sign]
	}
	if flavor2 == flInf {
		return zeroes[sign]
	}

	exp := exp1 - exp2 - 16
	significand := umul64(decimal64Base, significand1).div64(significand2)
	for significand.hi > 0 || significand.lo >= decimal64Base {
		exp++
		significand = significand.divBy10()
	}

	return newFromParts(sign, exp, significand.lo)
}

// Sqrt computes √d.
func (d Decimal64) Sqrt() Decimal64 {
	flavor, sign, exp, significand := d.parts()
	switch flavor {
	case flInf:
		if sign == 1 {
			return QNaN64
		}
		return d
	case flQNaN:
		return d
	case flSNaN:
		return signalNaN64()
	case flNormal:
	}

	if significand == 0 {
		return d
	}
	if sign == 1 {
		return QNaN64
	}
	if exp&1 == 1 {
		exp--
		significand *= 10
	}
	sqrt := umul64(decimal64Base, significand).sqrt()
	exp, significand = renormalize(exp/2-8, sqrt)
	return newFromParts(sign, exp, significand)
}

// Sub returns d - e.
func (d Decimal64) Sub(e Decimal64) Decimal64 {
	return d.Add(e.Neg())
}
