package envarfig

type settings struct {
	AutoLoadEnv bool
	CacheConfig bool
	EnvFiles    []string
}

type option func(*settings)

func loadSettings(opts ...option) *settings {
	setting := &settings{
		AutoLoadEnv: true,
		EnvFiles:    nil,
		CacheConfig: true,
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

// WithCacheConfig sets the cache config option
func WithCacheConfig(CacheConfig bool) option {
	return func(s *settings) {
		s.CacheConfig = CacheConfig
	}
}
