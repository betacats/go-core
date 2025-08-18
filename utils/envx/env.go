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

var (
	envMap = map[string]string{
		"dev":  EnvDev,
		"test": EnvTest,
		"uat":  EnvUat,
		"prod": EnvProd,
	}

	// 存储自定义的环境值，优先于系统环境变量
	customOSEnv string
)

// SetOSEnv 设置自定义环境值，会覆盖系统环境变量中的ENV值
func SetOSEnv(env string) {
	customOSEnv = env
}

// SetEnvMap 自定义环境变量映射表，用于扩展或修改环境变量映射关系
func SetEnvMap(m map[string]string) {
	if m != nil {
		envMap = m
	}
}

// ENV
func ENV() string {
	env := os.Getenv("ENV")
	// 优先使用自定义设置的环境值
	if customOSEnv != "" {
		env = os.Getenv(customOSEnv)
	}

	if result, ok := envMap[env]; ok {
		return result
	}
	return EnvDev
}

// Get 获取某个环境变量的值
func Get(str string) string {
	return os.Getenv(str)
}
