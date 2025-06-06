# BCD - Binary Coded Decimal Package for Go

[![Go Reference](https://pkg.go.dev/badge/tideland.dev/go/bcd.svg)](https://pkg.go.dev/tideland.dev/go/bcd)

Package `bcd` provides a Binary Coded Decimal (BCD) implementation for precise decimal arithmetic in Go. It's particularly useful for financial and currency calculations where floating-point errors are unacceptable.

## Features

- **Exact Decimal Arithmetic**: No floating-point rounding errors
- **Currency Support**: Built-in support for ISO 4217 currency codes
- **Multiple Rounding Modes**: Including banker's rounding (round half to even)
- **Arbitrary Precision**: Handle very large and very small numbers
- **Currency Allocation**: Split amounts without losing pennies
- **International Format Parsing**: Parse various currency formats (e.g., $1,234.56 or €1.234,56)

## Installation

```bash
go get tideland.dev/go/bcd
```

## Quick Start

### Basic BCD Arithmetic

```go
package main

import (
    "fmt"
    "tideland.dev/go/bcd"
)

func main() {
    // Create BCD numbers from strings
    a, _ := bcd.New("123.45")
    b, _ := bcd.New("67.89")
    
    // Perform exact arithmetic
    sum := a.Add(b)
    difference := a.Sub(b)
    product := a.Mul(b)
    quotient, _ := a.Div(b, 4, bcd.RoundHalfUp)
    
    fmt.Println("Sum:", sum)         // Sum: 191.34
    fmt.Println("Product:", product)  // Product: 8381.2305
}
```

### Currency Handling

```go
// Create currency amounts
price, _ := bcd.NewCurrency("19.99", "USD")
tax, _ := bcd.NewCurrency("1.60", "USD")

// Calculate total
total, _ := price.Add(tax)
fmt.Println("Total:", total)  // Total: $21.59

// Parse formatted currency strings
amount, _ := bcd.ParseCurrency("$1,234.56")
fmt.Println(amount)  // $1234.56
```

## Why Use BCD?

### The Floating-Point Problem

```go
// With float64
fmt.Println(0.1 + 0.2)  // 0.30000000000000004

// With BCD
a, _ := bcd.New("0.1")
b, _ := bcd.New("0.2")
fmt.Println(a.Add(b))  // 0.3
```

### Perfect for Money

```go
// Split $100 among 3 people
bill, _ := bcd.NewCurrency("100.00", "USD")
shares, _ := bill.Split(3)

// Results:
// Person 1: $33.34
// Person 2: $33.33
// Person 3: $33.33
// Total: $100.00 (no pennies lost!)
```

## Core Types

### BCD

The `BCD` type represents a decimal number using binary coded decimal encoding:

- `New(string) (*BCD, error)` - Create from string
- `NewFromInt(int64) *BCD` - Create from integer
- `NewFromFloat(float64, int) (*BCD, error)` - Create from float with precision
- `Zero() *BCD` - Create zero value

### Currency

The `Currency` type wraps BCD with currency-specific features:

- `NewCurrency(amount, code string) (*Currency, error)` - Create currency amount
- `NewCurrencyFromInt(minorUnits int64, code string) (*Currency, error)` - Create from minor units (cents)
- `ParseCurrency(string) (*Currency, error)` - Parse formatted currency string

## Arithmetic Operations

### BCD Operations

```go
a.Add(b)                    // Addition
a.Sub(b)                    // Subtraction  
a.Mul(b)                    // Multiplication
a.Div(b, scale, rounding)   // Division with scale and rounding
a.Mod(b)                    // Modulo
a.Abs()                     // Absolute value
a.Neg()                     // Negation
```

### Currency Operations

```go
curr1.Add(curr2)            // Add (same currency only)
curr1.Sub(curr2)            // Subtract (same currency only)
curr.Mul(factor)            // Multiply by BCD
curr.MulInt64(n)            // Multiply by integer
curr.Div(divisor)           // Divide by BCD
curr.DivInt64(n)            // Divide by integer
```

## Rounding Modes

The package supports multiple rounding modes:

- `RoundDown` - Round towards zero (truncate)
- `RoundUp` - Round away from zero
- `RoundHalfUp` - Round to nearest, ties away from zero
- `RoundHalfDown` - Round to nearest, ties towards zero
- `RoundHalfEven` - Round to nearest, ties to even (banker's rounding)
- `RoundCeiling` - Round towards positive infinity
- `RoundFloor` - Round towards negative infinity

```go
value, _ := bcd.New("1.2350")
rounded := value.Round(2, bcd.RoundHalfEven)  // 1.24
```

## Currency Allocation

Distribute amounts without losing pennies:

```go
// Split evenly
total, _ := bcd.NewCurrency("100.00", "USD")
parts, _ := total.Split(3)
// Results: $33.34, $33.33, $33.33

// Allocate by ratios
rent, _ := bcd.NewCurrency("2000.00", "USD")
shares, _ := rent.Allocate([]int{1, 2, 2})  // 1:2:2 ratio
// Results: $400.00, $800.00, $800.00
```

## Supported Currencies

The package includes built-in support for major world currencies:

- Fiat: USD, EUR, GBP, JPY, CHF, CAD, AUD, CNY, and many more
- Crypto: BTC, ETH (with 8 decimal places)
- Precious metals: XAU (gold), XAG (silver), XPT (platinum), XPD (palladium)

Each currency has the correct number of decimal places (e.g., 2 for USD, 0 for JPY).

## Format Parsing

Parse various international currency formats:

```go
amounts := []string{
    "$1,234.56",      // US format
    "€1.234,56",      // European format
    "CHF 2'500.00",   // Swiss format
    "($50.00)",       // Negative (accounting)
    "¥1,234",         // Japanese (no decimals)
}

for _, s := range amounts {
    curr, _ := bcd.ParseCurrency(s)
    fmt.Println(curr)
}
```

## Error Handling

The package defines several error types:

- `ErrDivisionByZero` - Division by zero attempted
- `ErrInvalidFormat` - Invalid decimal string format
- `ErrOverflow` - Arithmetic overflow
- `ErrUnknownCurrency` - Unknown currency code
- `ErrCurrencyMismatch` - Operation on different currencies

## Performance Considerations

BCD arithmetic is slower than native floating-point operations but provides exact decimal arithmetic. Use BCD when:

- Accuracy is more important than speed
- Working with money or financial calculations
- Decimal precision must be maintained
- Rounding behavior must be predictable

## Examples

See the [examples_test.go](examples_test.go) file for comprehensive examples including:

- Basic arithmetic operations
- Currency calculations
- Invoice generation
- International pricing
- Allocation and splitting
- Various rounding modes

## License

Copyright (c) 2024, Frank Mueller / Tideland
All rights reserved.

[License details here]