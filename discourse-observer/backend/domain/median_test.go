package domain

import "testing"

func TestMedian(t *testing.T) {
	tests := []struct {
		name   string
		values []int64
		want   *int64
	}{
		{"empty", nil, nil},
		{"single", []int64{42}, intPtr(42)},
		{"odd count", []int64{1, 3, 7}, intPtr(3)},
		{"even count truncates", []int64{1, 3}, intPtr(2)},
		{"even count truncates 2.5", []int64{1, 4}, intPtr(2)},
		{"unsorted input", []int64{7, 1, 3}, intPtr(3)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Median(tt.values)
			if tt.want == nil {
				if got != nil {
					t.Errorf("got %d, want nil", *got)
				}
				return
			}
			if got == nil {
				t.Fatalf("got nil, want %d", *tt.want)
			}
			if *got != *tt.want {
				t.Errorf("got %d, want %d", *got, *tt.want)
			}
		})
	}
}

func intPtr(v int64) *int64 { return &v }
