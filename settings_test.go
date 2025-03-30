//go:build unit

package envarfig

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSettings(t *testing.T) {
	t.Run("TestSettings", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name                string
			opts                []option
			expectedAutoLoadEnv bool
			expectedEnvFiles    []string
		}{
			{
				name:                "Default settings",
				opts:                nil,
				expectedAutoLoadEnv: true,
				expectedEnvFiles:    nil,
			},
			{
				name:                "WithEnvFiles option",
				opts:                []option{WithEnvFiles("file1.env", "file2.env")},
				expectedAutoLoadEnv: true,
				expectedEnvFiles:    []string{"file1.env", "file2.env"},
			},
			{
				name:                "WithAutoLoadEnv option",
				opts:                []option{WithAutoLoadEnv(false)},
				expectedAutoLoadEnv: false,
				expectedEnvFiles:    nil,
			},
			{
				name:                "Combined options",
				opts:                []option{WithEnvFiles("file1.env"), WithAutoLoadEnv(false)},
				expectedAutoLoadEnv: false,
				expectedEnvFiles:    []string{"file1.env"},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				settings := loadSettings(tt.opts...)
				assert.Equal(t, settings.AutoLoadEnv, tt.expectedAutoLoadEnv, "AutoLoadEnv should match expected value")
				assert.Equal(t, len(settings.EnvFiles), len(tt.expectedEnvFiles), "EnvFiles length should match expected value")
				for i, file := range tt.expectedEnvFiles {
					assert.Equal(t, settings.EnvFiles[i], file, "EnvFiles[%d] should match expected value", i)
				}
			})
		}
	})
}
