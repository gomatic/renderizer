package issues

import (
	"fmt"
	"testing"

	blackbox "github.com/gomatic/renderizer/test"
)

var allTests = map[int][]blackbox.TestCase{
	9: {
		{
			Template: `{{.Technical_user.Nginx}} == foobar`,
			Expects:  `foobar == foobar`,
		},
		{
			Template: `{{.A}} == map[B:[1 2 3]] {{.A.B}} == [1 2 3]`,
			Expects:  `map[B:[1 2 3]] == map[B:[1 2 3]] [1 2 3] == [1 2 3]`,
		},
		{
			Template: `{{.A}} == map[B:[A B C]] {{.A.B}} == [A B C]`,
			Expects:  `map[B:[A B C]] == map[B:[A B C]] [A B C] == [A B C]`,
		},
		{
			Template: `{{.A}} == map[B:[1 A 3]] {{.A.B}} == [1 A 3]`,
			Expects:  `map[B:[1 A 3]] == map[B:[1 A 3]] [1 A 3] == [1 A 3]`,
		},
		{
			Template: `{{.A}} == map[B:[A 2 3]] {{.A.B}} == [A 2 3]`,
			Expects:  `map[B:[A 2 3]] == map[B:[A 2 3]] [A 2 3] == [A 2 3]`,
		},
		{
			Template: `{{.A}} == map[B:map[C:kills(a,b)]] {{.A.B}} == map[C:kills(a,b)] {{.A.B.C}} == kills(a,b)`,
			Expects:  `map[B:map[C:kills(a,b)]] == map[B:map[C:kills(a,b)]] map[C:kills(a,b)] == map[C:kills(a,b)] kills(a,b) == kills(a,b)`,
		},
	},
	13: {
		{
			Template: `{{ range .OSLIST }} echo "{{ .UBUNTU }} aka {{ .OSID }}" {{ end }}`,
			Expects:  `echo "16.04 aka ubu1604" echo "18.04 aka ubu1804"`,
			Config: map[string]interface{}{
				"OSLIST": []map[string]interface{}{
					{
						"UBUNTU": 16.04,
						"OSID":   "ubu1604",
					},
					{
						"UBUNTU": 18.04,
						"OSID":   "ubu1804",
					},
				},
			},
		},
	},
}

//
func TestIssues(t *testing.T) {
	for issueNumber, issues := range allTests {
		for i, test := range issues {
			t.Run(fmt.Sprintf("%04d/%02d", issueNumber, i+1), func(t *testing.T) {
				t.Logf("%04d %+v", issueNumber, test.Template)
			})
		}
	}
}
