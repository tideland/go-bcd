// Tideland Go BCD
//
// Copyright (C) 2025 Frank Mueller / Tideland / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

// Package bcd provides a Binary Coded Decimal (BCD) implementation for precise
// decimal arithmetic in Go. It's particularly useful for financial and currency
// calculations where floating-point errors are unacceptable.
//
// # Overview
//
// The bcd package offers two main types:
//
//   - BCD: A general-purpose decimal number type with arbitrary precision
//   - Amount: A specialized type for monetary values with currency-specific formatting
//
// Both types guarantee exact decimal arithmetic without the rounding errors
// inherent in binary floating-point representations.
//
// # Why Use BCD?
//
// Binary floating-point numbers (float32, float64) cannot exactly represent
// many decimal fractions. This leads to subtle errors in financial calculations:
//
//	// Floating-point problem
//	fmt.Println(0.1 + 0.2)  // Output: 0.30000000000000004
//
//	// BCD solution
//	a, _ := bcd.New("0.1")
//	b, _ := bcd.New("0.2")
//	fmt.Println(a.Add(b))   // Output: 0.3
//
// # Creating BCD Numbers
//
// BCD numbers can be created from any numeric type using generics:
//
//	// From any numeric type - all using the same function
//	n1, err := bcd.New("123.45")              // from string
//	n2, err := bcd.New(12345)                 // from int
//	n3, err := bcd.New(123.45)                // from float64
//	n4, err := bcd.New(int32(123))            // from int32
//	n5, err := bcd.New(uint64(123))           // from uint64
//	n6, err := bcd.New("1.23e-4")             // scientific notation
//
//	// With options for floats
//	n7, err := bcd.New(123.456, bcd.WithScale(2))  // 123.46
//
//	// Must variant for known-good values
//	n8 := bcd.Must("123.45")  // panics on error
//
//	// Zero value
//	n9 := bcd.Zero()  // 0
//
// # Arithmetic Operations
//
// The BCD type supports all basic arithmetic operations:
//
//	a, _ := bcd.New("10.50")
//	b, _ := bcd.New("3.25")
//
//	sum := a.Add(b)                          // 13.75
//	diff := a.Sub(b)                         // 7.25
//	prod := a.Mul(b)                         // 34.125
//	quot, _ := a.Div(b, 4, bcd.RoundHalfUp) // 3.2308
//	rem, _ := a.Mod(b)                      // 0.75
//
// # Rounding
//
// The package provides seven rounding modes for fine control over decimal
// arithmetic:
//
//	n, _ := bcd.New("1.2350")
//	n.Round(2, bcd.RoundDown)      // 1.23 (truncate)
//	n.Round(2, bcd.RoundUp)        // 1.24 (round away from zero)
//	n.Round(2, bcd.RoundHalfUp)    // 1.24 (round half away from zero)
//	n.Round(2, bcd.RoundHalfDown)  // 1.23 (round half toward zero)
//	n.Round(2, bcd.RoundHalfEven)  // 1.24 (banker's rounding)
//	n.Round(2, bcd.RoundCeiling)   // 1.24 (toward +∞)
//	n.Round(2, bcd.RoundFloor)     // 1.23 (toward -∞)
//
// Banker's rounding (RoundHalfEven) is particularly useful for financial
// applications as it minimizes cumulative rounding bias.
//
// # Amount Type
//
// The Amount type combines BCD arithmetic with currency-specific features:
//
//	// Create currency amounts from any numeric type
//	price, _ := bcd.NewAmount("19.99", "USD")      // from string
//	tax, _ := bcd.NewAmount(1.60, "USD")           // from float
//	total, _ := price.Add(tax)  // $21.59
//
//	// From minor units (cents, pence, etc.)
//	amount, _ := bcd.NewAmountMinor(12345, "USD")  // 12345 cents = $123.45
//	amount2, _ := bcd.NewAmountMinor(int32(9999), "USD")  // works with any integer type
//
//	// Must variant for constants
//	amount := bcd.MustNewAmount("100.00", "USD")
//
//	// Parse formatted strings
//	c1, _ := bcd.ParseAmount("$1,234.56")    // US format
//	c2, _ := bcd.ParseAmount("€1.234,56")    // European format
//	c3, _ := bcd.ParseAmount("CHF 2'500.00") // Swiss format
//	c4, _ := bcd.ParseAmount("($50.00)")     // Negative (accounting)
//
// # Amount Operations
//
// Amount arithmetic ensures type safety and prevents mixing currencies:
//
//	usd1, _ := bcd.NewAmount("100.00", "USD")
//	usd2, _ := bcd.NewAmount("50.00", "USD")
//	eur1, _ := bcd.NewAmount("100.00", "EUR")
//
//	// Same currency operations
//	sum, _ := usd1.Add(usd2)     // OK: $150.00
//	diff, _ := usd1.Sub(usd2)    // OK: $50.00
//
//	// Different currencies
//	_, err := usd1.Add(eur1)     // Error: currency mismatch
//
//	// Multiplication/division with scalars
//	double := usd1.MulInt64(2)   // $200.00
//	half, _ := usd1.DivInt64(2)  // $50.00
//
// # Amount Allocation
//
// The package provides methods to split monetary amounts without losing pennies:
//
//	// Split evenly
//	bill, _ := bcd.NewAmount("100.00", "USD")
//	shares, _ := bill.Split(3)
//	// Results: [$33.34, $33.33, $33.33]
//	// Total: $100.00 (exact!)
//
//	// Allocate by ratios
//	budget, _ := bcd.NewAmount("1000.00", "USD")
//	allocated, _ := budget.Allocate([]int{3, 2, 5})  // 30%, 20%, 50%
//	// Results: [$300.00, $200.00, $500.00]
//
// # Supported Currencies
//
// The package includes built-in support for:
//
//   - Major fiat currencies (USD, EUR, GBP, JPY, CHF, etc.)
//   - Cryptocurrencies (BTC, ETH with appropriate decimal places)
//   - Precious metals (XAU, XAG, XPT, XPD)
//
// Each currency has the correct number of decimal places according to ISO 4217
// standards (e.g., 2 for USD, 0 for JPY, 8 for BTC).
//
// # Comparison Operations
//
// Both BCD and Amount types support comparison operations:
//
//	a, _ := bcd.New("10.50")
//	b, _ := bcd.New("10.50")
//	c, _ := bcd.New("20.00")
//
//	a.Equal(b)          // true
//	a.LessThan(c)       // true
//	c.GreaterThan(a)    // true
//	a.Cmp(c)            // -1 (a < c)
//
// # Error Handling
//
// The package defines several error types for common issues:
//
//   - ErrInvalidFormat: Invalid decimal string format
//   - ErrDivisionByZero: Attempted division by zero
//   - ErrOverflow: Arithmetic overflow
//   - ErrPrecisionLoss: Loss of precision in conversion
//   - ErrUnknownCurrency: Unknown currency code
//   - ErrCurrencyMismatch: Operation on different currencies
//   - ErrInvalidAmount: Invalid amount for currency operation
//
// # Performance Considerations
//
// BCD arithmetic is slower than native floating-point operations but provides
// exact decimal arithmetic. Use BCD when:
//
//   - Accuracy is more important than speed
//   - Working with money or financial calculations
//   - Decimal precision must be maintained exactly
//   - Rounding behavior must be predictable and controllable
//   - Regulatory compliance requires decimal arithmetic
//
// # Thread Safety
//
// BCD and Amount values are immutable. All operations return new instances
// rather than modifying existing values, making them safe for concurrent use.
// However, the types themselves are not safe for concurrent modification, so
// any shared mutable references should be protected with appropriate synchronization.
//
// # Best Practices
//
//  1. Always check errors when creating BCD numbers from strings or when
//     performing operations that can fail (division, parsing).
//
// 2. Use string literals for exact decimal values rather than float literals:
//
//		// Good
//		n, _ := bcd.New("0.1")
//
//		// OK with explicit scale (generic API handles this)
//		n, _ := bcd.New(0.1, bcd.WithScale(1))
//
//	 3. Choose appropriate rounding modes for your use case. Financial applications
//	    often use RoundHalfEven (banker's rounding) to minimize bias.
//
//	 4. When working with currencies, use the Amount type rather than plain BCD
//	    to ensure proper formatting and decimal places.
//
//	 5. For performance-critical code, minimize conversions between BCD and other
//	    numeric types. Perform calculations in BCD throughout.
//
// # Example: Invoice Calculation
//
// Here's a complete example showing invoice calculation with proper decimal handling:
//
//	// Line items - using generic API
//	widget := bcd.MustNewCurrency("49.99", "USD")
//	gadget := bcd.MustNewCurrency(129.95, "USD")  // from float
//	service := bcd.MustNewCurrency(25, "USD")      // from int
//
//	// Calculate subtotal
//	subtotal := widget.MulInt64(5)  // 5 widgets
//	temp, _ := gadget.MulInt64(2).Add(subtotal)  // 2 gadgets
//	subtotal, _ = temp.Add(service)  // 1 service
//	// Subtotal: $534.85
//
//	// Apply 10% discount
//	discountRate := bcd.Must("0.10")
//	discount := subtotal.Mul(discountRate)
//	afterDiscount, _ := subtotal.Sub(discount)
//	// After discount: $481.37
//
//	// Add 8.25% tax
//	taxRate := bcd.Must("0.0825")
//	tax := afterDiscount.Mul(taxRate)
//	total, _ := afterDiscount.Add(tax)
//	// Total: $521.08
//
// All calculations are performed with exact decimal arithmetic, ensuring
// accurate financial results without floating-point errors.
package bcd
