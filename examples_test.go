// Copyright (c) 2024, Frank Mueller / Tideland
// All rights reserved.

package bcd_test

import (
	"fmt"
	"log"

	"tideland.dev/go/bcd"
)

// ExampleBCD_basic demonstrates basic BCD creation and arithmetic.
func ExampleBCD_basic() {
	// Create BCD numbers from strings
	a, _ := bcd.New("123.45")
	b, _ := bcd.New("67.89")

	// Basic arithmetic
	sum := a.Add(b)
	difference := a.Sub(b)
	product := a.Mul(b)
	quotient, _ := a.Div(b, 4, bcd.RoundHalfUp)

	fmt.Println("a =", a)
	fmt.Println("b =", b)
	fmt.Println("a + b =", sum)
	fmt.Println("a - b =", difference)
	fmt.Println("a * b =", product)
	fmt.Println("a / b =", quotient)

	// Output:
	// a = 123.45
	// b = 67.89
	// a + b = 191.34
	// a - b = 55.56
	// a * b = 8381.2305
	// a / b = 1.8178
}

// ExampleBCD_precision demonstrates how BCD maintains exact decimal precision.
func ExampleBCD_precision() {
	// Classic floating-point problem: 0.1 + 0.2
	a, _ := bcd.New("0.1")
	b, _ := bcd.New("0.2")
	sum := a.Add(b)

	fmt.Printf("0.1 + 0.2 = %s (exact)\n", sum)

	// Compare with float
	floatSum := 0.1 + 0.2
	fmt.Printf("0.1 + 0.2 = %.17f (float)\n", floatSum)

	// Repeated additions
	penny, _ := bcd.New("0.01")
	total := bcd.Zero()
	for range 100 {
		total = total.Add(penny)
	}
	fmt.Printf("100 * 0.01 = %s\n", total)

	// Output:
	// 0.1 + 0.2 = 0.3 (exact)
	// 0.1 + 0.2 = 0.29999999999999999 (float)
	// 100 * 0.01 = 1
}

// ExampleCurrency_basic demonstrates basic currency operations.
func ExampleCurrency_basic() {
	// Create currency amounts
	price, _ := bcd.NewCurrency("19.99", "USD")
	tax, _ := bcd.NewCurrency("1.60", "USD")

	// Calculate total
	total, _ := price.Add(tax)

	fmt.Println("Price:", price)
	fmt.Println("Tax:", tax)
	fmt.Println("Total:", total)

	// Different formatting options
	fmt.Println("With code:", total.Format(true, true))
	fmt.Println("Symbol only:", total.Format(true, false))
	fmt.Println("Plain:", total.Format(false, false))

	// Output:
	// Price: $19.99
	// Tax: $1.60
	// Total: $21.59
	// With code: $21.59 USD
	// Symbol only: $21.59
	// Plain: 21.59
}

// ExampleCurrency_calculations demonstrates currency calculations.
func ExampleCurrency_calculations() {
	// Shopping cart example
	item1, _ := bcd.NewCurrency("29.99", "USD")
	item2, _ := bcd.NewCurrency("45.50", "USD")
	item3, _ := bcd.NewCurrency("12.99", "USD")

	// Calculate subtotal
	subtotal, _ := item1.Add(item2)
	subtotal, _ = subtotal.Add(item3)

	// Apply discount (15% off)
	discountRate, _ := bcd.New("0.15")
	discount := subtotal.Mul(discountRate)
	afterDiscount, _ := subtotal.Sub(discount)

	// Calculate tax (8.5%)
	taxRate, _ := bcd.New("0.085")
	tax := afterDiscount.Mul(taxRate)
	total, _ := afterDiscount.Add(tax)

	fmt.Println("Subtotal:", subtotal)
	fmt.Println("Discount (15%):", discount)
	fmt.Println("After discount:", afterDiscount)
	fmt.Println("Tax (8.5%):", tax)
	fmt.Println("Total:", total)

	// Output:
	// Subtotal: $88.48
	// Discount (15%): $13.27
	// After discount: $75.21
	// Tax (8.5%): $6.39
	// Total: $81.60
}

