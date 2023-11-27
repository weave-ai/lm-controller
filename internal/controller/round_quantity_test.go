package controller

import "testing"

// TestRoundQuantity tests the roundGigaQuantity function with various inputs.
func TestRoundQuantity(t *testing.T) {
	testCases := []struct {
		name         string
		input        string
		expected     string
		safetyFactor float64
	}{
		{"Round 5.24Gi", "5.24Gi", "6Gi", 1.00},
		{"Round 3.12Mi", "3.12Mi", "1Gi", 1.00},
		{"Round 0.99Mi", "0.99Mi", "1Gi", 1.00},
		{"Round 0Gi", "0Gi", "0Gi", 1.00},
		// Add more test cases as needed
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := roundGigaQuantity(tc.input, tc.safetyFactor)
			if err != nil {
				t.Errorf("roundGigaQuantity(%s) returned an error: %v", tc.input, err)
			}
			if got != tc.expected {
				t.Errorf("roundGigaQuantity(%s) = %s; want %s", tc.input, got, tc.expected)
			}
		})
	}
}
