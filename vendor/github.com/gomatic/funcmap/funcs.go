package funcmap

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"text/template"
	"time"

	"github.com/Masterminds/sprig"
	"github.com/gomatic/clock"
)

type opt struct {
	maps               []template.FuncMap
	rightmostOverrides bool
	timeFunc           clock.TimeFunction
}

//
type Optional func(*opt)

//
func WithMaps(fs ...template.FuncMap) Optional {
	return func(o *opt) {
		if len(fs) == 0 {
			return
		}
		o.maps = append(o.maps, fs...)
	}
}

//
func WithMap(fs template.FuncMap) Optional {
	return WithMaps(fs)
}

//
func WithRightmostOverrides() Optional {
	return func(o *opt) {
		o.rightmostOverrides = true
	}
}

//
func WithV1Map() Optional {
	return WithMaps(v1Map)
}

//
func WithV2Map() Optional {
	return WithMaps(v1Map, sprig.GenericFuncMap())
}

//
func WithV3Map() Optional {
	return WithMaps(sprig.GenericFuncMap(), v1Map)
}

//
func WithClock(timeFunc clock.TimeFunction) Optional {
	return func(o *opt) {
		if timeFunc == nil {
			return
		}
		o.timeFunc = timeFunc
	}
}

//
func New(options ...Optional) template.FuncMap {
	opts := opt{
		maps: []template.FuncMap{},
	}
	for _, f := range options {
		if f == nil {
			continue
		}
		f(&opts)
	}

	fm := template.FuncMap{}
	for _, fs := range opts.maps {
		for k, f := range fs {
			if f == nil {
				continue
			}
			_, exists := fm[k]
			if opts.rightmostOverrides || !exists {
				fm[k] = f
			}
		}
	}

	if opts.timeFunc != nil {
		fm["now"] = opts.timeFunc
	}

	return fm
}

//
func init() {

	keySequencer := KeySequencer()
	v1Map = template.FuncMap{
		"debug":        Debug,
		"debugging":    debugging,
		"debug_toggle": debugToggle,
		"debugToggle":  debugToggle,
		"pause":        Pause,
		"command_line": CommandLine,
		"commandLine":  CommandLine,
		"ip_math":      IPMath,
		"IPMath":       IPMath,
		"ip4_inc":      IP4Inc,
		"IP4Inc":       IP4Inc,
		"ip4_next":     IP4Next,
		"IP4Next":      IP4Next,
		"ip4_prev":     IP4Prev,
		"IP4Prev":      IP4Prev,
		"ip4_add":      IP4Add,
		"IP4Add":       IP4Add,
		"ip4_join":     IP4Join,
		"IP4Join":      IP4Join,
		"ip6_inc":      IP6Inc,
		"IP6Inc":       IP6Inc,
		"ip6_next":     IP6Next,
		"IP6Next":      IP6Next,
		"ip6_prev":     IP6Prev,
		"IP6Prev":      IP6Prev,
		"ip6_add":      IP6Add,
		"IP6Add":       IP6Add,
		"ip6_join":     IP6Join,
		"IP6Join":      IP6Join,
		"cidr_next":    CIDRNext,
		"CIDRNext":     CIDRNext,
		"ip_ints":      IPInts,
		"IPInts":       IPInts,
		"ip_split":     IPSplit,
		"IPSplit":      IPSplit,
		"to_int":       ToInt,
		"ToInt":        ToInt,
		"dec_to_int":   DecToInt,
		"DecToInt":     DecToInt,
		"hex_to_int":   HexToInt,
		"HexToInt":     HexToInt,
		"from_int":     FromInt,
		"FromInt":      FromInt,
		"next":         Sequencer(),
		"keynext":      keySequencer,
		"keyNext":      keySequencer,
		"inc":          Step,
		"add":          Add,
		"sub":          Sub,
		"mul":          Mul,
		"div":          SafeDiv,
		"div_":         Div,
		"mod":          Mod,
		"rand":         Rand,
		"identifier":   Cleanse(`^[^[:alpha:]_]+|[^[:alnum:]_]`),
		"cleanse":      Cleanse(`[^[:alpha:]]`),
		"cleanser":     Cleanser,
		"environment":  Environment,
		"env":          Environment,
		"now":          time.Now,
		"started":      Starter,
		"iindex":       Index,
		"split":        Split,
		"join":         Join,
		"substr":       Substr,
		"lower":        strings.ToLower,
		"toLower":      strings.ToLower,
		"replace":      strings.Replace,
		"replace_":     ReReplace,
		"title":        strings.Title,
		"initcap":      ReInitcap,
		"trim":         strings.Trim,
		"trim_":        ReTrim,
		"trim_left":    strings.TrimLeft,
		"trimLeft":     strings.TrimLeft,
		"trim_left_":   ReTrimLeft,
		"trimLeft_":    ReTrimLeft,
		"trim_right":   strings.TrimRight,
		"trimRight":    strings.TrimRight,
		"trim_right_":  ReTrimRight,
		"trimRight_":   ReTrimRight,
		"upper":        strings.ToUpper,
		"toUpper":      strings.ToUpper,
		"basename":     Basename,
		"dirname":      filepath.Dir,
		"ext":          filepath.Ext,
	}

	for k, f := range v1Map {
		Map[k] = f
	}
}

