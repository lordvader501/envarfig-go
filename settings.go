package envarfig

type settings struct {
	AutoLoadEnv bool
	EnvFiles    []string
}

type option func(*settings)

func loadSettings(opts ...option) *settings {
	setting := &settings{
		AutoLoadEnv: true,
		EnvFiles:    nil,
	}
	for _, opt := range opts {
		opt(setting)
	}
	return setting
}

// WithEnvFiles sets the env file paths
func WithEnvFiles(envFiles ...string) option {
	return func(s *settings) {
		s.EnvFiles = envFiles
	}
}

// WithAutoLoadEnv sets the use env file option
func WithAutoLoadEnv(AutoLoadEnv bool) option {
	return func(s *settings) {
		s.AutoLoadEnv = AutoLoadEnv
	}
}
