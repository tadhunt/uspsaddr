package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/tadhunt/uspsaddr"
)

func main() {
	// Get credentials from environment
	clientID := os.Getenv("USPS_CLIENT_ID")
	clientSecret := os.Getenv("USPS_CLIENT_SECRET")

	if clientID == "" || clientSecret == "" {
		log.Fatal("USPS_CLIENT_ID and USPS_CLIENT_SECRET environment variables are required")
	}

	// Create config
	config := uspsaddr.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		// ServerURL and TokenURL will use production defaults if not specified
	}

	// Create client (will automatically handle OAuth2 token acquisition and refresh)
	client, err := uspsaddr.NewClient(config)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Example addresses to validate
	addresses := []*uspsaddr.Address{
		{
			Firm:          "Chipotle",
			StreetAddress: "28th",
			City:          "boulder",
			State:         "co",
		},
		{
			StreetAddress: "100 broadway",
//			City:          "boulder",
//			State:         "co",
		},
		{
			StreetAddress: "100 broadway",
//			City:          "boulder",
			State:         "co",
		},
		{
			StreetAddress: "100 broadway",
			City:          "boulder",
			State:         "co",
		},
		{
			StreetAddress: "1600 Pennsylvania Ave NW",
			City:          "Washington",
			State:         "DC",
			ZIPCode:       "20500",
		},
		{
			StreetAddress: "350 Fifth Avenue",
			City:          "New York",
			State:         "NY",
			ZIPCode:       "10118",
		},
		{
			StreetAddress: "1 Apple Park Way",
			City:          "Cupertino",
			State:         "CA",
			ZIPCode:       "95014",
		},
	}

	for i, address := range addresses {
		fmt.Printf("\n=== Example %d ===\n", i+1)
		fmt.Printf("Input: %s, %s, %s %s\n", address.StreetAddress, address.City, address.State, address.ZIPCode)

		results, err := client.ValidateAddress(context.Background(), address)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			if uspsErr, ok := err.(*uspsaddr.Error); ok {
				fmt.Printf("USPS Error: status %q code %q title %q detail %q\n", uspsErr.Status, uspsErr.Code, uspsErr.Title, uspsErr.Detail)
				if uspsErr.Source != nil {
					fmt.Printf("  Parameter: %s\n", uspsErr.Source.Parameter)
					fmt.Printf("  Example: %s\n", uspsErr.Source.Example)
				}
			}
			continue
		}

		for j, result := range results {
			if len(results) > 1 {
				fmt.Printf("\nResult %d:\n", j+1)
			}

			addr := result.Address
			fmt.Printf("\nCanonical Address:\n")
			if addr.Firm != "" {
				fmt.Printf("  Firm: %s\n", addr.Firm)
			}
			fmt.Printf("  Street: %s\n", addr.StreetAddress)
			if addr.SecondaryAddress != "" {
				fmt.Printf("  Secondary: %s\n", addr.SecondaryAddress)
			}
			fmt.Printf("  City: %s\n", addr.City)
			fmt.Printf("  State: %s\n", addr.State)
			fmt.Printf("  ZIP: %s", addr.ZIPCode)
			if addr.ZIPPlus4 != "" {
				fmt.Printf("-%s", addr.ZIPPlus4)
			}
			fmt.Println()

			if len(result.Matches) > 0 {
				fmt.Println("\nMatches:")
				for _, m := range result.Matches {
					fmt.Printf("  [%s] %s\n", m.Code, m.Text)
				}
			}

			if len(result.Corrections) > 0 {
				fmt.Println("\nCorrections Needed:")
				for _, c := range result.Corrections {
					fmt.Printf("  [%s] %s\n", c.Code, c.Text)
				}
			}

			if len(result.Warnings) > 0 {
				fmt.Println("\nWarnings:")
				for _, w := range result.Warnings {
					fmt.Printf("  - %s\n", w)
				}
			}

			if result.AdditionalInfo != nil {
				info := result.AdditionalInfo
				if info.DPVConfirmation != "" || info.CarrierRoute != "" {
					fmt.Println("\nAdditional Info:")
					if info.DPVConfirmation != "" {
						fmt.Printf("  DPV Confirmation: %s\n", info.DPVConfirmation)
					}
					if info.CarrierRoute != "" {
						fmt.Printf("  Carrier Route: %s\n", info.CarrierRoute)
					}
					if info.Business != "" {
						fmt.Printf("  Business: %s\n", info.Business)
					}
				}
			}
		}
	}
}
