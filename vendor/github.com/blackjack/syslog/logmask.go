package syslog

// LogMask is a bit string with one bit corresponding to each of the possible
// message priorities. If the bit is on, syslog handles messages of that priority normally.
// If it is off, syslog discards messages of that priority
// Use LOG_MASK and LOG_UPTO to construct an appropriate mask value
type LogMask int

// Sets this logmask for the calling process, and returns the previous mask.
// If the mask argument is 0, the current logmask is not modified.
// Example:
// syslog.SetLogMask( syslog.LOG_MASK(LOG_EMERG) | syslog.LOG_MASK(LOG_ERROR) )
func SetLogMask(p LogMask) LogMask {
	mask := setlogmask(int(p))
	return LogMask(mask)
}

//Mask for one priority
func LOG_MASK(p Priority) LogMask {
	mask := (1 << uint(p))
	return LogMask(mask)
}

// Generates a mask with the bits on for a certain priority and all priorities above it
// The unfortunate naming is due to the fact that internally,
// higher numbers are used for lower message priorities.
func LOG_UPTO(p Priority) LogMask {
	mask := (1 << (uint(p) + 1)) - 1
	return LogMask(mask)
}
