package envx

import "os"

const (
	// EnvDev 开发环境
	EnvDev = "dev"
	// EnvTest 测试环境
	EnvTest = "test"
	// EnvTest 灰度环境
	EnvUat = "uat"
	// EnvProd 生产环境
	EnvProd = "prod"
)

var EnvMap = map[string]string{
	"dev":  EnvDev,
	"test": EnvTest,
	"uat":  EnvUat,
	"prod": EnvProd,
}

func ENV() string {
	env := os.Getenv("ENV")
	if result, ok := EnvMap[env]; ok {
		return result
	}
	return EnvDev
}

func Get(str string) string {
	return os.Getenv(str)
}
