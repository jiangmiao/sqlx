package sqlx

import (
	"strings"

	"github.com/lib/pq"
)

func Underscore(v string) string {
	var n = len(v)
	if n == 0 {
		return ""
	}
	us := make([]byte, n*2)
	us[0] = v[0]
	k := 1
	j := 1
	for j < n {
		if v[j] >= 'A' && v[j] <= 'Z' {
			us[k] = '_'
			k++
		}
		us[k] = v[j]
		k++
		j++
	}
	return strings.ToLower(string(us[:k]))
}

func Quote(v string) string {
	return pq.QuoteIdentifier(v)
}
