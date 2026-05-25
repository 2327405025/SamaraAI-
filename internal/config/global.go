package config

// global holds the loaded application config for packages that initialize at startup.
var global *Config

func InitGlobal(c *Config) {
	global = c
}

func Get() *Config {
	return global
}

func IsDevMode() bool {
	if global == nil {
		return false
	}
	return global.Mode == "dev"
}
