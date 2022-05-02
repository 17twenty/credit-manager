package main

import (
	"log"

	"github.com/shopspring/decimal"
)

// SimpleInterest is used for non-compounding interest
// It doesn’t take into account the effect of compounding.
// In many cases, interest compounds with each designated period of a loan,
// but in the case of simple interest, it does not.
// The calculation of simple interest is equal to the principal amount multiplied by the
// interest rate, multiplied by the number of periods.
//
// Do NOT call this using integer math inline, i.e.
// SimpleInterest(..., 0.10, 3/12 )
// will fail due to Go's conventions
func SimpleInterest(principal decimal.Decimal, ratePct float64, period float64) decimal.Decimal {
	return decimal.NewFromFloat(principal.InexactFloat64() * ratePct * period)
}

type Action int

const (
	Credit Action = iota
	Debit
)

func (a Action) String() string {
	if a == Credit {
		return "Credit"
	}
	return "Debit"
}

type transaction struct {
	day       int
	amount    decimal.Decimal
	direction Action
	// -- Unused Below
	drawDownFee   decimal.Decimal
	drawDownPct   decimal.Decimal
	billingPeriod int
	daysInYear    int
}

type Loan struct {
	creditLimit  decimal.Decimal
	apr          float64
	transactions []transaction
}

func NewLoan(creditLimit decimal.Decimal, APR float64) Loan {
	return Loan{
		creditLimit:  creditLimit,
		apr:          APR,
		transactions: []transaction{},
	}
}

func (l *Loan) Draw(amount decimal.Decimal, day int) {
	l.transactions = append(l.transactions, transaction{
		day:       day,
		amount:    amount,
		direction: Debit,
	})
}

func (l *Loan) Pay(amount decimal.Decimal, day int) {
	l.transactions = append(l.transactions, transaction{
		day:       day,
		amount:    amount,
		direction: Credit,
	})
}

func (l *Loan) getTransactions(day int) []transaction {
	txns := []transaction{}
	for i := range l.transactions {
		if l.transactions[i].day == day {
			txns = append(txns, l.transactions[i])
		}
	}
	return txns
}

func (l *Loan) dumpTransactions(day int) {
	balance := decimal.Zero
	var accInterest decimal.Decimal
	log.Println("Day\t Action \t Amount")
	for i := 0; i <= day; i++ {
		txns := l.getTransactions(i)
		for j := range txns {
			switch txns[j].direction {
			case Credit:
				balance = balance.Sub(txns[j].amount.Abs())
				log.Printf("%d \t %v \t %v", i, Credit, txns[j].amount)
			case Debit:
				balance = balance.Sub(txns[j].amount.Abs())
				log.Printf("%d \t %v \t (%v)", i, Debit, txns[j].amount)
			}
		}
		if balance.GreaterThan(decimal.Zero) {
			log.Println("Carried interest on day", i)
			dailyAPR := (l.apr / 365.00)
			accInterest = accInterest.Add(balance.Mul(decimal.NewFromFloat(dailyAPR)))
		}
		log.Println("____________________________________________")
		log.Println("Balance at end of day", i, "", balance)
		log.Println("Interest owed", accInterest)
	}
}

func (l *Loan) GetBalance(day int) decimal.Decimal {
	balance := decimal.Zero
	for i := 0; i <= day; i++ {
		txns := l.getTransactions(i)
		for j := range txns {
			switch txns[j].direction {
			case Credit:
				balance = balance.Sub(txns[j].amount.Abs())
			case Debit:
				balance = balance.Add(txns[j].amount.Abs())
			}
		}
	}
	return l.creditLimit.Sub(balance)
}

func (l *Loan) GetLimitAndBalance(day int) (decimal.Decimal, decimal.Decimal) {
	return l.creditLimit.Sub(l.GetBalance(day)), l.GetBalance(day)
}

func (l *Loan) GetInterestOwed(day int) decimal.Decimal {
	var accInterest decimal.Decimal

	// Starting at start of billing period
	// We zip through and accumulate interest for the days we hold a balance
	balance := decimal.Zero
	for i := 0; i <= day; i++ {
		txns := l.getTransactions(i)
		for j := range txns {
			switch txns[j].direction {
			case Credit:
				balance = balance.Sub(txns[j].amount.Abs())
			case Debit:
				balance = balance.Add(txns[j].amount.Abs())
			}
		}
		if balance.GreaterThan(decimal.Zero) {
			dailyAPR := (l.apr / 365.00)
			accInterest = accInterest.Add(balance.Mul(decimal.NewFromFloat(dailyAPR)))
		}
	}

	return accInterest.RoundBank(2)
}
