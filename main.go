package main

import (
	"log"

	"github.com/shopspring/decimal"
)

func main() {
	log.Println(Pmt(.07, 5, decimal.NewFromInt(100000), DueAtEnd))
	//
	log.Println(PmtWithFutureValue(.07/12, 5*12, decimal.NewFromInt(100000), decimal.Zero, DueAtEnd))
	log.Println(Pmt(.07/12, 5*12, decimal.NewFromInt(100000), DueAtEnd))

	log.Println("Total Repayable Monthly:", CalcTotalPayable(.07/Monthly, 5*Monthly, decimal.NewFromInt(100000), DueAtEnd))
	log.Println("Total Repayable Yearly:", CalcTotalPayable(.07, 5, decimal.NewFromInt(100000), DueAtEnd))

	// Verify our shit against excel
	log.Println("IPMT:", IPMT(.07, 2, 5, decimal.NewFromInt(100000), decimal.Zero, DueAtEnd))
	log.Println("PPMT:", PPMT(.07, 2, 5, decimal.NewFromInt(100000), decimal.Zero, DueAtEnd))
}
