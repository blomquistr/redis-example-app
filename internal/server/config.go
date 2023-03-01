package server

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
	"k8s.io/klog"
)

var (
	configPrefix string = "redistester"
)

type IConfig interface {
	getKeyFile() string
	setKeyFile(keyFile string)
	getCertFile() string
	setCertFile(certFile string)
	getPort() int
	setPort(port int)
	getRedisAddress() string
	setRedisAddress(address string)
	getRedisPassword() string
	setRedisPassword(password string)
	getRedisDB() int
	setRedisDB(db int)
	getRedisPort() int
	setRedisPort(port int)
	getDefaultTTL() int
	setDefaultTTL(ttl int)
	getMaxBodySize() int
	setMaxBodySize(size int)
}

func (c *Config) getCertFile() string {
	return c.CertFile
}

func (c *Config) setCertFile(certFile string) {
	c.CertFile = certFile
}

func (c *Config) getKeyFile() string {
	return c.KeyFile
}

func (c *Config) setKeyFile(keyFile string) {
	c.KeyFile = keyFile
}

func (c *Config) getPort() int {
	return c.Port
}

func (c *Config) setPort(port int) {
	c.Port = port
}

func (c *Config) getRedisAddress() string {
	return c.RedisAddress
}

func (c *Config) setRedisAddress(address string) {
	c.RedisAddress = address
}

func (c *Config) getRedisPassword() string {
	return c.RedisPassword
}

func (c *Config) setRedisPassword(password string) {
	c.RedisPassword = password
}

func (c *Config) getRedisDB() int {
	return c.RedisDB
}

func (c *Config) setRedisDB(db int) {
	c.RedisDB = db
}

func (c *Config) getRedisPort() int {
	return c.RedisPort
}

func (c *Config) setRedisPort(port int) {
	c.RedisPort = port
}

func (c *Config) getDefaultTTL() int {
	return c.DefaultTTL
}

func (c *Config) setDefaultTTL(ttl int) {
	c.DefaultTTL = ttl
}

func (c *Config) getMaxBodySize() int {
	return c.MaxBodySize
}

func (c *Config) setMaxBodySize(size int) {
	c.MaxBodySize = size
}

type Config struct {
	CertFile      string
	KeyFile       string
	Port          int
	RedisAddress  string
	RedisPort     int
	RedisPassword string
	RedisDB       int
	DefaultTTL    int
	MaxBodySize   int
}

func setConfigDefaults() {
	viper.SetDefault("server.cert-file", "")
	viper.SetDefault("server.key-file", "")
	viper.SetDefault("server.port", 5678)
	viper.SetDefault("server.redis-address", "localhost")
	viper.SetDefault("server.redis-port", 6379)
	viper.SetDefault("server.redis-password", "")
	viper.SetDefault("server.redis-db", 0)
	viper.SetDefault("server.default-ttl", 300)
	viper.SetDefault("server.max-body-size", 1048576)
}

func bindConfigEnvironment() {
	viper.SetEnvPrefix(configPrefix)
	viper.BindEnv("server.cert-file", fmt.Sprintf("%s_SERVER_CERT_FILE", strings.ToUpper(configPrefix)))
	viper.BindEnv("server.key-file", fmt.Sprintf("%s_SERVER_KEY_FILE", strings.ToUpper(configPrefix)))
	viper.BindEnv("server.port", fmt.Sprintf("%s_SERVER_PORT", strings.ToUpper(configPrefix)))
	viper.BindEnv("server.redis-address", fmt.Sprintf("%s_SERVER_REDIS_ADDRESS", strings.ToUpper(configPrefix)))
	viper.BindEnv("server.redis-port", fmt.Sprintf("%s_SERVER_REDIS_PORT", strings.ToUpper(configPrefix)))
	viper.BindEnv("server.redis-password", fmt.Sprintf("%s_SERVER_REDIS_PASSWORD", strings.ToUpper(configPrefix)))
	viper.BindEnv("server.redis-db", fmt.Sprintf("%s_SERVER_REDIS_DB", strings.ToUpper(configPrefix)))
	viper.BindEnv("server.default-ttl", fmt.Sprintf("%s_SERVER_DEFAULT_TTL", strings.ToUpper(configPrefix)))
	viper.BindEnv("server.max-body-size", fmt.Sprintf("%s_SERVER_MAX_BODY_SIZE", strings.ToUpper(configPrefix)))
}

func configureConfigFile() {
	viper.SetConfigName("config")             // the name of the confg file
	viper.SetConfigType("yaml")               // the file type of the config file - YAML for our example app
	viper.AddConfigPath("$HOME/.redistester") // where I'd put config on a local machine
	viper.AddConfigPath("/etc/redistester")   // a conventionally accepted place to put an app in a container
	viper.AddConfigPath("/usr/src/app")       // conventionally where Datasite puts its apps in containers
	viper.AddConfigPath(".")                  // always search the app directory, because we should have a testing space
}

func newConfig() IConfig {
	// use Viper to set config defaults
	setConfigDefaults()
	bindConfigEnvironment()
	configureConfigFile()

	// we want to prefer environment variables where they
	// appear; prod should use the environment and so should
	// passwords even in dev.
	viper.AutomaticEnv()

	// use Viper to read a config file and process the results
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			klog.Info("No config file provided, proceeding to OS environment")
		} else {
			klog.Fatal(err)
		}
	}

	return &Config{
		CertFile:      viper.GetString("server.cert-file"),
		KeyFile:       viper.GetString("server.key-file"),
		Port:          viper.GetInt("server.port"),
		RedisAddress:  viper.GetString("server.redis-address"),
		RedisPort:     viper.GetInt("server.redis-port"),
		RedisPassword: viper.GetString("server.redis-password"),
		RedisDB:       viper.GetInt("server.redis-db"),
		DefaultTTL:    viper.GetInt("server.default-ttl"),
		MaxBodySize:   viper.GetInt("server.max-body-size"),
	}
}