var Map = template.FuncMap{}

//
var v1Map template.FuncMap

// To report a consistent time through a single template.
func Starter() func() time.Time {
	started := clock.Now("")
	return func() time.Time { return started() }
}

//
func Debug(any ...interface{}) string {
	s := make([]string, len(any))
	for i, a := range any {
		s[i] = fmt.Sprintf("%[1]T %[1]v", a)
	}
	return Join(" ", s)
}

// toggle debugging
func Debugger() (func() bool, func() bool) {
	lock := sync.RWMutex{}
	_debugging := false
	get := func() bool {
		lock.Lock()
		defer lock.Unlock()
		return _debugging
	}
	toggle := func() bool {
		lock.Lock()
		defer lock.Unlock()
		_debugging = !_debugging
		return _debugging
	}
	return get, toggle
}

// toggle debugging
var debugging, debugToggle = Debugger()

//
func Pause(t int64) time.Time {
	time.Sleep(time.Duration(t) * time.Millisecond)
	return time.Now()
}

//
func ReReplace(n int, old, new, s string) string { return strings.Replace(s, old, new, n) }
func ReInitcap(s string) string                  { return strings.Title(strings.ToLower(s)) }
func ReTrim(cut, s string) string                { return strings.Trim(s, cut) }
func ReTrimLeft(cut, s string) string            { return strings.TrimLeft(s, cut) }
func ReTrimRight(cut, s string) string           { return strings.TrimRight(s, cut) }
func Rand() int64                                { return rand.Int63() }

// simple sequence generation.
func Sequencer() func() int64 {
	i := int64(0)
	return func() int64 {
		return atomic.AddInt64(&i, 1)
	}
}

// key-based sequencing.
func KeySequencer() func(string) int64 {
	lock := sync.RWMutex{}
	is := map[string]*int64{}
	return func(k string) int64 {
		lock.Lock()
		defer lock.Unlock()
		if _, exists := is[k]; !exists {
			i := int64(0)
			is[k] = &i
		}
		return atomic.AddInt64(is[k], 1)
	}
}

//
func Step(a int64, is ...int) int64 {
	if len(is) == 0 {
		is = []int{1}
	}
	for _, i := range is {
		a += int64(i)
	}
	return a
}

// `b` + `a`
func Add(a, b int64) int64 { return b + a }

// `b` - `a`
func Sub(a, b int64) int64 { return b - a }

// `b` * `a`
func Mul(a, b int64) int64 { return b * a }

// `b` modulo `a`
func Mod(a, b int64) int64 { return b % a }

// `b` / `a`
func Div(a, b int64) int64 {
	return b / a
}

// `b` divided by `a`. Returns `0` if `a == 0`.
func SafeDiv(a, b int64) int64 {
	if a == 0 {
		return 0
	}
	return b / a
}

