package phonenumbers

import (
	"fmt"

	"github.com/nyaruka/phonenumbers"
)

func NormalizePhoneNumber(rawInput string, defaultRegion string) (string, error) {
	num, err := phonenumbers.Parse(rawInput, defaultRegion)
	if err != nil {
		return "", fmt.Errorf("invalid phone number: %v", err)
	}
	if !phonenumbers.IsValidNumber(num) {
		return "", fmt.Errorf("phone number is not valid")
	}
	return phonenumbers.Format(num, phonenumbers.E164), nil
}