// ExampleCurrency_allocation demonstrates splitting amounts without losing pennies.
func ExampleCurrency_allocation() {
	// Split a bill among friends
	bill, _ := bcd.NewCurrency("100.00", "USD")

	// Split evenly among 3 people
	shares, _ := bill.Split(3)
	fmt.Println("Splitting $100 among 3 people:")
	for i, share := range shares {
		fmt.Printf("  Person %d: %s\n", i+1, share)
	}

	// Verify total
	total := bcd.Zero()
	for _, share := range shares {
		total = total.Add(share.Amount())
	}
	fmt.Printf("  Total: $%s (no pennies lost!)\n", total)

	// Allocate by ratios (e.g., splitting rent by room size)
	rent, _ := bcd.NewCurrency("2000.00", "USD")
	// Room sizes: 100, 150, 250 sq ft
	roomShares, _ := rent.Allocate([]int{100, 150, 250})
	fmt.Println("\nSplitting $2000 rent by room size:")
	sizes := []int{100, 150, 250}
	for i, share := range roomShares {
		fmt.Printf("  Room %d (%d sq ft): %s\n", i+1, sizes[i], share)
	}

	// Output:
	// Splitting $100 among 3 people:
	//   Person 1: $33.34
	//   Person 2: $33.33
	//   Person 3: $33.33
	//   Total: $100.00 (no pennies lost!)
	// 
	// Splitting $2000 rent by room size:
	//   Room 1 (100 sq ft): $400.00
	//   Room 2 (150 sq ft): $600.00
	//   Room 3 (250 sq ft): $1000.00
}

// ExampleParseCurrency demonstrates parsing formatted currency strings.
func ExampleParseCurrency() {
	inputs := []string{
		"$1,234.56",
		"€1.234,56",
		"¥1,234",
		"USD 999.99",
		"CHF 2'500.00",
		"($50.00)",  // Negative amount
	}

	for _, input := range inputs {
		curr, err := bcd.ParseCurrency(input)
		if err != nil {
			fmt.Printf("Error parsing %q: %v\n", input, err)
			continue
		}
		fmt.Printf("%-15s -> %s (%s)\n", input, curr, curr.Code())
	}

	// Output:
	// $1,234.56       -> $1234.56 (USD)
	// €1.234,56       -> €1234.56 (EUR)
	// ¥1,234          -> ¥1234 (JPY)
	// USD 999.99      -> $999.99 (USD)
	// CHF 2'500.00    -> Fr2500.00 (CHF)
	// ($50.00)        -> -$50.00 (USD)
}

// ExampleCurrency_internationalPrices demonstrates handling multiple currencies.
func ExampleCurrency_internationalPrices() {
	// Product prices in different markets
	priceUS, _ := bcd.NewCurrency("99.99", "USD")
	priceEU, _ := bcd.NewCurrency("89.99", "EUR")
	priceUK, _ := bcd.NewCurrency("79.99", "GBP")
	priceJP, _ := bcd.NewCurrency("10000", "JPY")

	fmt.Println("International Pricing:")
	fmt.Printf("  US: %s\n", priceUS.Format(true, true))
	fmt.Printf("  EU: %s\n", priceEU.Format(true, true))
	fmt.Printf("  UK: %s\n", priceUK.Format(true, true))
	fmt.Printf("  JP: %s\n", priceJP.Format(true, true))

	// Note: Currency conversion would require exchange rates
	// This package focuses on accurate arithmetic within a single currency

	// Output:
	// International Pricing:
	//   US: $99.99 USD
	//   EU: €89.99 EUR
	//   UK: £79.99 GBP
	//   JP: ¥10000 JPY
}

// ExampleBCD_rounding demonstrates different rounding modes.
func ExampleBCD_rounding() {
	value, _ := bcd.New("1.2350")

	modes := []struct {
		name string
		mode bcd.RoundingMode
	}{
		{"RoundUp", bcd.RoundUp},
		{"RoundDown", bcd.RoundDown},
		{"RoundHalfUp", bcd.RoundHalfUp},
		{"RoundHalfDown", bcd.RoundHalfDown},
		{"RoundHalfEven", bcd.RoundHalfEven},
		{"RoundCeiling", bcd.RoundCeiling},
		{"RoundFloor", bcd.RoundFloor},
	}

	fmt.Printf("Rounding %s to 2 decimal places:\n", value)
	for _, m := range modes {
		rounded := value.Round(2, m.mode)
		fmt.Printf("  %-15s: %s\n", m.name, rounded)
	}

	// Banker's rounding examples
	fmt.Println("\nBanker's rounding (RoundHalfEven):")
	examples := []string{"1.25", "1.35", "2.25", "2.35"}
	for _, ex := range examples {
		v, _ := bcd.New(ex)
		rounded := v.Round(1, bcd.RoundHalfEven)
		fmt.Printf("  %s -> %s\n", ex, rounded)
	}

	// Output:
	// Rounding 1.2350 to 2 decimal places:
	//   RoundUp        : 1.24
	//   RoundDown      : 1.23
	//   RoundHalfUp    : 1.24
	//   RoundHalfDown  : 1.23
	//   RoundHalfEven  : 1.24
	//   RoundCeiling   : 1.24
	//   RoundFloor     : 1.23
	// 
	// Banker's rounding (RoundHalfEven):
	//   1.25 -> 1.2
	//   1.35 -> 1.4
	//   2.25 -> 2.2
	//   2.35 -> 2.4
}

