package shared

import "testing"

// TrapOsExitPanic replaces OsExit with a function that records the call
// and panics with "exit". restore must be deferred (or called from TearDownTest)
// to put OsExit back.
func TrapOsExitPanic(t testing.TB) (restore func(), exitCalled *bool) {
	t.Helper()
	orig := OsExit
	called := false
	OsExit = func(int) {
		called = true
		panic("exit")
	}
	return func() { OsExit = orig }, &called
}
