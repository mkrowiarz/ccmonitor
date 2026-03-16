package claude

type processInfo struct {
	PID        int
	CPUPercent float64
	MemPercent float64
	Elapsed    string  // raw elapsed string from ps
	Comm       string  // executable basename
}
