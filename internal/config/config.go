package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"github.com/spf13/viper"
)

type Config struct {
	Server ServerConfig `mapstructure:"server"`
	DB     DBConfig     `mapstructure:"db"`
	Logger LoggerConfig `mapstructure:"logger"`
	Cart   CartConfig   `mapstructure:"cart"`
}

type ServerConfig struct {
	Port              string        `mapstructure:"port"`
	ReadTimeout       time.Duration `mapstructure:"read_timeout"`
	WriteTimeout      time.Duration `mapstructure:"write_timeout"`
	CtxDefaultTimeout time.Duration `mapstructure:"ctx_default_timeout"`
}

type DBConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Name     string `mapstructure:"name"`
	SSLMode  string `mapstructure:"sslmode"`
	DSN      string `mapstructure:"dsn"`
}

type LoggerConfig struct {
	Level string `mapstructure:"level"`
}

type CartConfig struct {
	ItemLimit            int             `mapstructure:"item_limit"`
	DiscountTotalPrice   decimal.Decimal `mapstructure:"-"`
	DiscountItemAmount   int             `mapstructure:"discount_item_amount"`
	DiscountPercentBig   float64         `mapstructure:"discount_percent_big"`
	DiscountPercentSmall float64         `mapstructure:"discount_percent_small"`
}

func Load(path string) (Config, error) {
	v := viper.New()

	v.AddConfigPath(path)
	v.SetConfigName("config")
	v.SetConfigType("yaml")

	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		return Config{}, fmt.Errorf("cfg Load path %q: %w", path, err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return Config{}, fmt.Errorf("cfg Load: %w", err)
	}

	cfg.Cart.DiscountTotalPrice = decimal.NewFromFloat(v.GetFloat64("cart.discount_total_price"))

	if cfg.DB.DSN == "" {
		cfg.DB.DSN = cfg.DB.BuildDSN()
	}

	return cfg, nil
}

func (db DBConfig) BuildDSN() string {
	return fmt.Sprintf(
		"host='%s' port='%d' user='%s' password='%s' dbname='%s' sslmode='%s'",
		db.Host, db.Port, db.User, db.Password, db.Name, db.SSLMode,
	)
}
