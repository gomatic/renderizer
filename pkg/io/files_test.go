package io

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

// mock stdin

func TestStdin(t *testing.T) {
	stdin = mocker(t, 3)
	f, err := NewFile("")
	assert.Nil(t, err)
	eofHelper(t, 3, EOFCloser(f))
}

// mock reader

//
var mocked = 0

//
func mocker(t *testing.T, reads int) io.ReadCloser {
	mocked++
	m := &mockReader{
		name:  mocked,
		t:     t,
		reads: reads,
	}
	return EOFCloser(m)
}

//
type mockReader struct {
	name  int
	reads int
	t     *testing.T
}

//
func (m *mockReader) Read(buf []byte) (n int, err error) {
	m.reads--
	if m.reads >= 0 {
		m.t.Logf("mock %d still reading: %d", m.name, m.reads)
		return 2, nil
	}
	m.t.Logf("mock %d returning EOF: %d", m.name, m.reads)
	return 0, io.EOF
}

//
func (m *mockReader) Close() error {
	m.t.Logf("mock %d closing", m.name)
	return nil
}

// helpers

//
func readHelper(t *testing.T, reads int, r io.Reader) {
	for i := reads; i > 0; i-- {
		n, err := r.Read([]byte{})
		assert.Equal(t, 2, n)
		assert.Nil(t, err)
	}
}

//
func eofHelper(t *testing.T, reads int, r io.Reader) {
	readHelper(t, reads, r)
	n, err := r.Read([]byte{})
	assert.Equal(t, 0, n)
	assert.Equal(t, err, io.EOF)
}

// test cases

//
func TestCloser_EOF(t *testing.T) {
	eofHelper(t, 0, mocker(t, 0))
}

//
func TestCloser_1(t *testing.T) {
	eofHelper(t, 1, mocker(t, 1))
}

//
func TestCloser_5(t *testing.T) {
	eofHelper(t, 5, mocker(t, 5))
}

//
func TestMultiCloser_111(t *testing.T) {
	r := io.MultiReader(mocker(t, 1), mocker(t, 1), mocker(t, 1))
	eofHelper(t, 3, r)
}

//
func TestMultiCloser_311(t *testing.T) {
	r := io.MultiReader(mocker(t, 3), mocker(t, 1), mocker(t, 1))
	eofHelper(t, 5, r)
}

//
func TestMultiCloser_131(t *testing.T) {
	r := io.MultiReader(mocker(t, 1), mocker(t, 3), mocker(t, 1))
	eofHelper(t, 5, r)
}

//
func TestMultiCloser_113(t *testing.T) {
	r := io.MultiReader(mocker(t, 1), mocker(t, 1), mocker(t, 3))
	eofHelper(t, 5, r)
}

//
func TestMultiCloser_333(t *testing.T) {
	r := io.MultiReader(mocker(t, 3), mocker(t, 3), mocker(t, 3))
	eofHelper(t, 9, r)
}

//
func TestMultiCloser_Nil11_panic(t *testing.T) {
	assert.Panics(t, func() {
		r := io.MultiReader(nil, mocker(t, 1), mocker(t, 1))
		eofHelper(t, 3, r)
	})
}

//
func TestMultiCloser_1Nil1_panic(t *testing.T) {
	assert.Panics(t, func() {
		r := io.MultiReader(mocker(t, 1), nil, mocker(t, 1))
		eofHelper(t, 3, r)
	})
}

//
func TestMultiCloser_11Nil_panic(t *testing.T) {
	assert.Panics(t, func() {
		r := io.MultiReader(mocker(t, 1), mocker(t, 1), nil)
		eofHelper(t, 3, r)
	})
}

//
func TestMultiCloser_11Nil_1(t *testing.T) {
	r := io.MultiReader(mocker(t, 1), mocker(t, 1), nil)
	readHelper(t, 1, r)
}

//
func TestMultiCloser_11Nil_2(t *testing.T) {
	r := io.MultiReader(mocker(t, 1), mocker(t, 1), nil)
	readHelper(t, 2, r)
}

//
func TestMultiCloser_11Nil_3(t *testing.T) {
	assert.Panics(t, func() {
		r := io.MultiReader(mocker(t, 1), mocker(t, 1), nil)
		readHelper(t, 3, r)
	})
}
