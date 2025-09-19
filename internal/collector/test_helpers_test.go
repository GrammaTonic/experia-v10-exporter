package collector

// test helpers shared by collector tests
type simpleErr struct{ s string }

func (e *simpleErr) Error() string { return e.s }
