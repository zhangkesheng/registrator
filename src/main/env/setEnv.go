package env

import "os"

// set env
func Set_env() {
	Init()
	if envMap != nil {
		for k, v := range envMap {
			if len(k) > 0 && len(v) > 0 && len(os.Getenv(k)) == 0 {
				os.Setenv(k, v)
			}
		}
	}
}
