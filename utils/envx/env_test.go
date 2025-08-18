package envx

import (
	"os"
	"testing"
)

func TestGet(t *testing.T) {
	os.Setenv("ENV2", "test1")
	SetOSEnv("ENV2")

	SetEnvMap(map[string]string{"test1": EnvTest})
	e := ENV()
	t.Log(e)
}
