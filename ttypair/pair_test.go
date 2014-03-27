package ttypair

import (
	"code.google.com/p/goplan9/plan9/acme"
	"github.com/rjkroege/wikitools/testhelpers"
	"testing"
)

func Test_Israw(t *testing.T) {
	tp := New()
	testhelpers.AssertBool(t, false, tp.Israw())

	tp.Setcook(false)
	testhelpers.AssertBool(t, true, tp.Israw())
	tp.Setcook(true)
	testhelpers.AssertBool(t, false, tp.Israw())
}

// TODO(rjkroege): Make error testing more robust.
type mockttyfd struct {
	writes [][]byte
}

func (mt *mockttyfd) UnbufferedWrite(b []byte) error {
	mt.writes = append(mt.writes, b)
	return nil
}

func Test_addtype(t *testing.T) {
	tp := New()

	tp.addtype([]byte("hello"), 0, false)
	testhelpers.AssertString(t, "hello", tp.String())

	// addtype is doing the wrong thing...
	tp.addtype([]byte{3}, len("hello"), false)
	testhelpers.AssertString(t, "hello", tp.String())

	// addtype is doing the wrong thing...
	tp.addtype([]byte{3}, len("hello_"), true)
	testhelpers.AssertString(t, "", tp.String())

}

func Test_Sendtype(t *testing.T) {
	tp := New()
	mock := &mockttyfd{make([][]byte, 0, 10)}
	tp.fd = mock

	tp.addtype([]byte("hello\nbye"), 0, false)
	tp.Sendtype()

	testhelpers.AssertInt(t, 1, len(mock.writes))
	testhelpers.AssertString(t, "hello\n", string(mock.writes[0]))
	testhelpers.AssertString(t, "bye", string(tp.Typing))
}

func Test_SendtypeOnechar(t *testing.T) {
	tp := New()
	mock := &mockttyfd{make([][]byte, 0, 10)}
	tp.fd = mock

	tp.addtype([]byte("h"), 0, true)
	tp.Sendtype()

	testhelpers.AssertInt(t, 0, len(mock.writes))
	testhelpers.AssertString(t, "h", string(tp.Typing))
}

func Test_SendtypeMultiblock(t *testing.T) {
	tp := New()
	mock := &mockttyfd{make([][]byte, 0, 10)}
	tp.fd = mock

	tp.addtype([]byte("hello\nworld\nbye"), 0, true)
	tp.Sendtype()

	testhelpers.AssertInt(t, 2, len(mock.writes))
	testhelpers.AssertString(t, "hello\n", string(mock.writes[0]))
	testhelpers.AssertString(t, "world\n", string(mock.writes[1]))
	testhelpers.AssertString(t, "bye", string(tp.Typing))
}

func Test_Type(t *testing.T) {
	tp := New()
	mock := &mockttyfd{make([][]byte, 0, 10)}
	tp.fd = mock

	e := &acme.Event{Nr: len("hello"), Text: []byte("hello")}
	tp.Type(e)

	testhelpers.AssertString(t, "hello", string(tp.Typing))
}

func Test_TypeCook(t *testing.T) {
	tp := New()
	mock := &mockttyfd{make([][]byte, 0, 10)}
	tp.fd = mock

	s := "hello\n"
	e := &acme.Event{Nr: len(s), Text: []byte(s)}
	tp.Type(e)

	testhelpers.AssertInt(t, 1, len(mock.writes))
	testhelpers.AssertString(t, "hello\n", string(mock.writes[0]))
	testhelpers.AssertString(t, "", string(tp.Typing))
	testhelpers.AssertBool(t, true, tp.cook)
}