//
func Cleanser(r, s string) string {
	return regexp.MustCompile(r).ReplaceAllString(s, "")
}

//
func Cleanse(r string) func(string) string {
	re := regexp.MustCompile(r)
	return func(s string) string {
		return re.ReplaceAllString(s, "")
	}
}

//
func IntParser(base int) func(s string) (int64, error) {
	return func(s string) (int64, error) {
		return strconv.ParseInt(s, base, 64)
	}
}

//
func Environment(n string) string {
	v, _ := os.LookupEnv(n)
	return v
}

var (
	parseDec = IntParser(10)
	parseHex = IntParser(16)
)

// TODO increment CIDR
func CIDRNext(cidr uint8, lowest, count, inc int8, addr []int64) []int64 {
	return addr
}

//
func IPCalc(bits int32, lowest, count, inc, value int64) int64 {
	if value < lowest {
		value += int64(bits)
	}
	return (lowest + (value-lowest+inc)%count) % int64(bits)
}

// Given a zero-based, left-to-right IP group index, lowest value, count, and increment,
// increment the group, cyclically.
func IPAdd(bits int32, group uint8, lowest, count uint16, inc int16, addr []int64) []int64 {
	if group >= uint8(len(addr)) {
		return addr
	}
	if lowest == 0 && count == 0 {
		addr[group] = (addr[group] + int64(inc)) % int64(bits)
	} else {
		addr[group] = IPCalc(int32(bits), int64(lowest), int64(count), int64(inc), addr[group])
	}
	return addr
}

//
func IP4Inc(group uint8, inc int8, addr string) string {
	return IP4Join(IP4Add(group, 0, 0, inc, IPInts(addr)))
}

//
func IP4Next(group uint8, lowest, count uint8, addr string) string {
	return IP4Join(IP4Add(group, lowest, count, 1, IPInts(addr)))
}

//
func IP4Prev(group uint8, lowest, count uint8, addr string) string {
	return IP4Join(IP4Add(group, lowest, count, -1, IPInts(addr)))
}

// Given a zero-based, left-to-right IP group index, lowest value, count, and increment,
// increment the group, cyclically.
func IP4Add(group uint8, lowest, count uint8, inc int8, addr []int64) []int64 {
	return IPAdd(int32(256), group, uint16(lowest), uint16(count), int16(inc), addr)
}

//
func IP6Inc(group uint8, inc int16, addr string) string {
	return IP6Join(IP6Add(group, 0, 0, inc, IPInts(addr)))
}

//
func IP6Next(group uint8, lowest, count uint16, addr string) string {
	return IP6Join(IP6Add(group, lowest, count, 1, IPInts(addr)))
}

//
func IP6Prev(group uint8, lowest, count uint16, addr string) string {
	return IP6Join(IP6Add(group, lowest, count, -1, IPInts(addr)))
}

// given a group, lowest, count, and increment, increment the group, circling around
func IP6Add(group uint8, lowest, count uint16, inc int16, addr []int64) []int64 {
	return IPAdd(int32(65536), group, lowest, count, inc, addr)
}

//
func Join(sep string, arr []string) (s string) {
	return strings.Join(arr, sep)
}

//
func Substr(start, end int, s string) string {
	l := len(s)
	if l == 0 {
		return s
	}
	start, end = start%l, end%l
	if start < 0 {
		start = l + start
	}
	if end < 0 {
		end = l + end
	}
	if start > end {
		start, end = end, start
	}
	if start > l || start < 0 || end < 0 {
		return s
	} else if end > l {
		end = l
	}
	return s[start:end]
}

//
func Split(sep, s string) []string {
	//
	return strings.Split(s, sep)
}

//
func Index(i int, a interface{}) interface{} {
	if a == nil {
		return nil
	}
	switch a := a.(type) {
	case []string:
		if i < 0 || i >= len(a) {
			return -1
		}
		return a[i]
	case []int64:
		if i < 0 || i >= len(a) {
			return -1
		}
		return a[i]
	case string:
		if i < 0 || i >= len(a) {
			return -1
		}
		return fmt.Sprintf("%c", a[i])
	}
	return a
}

