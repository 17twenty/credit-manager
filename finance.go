package main

import (
	"math"

	"github.com/shopspring/decimal"
)

// Pmt calculates the payment for a loan based on a
// constant interest rate, period and loan amount.
// Rate (required) - the constant interest rate per period. Can be supplied as percentage or decimal number.
// For example, if you make annual payments on a loan at an annual interest rate of 10 percent, use 10% or 0.1 for rate.
// If you make monthly payments on the same loan, then use 10%/12 or 0.00833 for rate.
//
// Nper (required) - the number of payments for the loan, i.e. the total number of periods over which the loan should be paid.
// For example, if you make annual payments on a 5-year loan, supply 5 for nper.
// If you make monthly payments on the same loan, then multiply the number of years by 12, and use 5*12 or 60 for nper.
//
// Pv (required) - the present value, i.e. the total amount that all future payments are worth now. In case of a loan, it's simply the original amount borrowed.
// Fv (optional) - the future value, or the cash balance you wish to have after the last payment is made. If omitted, the future value of the loan is assumed to be zero (0).
// When (optional) - specifies when the payments are due:
// DueAtStart - payments are due at the beginning of each period.
// DueAtEnd - payments are due at the end of each period.
// For example, if you borrow $100,000 for 5 years with an annual interest rate of 7%, the following formula will calculate the annual payment:
// =PMT(7%, 5, 100000)
//
// To find the monthly payment for the same loan, use this formula:
// =PMT(7%/12, 5*12, 100000)
// = 24380.97
//
// PMT
// Months	-1,980.12
// Years	-24,389.07

type When int

const (
	DueAtStart When = iota
	DueAtEnd
)

type PaymentFrequency int

const (
	Weekly       = 52
	Monthly      = 12
	Quarterly    = 4
	SemiAnnually = 2
	Annually     = 1
)

func PmtWithFutureValue(ratePct float64, numPayments int, presentValue decimal.Decimal, futureValue decimal.Decimal, Due When) decimal.Decimal {
	pmt := 0.00

	// Floats make Nicks life easier
	pv := presentValue.InexactFloat64()
	fv := futureValue.InexactFloat64()

	pv = pv * -1
	pvMinusFV := pv - fv

	if ratePct == 0.00 {
		pmt = pvMinusFV / float64(numPayments)
	} else {
		raterPerAnnum := ratePct
		rateToNPER := math.Pow((raterPerAnnum + 1), float64(numPayments))
		pmt = (pvMinusFV * (raterPerAnnum * rateToNPER)) / (rateToNPER - 1)

		fvRate := fv * raterPerAnnum
		pmt = (pmt + fvRate)

		// one less months worth of rate is
		// being paid if start of month
		if Due == DueAtStart {
			pmt = pmt / (1 + raterPerAnnum)
		}
	}

	return decimal.NewFromFloat(pmt).RoundBank(2)
}

func Pmt(ratePct float64, numPayments int, presentValue decimal.Decimal, Due When) decimal.Decimal {
	return PmtWithFutureValue(ratePct, numPayments, presentValue, decimal.Zero, Due)
}

func CalcTotalPayable(ratePct float64, numPayments int, presentValue decimal.Decimal, Due When) decimal.Decimal {
	return Pmt(ratePct, numPayments, presentValue, Due).Mul(decimal.NewFromInt32(int32(numPayments))).Abs()
}

// IPMT returns the interest payment for a given period for a cash flow with constant periodic payments
func IPMT(rate float64, period int, numPeriods int, presentValue decimal.Decimal, futureValue decimal.Decimal, Due When) decimal.Decimal {
	interest, _ := iAndP(rate, period, numPeriods, presentValue, futureValue, Due)
	return interest.RoundBank(2)
}

// PPMT returns the principal payment for a given period for a cash flow with constant periodic payments
func PPMT(rate float64, period int, numPeriods int, presentValue decimal.Decimal, futureValue decimal.Decimal, Due When) decimal.Decimal {
	_, principal := iAndP(rate, period, numPeriods, presentValue, futureValue, Due)
	return principal.RoundBank(2)
}

// iAndP is our internal function that returns the interest and principal payment
// for a given period for a cash flow with constant periodic payments
// and interest rate.
func iAndP(rate float64, period int, numPeriods int, presentValue decimal.Decimal, futureValue decimal.Decimal, Due When) (decimal.Decimal, decimal.Decimal) {
	pmt := PmtWithFutureValue(rate, numPeriods, presentValue, futureValue, Due).InexactFloat64()

	capital := presentValue.InexactFloat64()
	var interest, principal float64
	// for loop goes brrrrrr
	for i := 1; i <= period; i++ {
		// in first period of advanced payments no interests are paid
		if Due == DueAtStart && i == 1 {
			interest = 0
		} else {
			interest = -capital * rate
		}
		principal = pmt - interest
		capital += principal
	}
	return decimal.NewFromFloat(interest), decimal.NewFromFloat(principal)
}
