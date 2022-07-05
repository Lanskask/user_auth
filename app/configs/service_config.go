package configs

type ServiceConfig struct {
	AuthKey    string `yaml:"auth_key"`
	UserIDKey  string `yaml:"user_id_key"`
	BcryptCost int    `yaml:"bcrypt_cost"`
	Port       string `yaml:"port"`
}

func NewDefaultServiceConf() ServiceConfig {
	return ServiceConfig{
		AuthKey:    "authenticated",
		UserIDKey:  "user_id",
		BcryptCost: 14,
		Port:       ":5000",
	}
}
