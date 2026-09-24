package rollover

import "testing"

func TestWholeTermEligibilityAndCentRounding(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name                          string
		allocated, used, paid         int64
		minimumBPS                    int
		wantEligible, wantCreditMinor int64
	}{
		{name: "reported 85 percent example", allocated: 53472, used: 7518, paid: 58800, minimumBPS: 8500, wantEligible: 45954, wantCreditMinor: 50533},
		{name: "equal threshold is excluded", allocated: 1000, used: 150, paid: 58800, minimumBPS: 8500},
		{name: "under threshold is excluded", allocated: 1000, used: 151, paid: 58800, minimumBPS: 8500},
		{name: "all unused at zero threshold", allocated: 1000, paid: 58800, wantEligible: 1000, wantCreditMinor: 58800},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			eligible := EligibleUnused(test.allocated, test.used, test.minimumBPS)
			credit := CreditMinor(test.paid, eligible, test.allocated)
			if eligible != test.wantEligible || credit != test.wantCreditMinor {
				t.Fatalf("eligible/credit = %d/%d, want %d/%d", eligible, credit, test.wantEligible, test.wantCreditMinor)
			}
		})
	}
}

func TestWholeTermIncludesPeriodsBelowIndividualThreshold(t *testing.T) {
	t.Parallel()
	allocated, used := int64(2000), int64(280)
	if got := EligibleUnused(allocated, used, 8500); got != 1720 {
		t.Fatalf("eligible = %d, want 1720", got)
	}
}