//
func IPSplit(addr string) []string {
	ip_groups := Split(".", addr)
	if len(ip_groups) > 1 {
		return ip_groups
	}
	return Split(":", addr)
}

//
func IP4Join(addr []int64) string {
	return Join(".", FromInt("%d", addr))
}

//
func IP6Join(addr []int64) string {
	return Join(":", FromInt("%04x", addr))
}

//
func IPInts(addr string) []int64 {
	if ip_groups := Split(".", addr); len(ip_groups) > 1 {
		return DecToInt(ip_groups)
	} else {
		return HexToInt(strings.Split(":", addr))
	}
}

//
func DecToInt(arr []string) []int64 {
	return ToInt(10, arr)
}

//
func HexToInt(arr []string) []int64 {
	return ToInt(16, arr)
}

//
func ToInt(base int, arr []string) []int64 {
	is := make([]int64, len(arr))
	parser := IntParser(base)
	for i, m := range arr {
		p, err := parser(m)
		if err != nil {
			continue
		}
		is[i] = p
	}
	return is
}

//
func FromInt(format string, arr []int64) []string {
	ss := make([]string, len(arr))
	for i, m := range arr {
		ss[i] = fmt.Sprintf(format, m)
	}
	return ss
}

// Performs IP math using a simple sequence of operations.
// e.g. _.[+2]._.[+1,%10]
func IPMath(math, addr string) string {
	sep, format, width := ".", "%d", uint(256)
	ip_groups := Split(sep, addr)
	th_groups := Split(sep, math)
	parser := parseDec
	if len(ip_groups) == 1 {
		parser = parseHex
		sep, format, width = ":", "%04x", uint(65536)
		ip_groups = Split(sep, addr)
		th_groups = Split(sep, math)
	}
	if len(ip_groups) != len(th_groups) {
		return addr
	}
	ip_values := make([]int64, len(ip_groups))
	for i, m := range ip_groups {
		p, err := parser(m)
		if err != nil {
			continue
		}
		ip_values[i] = p
	}
	for i, m := range th_groups {
		m := m
		lm := len(m)
		if lm < 3 {
			continue
		}
		switch m {
		case "_":
			continue
		}
		if m[0] != '[' || m[lm-1] != ']' {
			continue
		}
		m = m[1 : lm-1]
		p := ip_values[i]
		for _, a := range strings.Split(m, ",") {
			a := a
			op := a[0]
			switch op {
			case '+', '-', '*', '/', '%':
				a = a[1:]
			default:
			}

			n := int64(0)
			switch a {
			case "R":
				n = rand.Int63n(int64(width))
			default:
				x, err := parser(a)
				if err != nil {
					continue
				}
				n = x
			}

			switch op {
			case '+':
				p += n
			case '-':
				p -= n
			case '*':
				p *= n
			case '/':
				p /= n
			case '%':
				p %= n
			default:
				p = n
			}
			p %= int64(width)
		}
		ip_groups[i] = fmt.Sprintf(format, uint(p)%width)
	}
	return Join(sep, ip_groups)
}

// Reproduce a command line string that reflects a usable command line.
func CommandLine() string {

	quoter := func(e string) string {
		if !strings.Contains(e, " ") {
			return e
		}
		p := strings.SplitN(e, "=", 2)
		if strings.Contains(p[0], " ") {
			p[0] = `"` + strings.Replace(p[0], `"`, `\"`, -1) + `"`
		}
		if len(p) == 1 {
			return p[0]
		}
		return p[0] + `="` + strings.Replace(p[1], `"`, `\"`, -1) + `"`
	}
	each := func(s []string) (o []string) {
		o = make([]string, len(s))
		for i, t := range s {
			o[i] = quoter(t)
		}
		return
	}
	return filepath.Base(os.Args[0]) + " " + strings.Join(each(os.Args[1:]), " ")
}

//
func Basename(path string, extensions ...string) string {
	name := filepath.Base(path)
	for _, ext := range extensions {
		name = strings.TrimSuffix(name, "."+ext)
	}
	return name
}
