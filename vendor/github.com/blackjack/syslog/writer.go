package syslog

// An io.Writer() interface
type Writer struct {
	LogPriority Priority
}

func (w *Writer) Write(b []byte) (int, error) {
	Syslog(w.LogPriority, string(b))
	return len(b), nil
}
