package httpapi

import "testing"

func TestBulkExtensionRequestNormalizesDuration(t *testing.T) {
	t.Parallel()
	minutes, days := 17, 2
	tests := []struct {
		name    string
		request bulkExtensionRequest
		want    int
		wantErr bool
	}{
		{name: "minutes", request: bulkExtensionRequest{DurationMinutes: &minutes}, want: 17},
		{name: "legacy days", request: bulkExtensionRequest{Days: &days}, want: 2880},
		{name: "missing", request: bulkExtensionRequest{}, wantErr: true},
		{name: "ambiguous", request: bulkExtensionRequest{DurationMinutes: &minutes, Days: &days}, wantErr: true},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			result, err := test.request.domain()
			if (err != nil) != test.wantErr {
				t.Fatalf("domain() error = %v, wantErr %v", err, test.wantErr)
			}
			if err == nil && result.DurationMinutes != test.want {
				t.Fatalf("duration minutes = %d, want %d", result.DurationMinutes, test.want)
			}
		})
	}
}
