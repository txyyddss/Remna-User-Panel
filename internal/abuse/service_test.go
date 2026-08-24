package abuse

import "testing"

func TestReachedLimitQualifiesForAStreak(t *testing.T) {
	for _, test := range []struct {
		name  string
		qps   int
		limit int
		want  bool
	}{
		{name: "below", qps: 9, limit: 10, want: false},
		{name: "equal", qps: 10, limit: 10, want: true},
		{name: "above", qps: 11, limit: 10, want: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			if got := reachesLimit(test.qps, test.limit); got != test.want {
				t.Fatalf("reachesLimit(%d, %d) = %t, want %t", test.qps, test.limit, got, test.want)
			}
		})
	}
}
