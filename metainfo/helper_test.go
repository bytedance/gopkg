package metainfo_test

type tb interface {
	Fatal(args ...interface{})
	Fatalf(format string, args ...interface{})
	Helper()
}

func assert(t tb, cond bool, val ...interface{}) {
	t.Helper()
	if !cond {
		if len(val) > 0 {
			val = append([]interface{}{"assertion failed:"}, val...)
			t.Fatal(val...)
		} else {
			t.Fatal("assertion failed")
		}
	}
}