// ExampleCurrency_minorUnits demonstrates working with minor units (cents, pence, etc).
func ExampleCurrency_minorUnits() {
	// Create from minor units
	cents := int64(12345)
	amount, _ := bcd.NewCurrencyFromInt(cents, "USD")
	fmt.Printf("%d cents = %s\n", cents, amount)

	// Convert to minor units
	price, _ := bcd.NewCurrency("99.99", "USD")
	minorUnits, _ := price.ToMinorUnits()
	fmt.Printf("%s = %d cents\n", price, minorUnits)

	// Japanese Yen has no minor units
	yen, _ := bcd.NewCurrency("1234", "JPY")
	yenUnits, _ := yen.ToMinorUnits()
	fmt.Printf("%s = %d yen\n", yen, yenUnits)

	// Output:
	// 12345 cents = $123.45
	// $99.99 = 9999 cents
	// ¥1234 = 1234 yen
}

// ExampleCurrency_invoice demonstrates a complete invoice calculation.
func ExampleCurrency_invoice() {
	// Line items
	type LineItem struct {
		Description string
		Quantity    int64
		UnitPrice   *bcd.Currency
	}

	items := []LineItem{
		{"Widget Pro", 5, mustCurrency("49.99", "USD")},
		{"Gadget Plus", 2, mustCurrency("129.95", "USD")},
		{"Service Fee", 1, mustCurrency("25.00", "USD")},
	}

	// Calculate line totals
	fmt.Println("Invoice:")
	fmt.Println(repeatString("-", 50))
	
	var subtotal *bcd.Currency
	for _, item := range items {
		lineTotal := item.UnitPrice.MulInt64(item.Quantity)
		fmt.Printf("%-20s %2d x %8s = %8s\n", 
			item.Description, item.Quantity, item.UnitPrice, lineTotal)
		
		if subtotal == nil {
			subtotal = lineTotal
		} else {
			subtotal, _ = subtotal.Add(lineTotal)
		}
	}

	fmt.Println(repeatString("-", 50))
	fmt.Printf("%-35s %8s\n", "Subtotal:", subtotal)

	// Apply discount
	discountPercent, _ := bcd.New("0.10") // 10% discount
	discount := subtotal.Mul(discountPercent)
	afterDiscount, _ := subtotal.Sub(discount)
	fmt.Printf("%-35s %8s\n", "Discount (10%):", discount.Neg())
	
	// Calculate tax
	taxRate, _ := bcd.New("0.0825") // 8.25% tax
	tax := afterDiscount.Mul(taxRate)
	total, _ := afterDiscount.Add(tax)
	
	fmt.Printf("%-35s %8s\n", "Tax (8.25%):", tax)
	fmt.Println(repeatString("=", 50))
	fmt.Printf("%-35s %8s\n", "Total:", total)

	// Output:
	// Invoice:
	// --------------------------------------------------
	// Widget Pro           5 x   $49.99 =  $249.95
	// Gadget Plus          2 x  $129.95 =  $259.90
	// Service Fee          1 x   $25.00 =   $25.00
	// --------------------------------------------------
	// Subtotal:                           $534.85
	// Discount (10%):                     -$53.49
	// Tax (8.25%):                         $39.71
	// ==================================================
	// Total:                              $521.07
}

// Helper functions for examples
func mustCurrency(amount, code string) *bcd.Currency {
	c, err := bcd.NewCurrency(amount, code)
	if err != nil {
		log.Fatal(err)
	}
	return c
}

func repeatString(s string, n int) string {
	result := ""
	for range n {
		result += s
	}
	return result
}