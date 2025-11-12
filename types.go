package uspsaddr

// Address represents a canonicalized USPS address
type Address struct {
	// Firm/business name at the address
	Firm string

	// Street address (primary line)
	StreetAddress string

	// Abbreviated street address
	StreetAddressAbbreviation string

	// Secondary address (apartment, suite, etc.)
	SecondaryAddress string

	// City name
	City string

	// Abbreviated city name
	CityAbbreviation string

	// Two-letter state code
	State string

	// 5-digit ZIP code
	ZIPCode string

	// 4-digit ZIP+4 extension
	ZIPPlus4 string

	// Urbanization code (Puerto Rico addresses)
	Urbanization string
}

// ValidationResult contains the results of address validation
type ValidationResult struct {
	// The canonicalized address
	Address Address

	// Codes indicating how to improve the address
	Corrections []Correction

	// Codes indicating match quality
	Matches []Match

	// Warning messages
	Warnings []string

	// Additional information about the address
	AdditionalInfo *AdditionalInfo
}

// Correction indicates how to improve the address input
type Correction struct {
	Code string
	Text string
}

// Match indicates if an address is an exact match
type Match struct {
	Code string
	Text string
}

// AdditionalInfo contains additional address information
type AdditionalInfo struct {
	// Delivery point code
	DeliveryPoint string

	// Carrier route code
	CarrierRoute string

	// DPV (Delivery Point Validation) confirmation indicator
	DPVConfirmation string

	// DPVCMRA (Commercial Mail Receiving Agency) indicator
	DPVCMRA string

	// Business flag
	Business string

	// Central delivery flag
	CentralDeliveryPoint string

	// Vacant flag
	Vacant string
}

// Error represents a USPS API error
type Error struct {
	Status string
	Code   string
	Title  string
	Detail string
	Source *ErrorSource
}

// ErrorSource identifies the source of an error
type ErrorSource struct {
	Parameter string
	Example   string
}

func (e *Error) Error() string {
	if e.Detail != "" {
		return e.Detail
	}
	if e.Title != "" {
		return e.Title
	}
	return "USPS API error"
}
