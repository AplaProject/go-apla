package smart

import (
	"testing"
)

func TestDateTimeLocation(t *testing.T) {
	type args struct {
		unix         int64
		locationName string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{"Error", args{0, "Location/Bad"}, "", true},
		{"Luxembourg", args{1560938400, "Europe/Luxembourg"}, "2019-06-19 12:00:00", false},
		{"Moscow", args{1560938400, "Europe/Moscow"}, "2019-06-19 13:00:00", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DateTimeLocation(tt.args.unix, tt.args.locationName)
			if (err != nil) != tt.wantErr {
				t.Errorf("DateTimeLocation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("DateTimeLocation() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUnixDateTimeLocation(t *testing.T) {
	type args struct {
		value        string
		locationName string
	}
	tests := []struct {
		name    string
		args    args
		want    int64
		wantErr bool
	}{
		{"BadLocation", args{"", "Location/Bad"}, 0, true},
		{"BadFormat", args{"2019-06-19", "Europe/Luxembourg"}, 0, true},
		{"Luxembourg", args{"2019-06-19 12:00:00", "Europe/Luxembourg"}, 1560938400, false},
		{"Moscow", args{"2019-06-19 12:00:00", "Europe/Moscow"}, 1560934800, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := UnixDateTimeLocation(tt.args.value, tt.args.locationName)
			if (err != nil) != tt.wantErr {
				t.Errorf("UnixDateTimeLocation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("UnixDateTimeLocation() = %v, want %v", got, tt.want)
			}
		})
	}
}
