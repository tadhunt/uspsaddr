# USPS Address Validation Library

A Go library for validating and canonicalizing US addresses using the USPS Address Validation API.

## Features

- Simple, clean API - just pass an address string
- Validates and standardizes addresses according to USPS standards
- Returns canonicalized addresses with ZIP+4
- Handles address parsing automatically
- Provides detailed validation results including corrections and warnings
- Does not expose internal USPS API types - uses custom types for better API stability

## Installation

```bash
go get github.com/tadhunt/uspsaddr
```

## Prerequisites

You need USPS API credentials to use this library:

1. Register at https://developer.usps.com/
2. Create an application to get:
   - **Client ID**
   - **Client Secret**

The library automatically handles OAuth2 token acquisition and refresh - you just provide the credentials!

## Usage

### Basic Example

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"

    "github.com/tadhunt/uspsaddr"
)

func main() {
    // Create config with your USPS credentials
    config := uspsaddr.Config{
        ClientID:     os.Getenv("USPS_CLIENT_ID"),
        ClientSecret: os.Getenv("USPS_CLIENT_SECRET"),
    }

    // Create client - it will automatically handle OAuth2 tokens
    client, err := uspsaddr.NewClient(config)
    if err != nil {
        log.Fatal(err)
    }

    // Create an address to validate
    address := &uspsaddr.Address{
        StreetAddress: "123 Main St",
        City:          "Anytown",
        State:         "CA",
        ZIPCode:       "12345",
    }

    // Validate the address
    results, err := client.ValidateAddress(context.Background(), address)
    if err != nil {
        log.Fatal(err)
    }

    // Print the canonicalized address
    for _, result := range results {
        addr := result.Address
        fmt.Printf("Street: %s\n", addr.StreetAddress)
        fmt.Printf("City: %s\n", addr.City)
        fmt.Printf("State: %s\n", addr.State)
        fmt.Printf("ZIP: %s-%s\n", addr.ZIPCode, addr.ZIPPlus4)
    }
}
```

### Address Input

Create an `Address` struct with the following fields:
- `StreetAddress` (required) - Street address
- `State` (required) - Two-letter state code
- `City` (optional if ZIP provided) - City name
- `ZIPCode` (optional) - 5-digit ZIP code
- `SecondaryAddress` (optional) - Apartment, suite, etc.
- `Firm` (optional) - Business name
- `Urbanization` (optional) - Urbanization code (Puerto Rico only)

### Using the Test Environment

To use the USPS testing environment instead of production:

```go
config := uspsaddr.Config{
    ClientID:     "your-client-id",
    ClientSecret: "your-client-secret",
    ServerURL:    "https://apis-tem.usps.com/addresses/v3",
    TokenURL:     "https://apis-tem.usps.com/oauth2/v3/token",
}

client, err := uspsaddr.NewClient(config)
```

### Checking Validation Results

```go
address := &uspsaddr.Address{
    StreetAddress: "123 Main St",
    City:          "Anytown",
    State:         "CA",
    ZIPCode:       "12345",
}

results, err := client.ValidateAddress(ctx, address)
if err != nil {
    // Handle error
    if uspsErr, ok := err.(*uspsaddr.Error); ok {
        fmt.Printf("USPS Error: %s\n", uspsErr.Detail)
        if uspsErr.Source != nil {
            fmt.Printf("Problem with: %s\n", uspsErr.Source.Parameter)
            fmt.Printf("Example: %s\n", uspsErr.Source.Example)
        }
    }
    return
}

for _, result := range results {
    // Check for corrections needed
    if len(result.Corrections) > 0 {
        fmt.Println("Address needs corrections:")
        for _, c := range result.Corrections {
            fmt.Printf("  %s: %s\n", c.Code, c.Text)
        }
    }

    // Check match quality
    if len(result.Matches) > 0 {
        for _, m := range result.Matches {
            fmt.Printf("Match: %s - %s\n", m.Code, m.Text)
        }
    }

    // Check warnings
    if len(result.Warnings) > 0 {
        fmt.Println("Warnings:")
        for _, w := range result.Warnings {
            fmt.Printf("  - %s\n", w)
        }
    }

    // Use the canonicalized address
    addr := result.Address
    fmt.Printf("%s\n%s, %s %s-%s\n",
        addr.StreetAddress,
        addr.City,
        addr.State,
        addr.ZIPCode,
        addr.ZIPPlus4,
    )
}
```

## Types

### Address

The canonicalized address structure:

```go
type Address struct {
    Firm                      string // Business name
    StreetAddress             string // Street address
    StreetAddressAbbreviation string // Abbreviated street
    SecondaryAddress          string // Apt, Suite, etc.
    City                      string // City name
    CityAbbreviation          string // Abbreviated city
    State                     string // 2-letter state code
    ZIPCode                   string // 5-digit ZIP
    ZIPPlus4                  string // 4-digit extension
    Urbanization              string // Puerto Rico only
}
```

### ValidationResult

```go
type ValidationResult struct {
    Address        Address        // Canonicalized address
    Corrections    []Correction   // How to improve input
    Matches        []Match        // Match quality indicators
    Warnings       []string       // Warning messages
    AdditionalInfo *AdditionalInfo // Delivery info
}
```

### AdditionalInfo

Additional delivery information:

```go
type AdditionalInfo struct {
    DeliveryPoint        string // Delivery point code
    CarrierRoute         string // Carrier route
    DPVConfirmation      string // Delivery point validation
    DPVCMRA              string // Commercial mail receiving agency
    Business             string // Business flag
    CentralDeliveryPoint string // Central delivery
    Vacant               string // Vacant flag
}
```

## Building

The library uses `oapi-codegen` to generate the USPS API client from the OpenAPI spec:

```bash
# Generate client and build
make

# Just generate
make generate

# Clean generated files
make clean

# Run tests
make test
```

## Development

### Prerequisites

- Go 1.23 or higher
- `oapi-codegen` tool (will be downloaded by go mod)

### Project Structure

- `types.go` - Public API types
- `client.go` - Main client implementation
- `convert.go` - Conversion between USPS and public types
- `uspsinternal/` - Generated USPS API client (not public)
- `usps-addresses-v3r2_2.yaml` - USPS OpenAPI spec

## License

See LICENSE.md

## Credits

Uses the [USPS Web Tools API](https://www.usps.com/business/web-tools-apis/).
