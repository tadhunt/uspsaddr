package uspsaddr

import (
	"github.com/tadhunt/uspsaddr/uspsinternal"
)

// convertResponse converts a USPS API response to our public types
func convertResponse(resp *uspsinternal.AddressResponse) ValidationResult {
	result := ValidationResult{}

	// Convert address
	if resp.Address != nil {
		result.Address = convertAddress(resp.Address, resp.Firm)
	}

	// Get DPV confirmation for generating user messages
	dpvConfirmation := ""
	if resp.AdditionalInfo != nil && resp.AdditionalInfo.DPVConfirmation != nil {
		dpvConfirmation = string(*resp.AdditionalInfo.DPVConfirmation)
	}

	// Check if secondary address is present in response
	hasSecondaryAddress := resp.Address != nil && resp.Address.SecondaryAddress != nil && *resp.Address.SecondaryAddress != ""

	// Convert corrections
	if resp.Corrections != nil {
		result.Corrections = make([]Correction, 0, len(*resp.Corrections))
		for _, c := range *resp.Corrections {
			code := stringValue(c.Code)
			text := stringValue(c.Text)
			// Only add non-empty corrections
			if code != "" || text != "" {
				userMessage := generateUserMessage(code, text, dpvConfirmation, hasSecondaryAddress)
				result.Corrections = append(result.Corrections, Correction{
					Code:        code,
					Text:        text,
					UserMessage: userMessage,
				})
			}
		}
	}

	// Convert matches
	if resp.Matches != nil {
		result.Matches = make([]Match, 0, len(*resp.Matches))
		for _, m := range *resp.Matches {
			code := stringValue(m.Code)
			text := stringValue(m.Text)
			// Only add non-empty matches
			if code != "" || text != "" {
				result.Matches = append(result.Matches, Match{
					Code: code,
					Text: text,
				})
			}
		}
	}

	// Convert warnings
	if resp.Warnings != nil {
		result.Warnings = *resp.Warnings
	}

	// Convert additional info
	if resp.AdditionalInfo != nil {
		result.AdditionalInfo = convertAdditionalInfo(resp.AdditionalInfo)
	}

	return result
}

// convertAddress converts a USPS DomesticAddress to our Address type
func convertAddress(addr *uspsinternal.DomesticAddress, firm *string) Address {
	result := Address{}

	if firm != nil {
		result.Firm = *firm
	}

	if addr.StreetAddress != nil {
		result.StreetAddress = *addr.StreetAddress
	}

	if addr.StreetAddressAbbreviation != nil {
		result.StreetAddressAbbreviation = *addr.StreetAddressAbbreviation
	}

	if addr.SecondaryAddress != nil {
		result.SecondaryAddress = *addr.SecondaryAddress
	}

	if addr.City != nil {
		result.City = *addr.City
	}

	if addr.CityAbbreviation != nil {
		result.CityAbbreviation = *addr.CityAbbreviation
	}

	if addr.State != nil {
		result.State = *addr.State
	}

	if addr.ZIPCode != nil {
		result.ZIPCode = *addr.ZIPCode
	}

	if addr.ZIPPlus4 != nil {
		result.ZIPPlus4 = *addr.ZIPPlus4
	}

	if addr.Urbanization != nil {
		result.Urbanization = *addr.Urbanization
	}

	return result
}

// convertAdditionalInfo converts USPS additional info to our type
func convertAdditionalInfo(info *uspsinternal.AddressAdditionalInfo) *AdditionalInfo {
	result := &AdditionalInfo{}

	if info.DeliveryPoint != nil {
		result.DeliveryPoint = *info.DeliveryPoint
	}

	if info.CarrierRoute != nil {
		result.CarrierRoute = *info.CarrierRoute
	}

	if info.DPVConfirmation != nil {
		result.DPVConfirmation = string(*info.DPVConfirmation)
	}

	if info.DPVCMRA != nil {
		result.DPVCMRA = string(*info.DPVCMRA)
	}

	if info.Business != nil {
		result.Business = string(*info.Business)
	}

	if info.CentralDeliveryPoint != nil {
		result.CentralDeliveryPoint = string(*info.CentralDeliveryPoint)
	}

	if info.Vacant != nil {
		result.Vacant = string(*info.Vacant)
	}

	return result
}

// convertError converts a USPS error response to our Error type
func convertError(errResp *uspsinternal.ErrorMessage) *Error {
	if errResp == nil || errResp.Error == nil {
		return &Error{Title: "Unknown error"}
	}

	e := errResp.Error
	result := &Error{}

	if e.Code != nil {
		result.Code = *e.Code
		result.Status = *e.Code // Use code as status since no separate status field
	}

	if e.Message != nil {
		result.Detail = *e.Message
		result.Title = *e.Message
	}

	// Check if there are detailed errors
	if e.Errors != nil && len(*e.Errors) > 0 {
		// Use first error for more details
		firstErr := (*e.Errors)[0]
		if firstErr.Detail != nil {
			result.Detail = *firstErr.Detail
		}
		if firstErr.Title != nil {
			result.Title = *firstErr.Title
		}
		if firstErr.Source != nil {
			result.Source = &ErrorSource{}
			if firstErr.Source.Parameter != nil {
				result.Source.Parameter = *firstErr.Source.Parameter
			}
			if firstErr.Source.Example != nil {
				result.Source.Example = *firstErr.Source.Example
			}
		}
	}

	return result
}

// stringValue safely dereferences a string pointer
func stringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// generateUserMessage creates a user-friendly message based on correction code and DPV confirmation
func generateUserMessage(code, text, dpvConfirmation string, hasSecondaryAddress bool) string {
	// Handle correction code 32 (more information needed)
	if code == "32" {
		if dpvConfirmation == "D" {
			// D = Missing secondary information
			return text // Use the original USPS message
		} else if dpvConfirmation == "S" && hasSecondaryAddress {
			// S = Secondary information present but not confirmed
			return "USPS does not have enough data to validate the secondary address. Please double check what you entered."
		}
	}

	// For all other cases, use the original USPS text
	return text
}
