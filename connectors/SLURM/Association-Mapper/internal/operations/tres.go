// cli/internal/client/tres.go
package operations

import (
	"fmt"
	"strconv"
	"strings"
)

// ParseTRES parses a comma-separated TRES spec:
//
//	"cpu=100", "cpu=100,mem=8000", "gres/gpu=2", "cpu=10,gres/gpu=4"
func ParseTRES(s string) ([]TRES, error) {
	if strings.TrimSpace(s) == "" {
		return nil, nil
	}
	parts := strings.Split(s, ",")
	out := make([]TRES, 0, len(parts))
	for _, p := range parts {
		kv := strings.SplitN(p, "=", 2)
		if len(kv) != 2 {
			return nil, fmt.Errorf("malformed TRES entry %q (want key=value)", p)
		}
		n, err := strconv.ParseInt(strings.TrimSpace(kv[1]), 10, 64)
		if err != nil {
			return nil, fmt.Errorf("TRES count %q not an integer", kv[1])
		}
		t := TRES{Count: n}
		key := strings.TrimSpace(kv[0])
		if slash := strings.Index(key, "/"); slash >= 0 {
			t.Type = key[:slash]
			t.Name = key[slash+1:]
		} else {
			t.Type = key
		}
		out = append(out, t)
	}
	return out, nil
}
