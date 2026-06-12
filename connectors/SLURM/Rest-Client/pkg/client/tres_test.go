// cli/internal/client/tres_test.go
package client

import (
	"reflect"
	"testing"
)

func TestParseTRES(t *testing.T) {
	cases := []struct {
		in   string
		want []TRES
	}{
		{"cpu=100", []TRES{{Type: "cpu", Count: 100}}},
		{"cpu=100,mem=8000", []TRES{{Type: "cpu", Count: 100}, {Type: "mem", Count: 8000}}},
		{"gres/gpu=4", []TRES{{Type: "gres", Name: "gpu", Count: 4}}},
		{"cpu=10,gres/gpu=2", []TRES{{Type: "cpu", Count: 10}, {Type: "gres", Name: "gpu", Count: 2}}},
	}
	for _, c := range cases {
		got, err := ParseTRES(c.in)
		if err != nil {
			t.Errorf("ParseTRES(%q): err=%v", c.in, err)
			continue
		}
		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("ParseTRES(%q) = %+v, want %+v", c.in, got, c.want)
		}
	}

	if _, err := ParseTRES("nope"); err == nil {
		t.Error("expected error on malformed input")
	}
}
