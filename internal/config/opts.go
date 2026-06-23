package config

type Options struct {
	URL          string
	Background   bool
	Output       string
	Directory    string
	RateLimit    string
	InputFile    string
	Mirror       bool
	Reject       []string
	Exclude      []string
	ConvertLinks bool
	Timeout      int
}
