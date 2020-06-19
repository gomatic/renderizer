package blackbox

type TestCase struct {
	Template string
	Expects  string
	Config   map[string]interface{}
}
