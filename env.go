package envarfig

import (
	"github.com/joho/godotenv"
)

var envLoader = godotenv.Load

/*
info: loads the env file

useage: loadEnvFile(true, "path/to/envfile") or loadEnvFile(true, []string{"path/to/envfile1", "path/to/envfile2"})

args:
  - useEnvFile: a boolean value to determine if the env file should be used(uses godotenv)
  - filePath: the file path of the env variables or list of paths
*/
func loadEnvFile(autoLoadEnv bool, filePath []string) error {
	if autoLoadEnv && filePath == nil {
		// if filePath is nil, load the default env file
		// this will load the .env file in the current directory
		return envLoader()
	}
	if autoLoadEnv && filePath != nil {
		return envLoader(filePath...)
	}
	if !autoLoadEnv && filePath != nil {
		return errAutoLoadFalseFilePath
	}
	return nil

}
