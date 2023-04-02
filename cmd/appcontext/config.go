package appcontext

import (
	"encoding/json"
	"strings"
	"time"
)

type serverConfig struct {
	HTTPServer struct {
		Listen      string        `default:":8080" field:"listen" json:"listen" yaml:"listen" toml:"listen" cli:"http_listen" env:"SERVER_HTTP_LISTEN"`
		ReadTimeout time.Duration `default:"120s" field:"read_timeout" json:"read_timeout" yaml:"read_timeout" toml:"read_timeout" env:"SERVER_HTTP_READ_TIMEOUT"`
	} `json:"http_server" yaml:"http_server" toml:"http_server"`
	RedisServer struct {
		Listen      string        `default:":6380" field:"listen" json:"listen" yaml:"listen" toml:"listen" cli:"redis_listen" env:"SERVER_REDIS_LISTEN"`
		ReadTimeout time.Duration `default:"120s" field:"read_timeout" json:"read_timeout" yaml:"read_timeout" toml:"read_timeout" env:"SERVER_REDIS_READ_TIMEOUT"`
	} `json:"redis_server" yaml:"redis_server" toml:"redis_server"`
	Profile struct {
		Listen string `field:"listen" json:"listen" yaml:"listen" toml:"listen" default:":6060" env:"SERVER_PROFILE_LISTEN"`
		Mode   string `field:"mode" json:"mode" yaml:"mode" toml:"mode" default:"net" env:"SERVER_PROFILE_MODE"`
	} `json:"profile" yaml:"profile" toml:"profile"`
}

type cacheConfig struct {
	Connect string        `field:"connect" json:"connect" yaml:"connect" toml:"connect" env:"CACHE_CONNECT" default:"memory"`
	Size    int           `field:"size" json:"size" yaml:"size" toml:"size" env:"CACHE_SIZE" default:"1000"`
	TTL     time.Duration `field:"ttl" json:"ttl" yaml:"ttl" toml:"ttl" env:"CACHE_TTL" default:"60s"`
}

type dataSourceKeyBind struct {
	DBNum            int    `field:"dbnum" json:"dbnum" yaml:"dbnum" toml:"dbnum"`
	TableName        string `field:"table_name" json:"table_name,omitempty" yaml:"table_name" toml:"table_name"`
	Key              string `field:"key" json:"key" yaml:"key" toml:"key"` // Pattern prefix1_{{id}}_suffix, prefix2_{{id}}_{{codename}}
	Readonly         bool   `field:"readonly" json:"readonly" yaml:"readonly" toml:"readonly"`
	WhereExt         string `field:"where_ext" json:"where_ext,omitempty" yaml:"where_ext" toml:"where_ext"`
	GetQuery         string `field:"get_query" json:"get_query,omitempty" yaml:"get_query" toml:"get_query"`
	ListQuery        string `field:"list_query" json:"list_query,omitempty" yaml:"list_query" toml:"list_query"`
	UpsertQuery      string `field:"upsert_query" json:"upsert_query,omitempty" yaml:"upsert_query" toml:"upsert_query"`
	DelQuery         string `field:"del_query" json:"del_query,omitempty" yaml:"del_query" toml:"del_query"`
	ReorganizeNested bool   `field:"reorganize_nested" json:"reorganize_nested,omitempty" yaml:"reorganize_nested" toml:"reorganize_nested"` // Reorganize nested data to flat structure
}

type dataSource struct {
	Connect       string              `field:"connect" json:"connect" yaml:"connect" toml:"connect"`
	NotifyChannel string              `field:"notify_channel" json:"notify_channel,omitempty" yaml:"notify_channel" toml:"notify_channel"`
	Binds         []dataSourceKeyBind `field:"binds" json:"binds" yaml:"binds" toml:"binds"`
}

// ConfigType contains all application options
type ConfigType struct {
	ConfigPath string `cli:"conf" env:"CONFIG_PATH"`

	ServiceName string `field:"service_name" json:"service_name" yaml:"service_name" toml:"service_name" env:"SERVICE_NAME" default:"redify"`
	Hostname    string `field:"hostname" json:"hostname" yaml:"hostname" toml:"hostname" env:"HOSTNAME"`
	Hostcode    string `field:"hostcode" json:"hostcode" yaml:"hostcode" toml:"hostcode" env:"HOSTCODE"`

	LogAddr    string `field:"log_addr" json:"log_addr" yaml:"log_addr" toml:"log_addr" default:"" env:"LOG_ADDR"`
	LogLevel   string `field:"log_level" json:"log_level" yaml:"log_level" toml:"log_level" default:"debug" env:"LOG_LEVEL"`
	LogEncoder string `field:"log_encoder" json:"log_encoder" yaml:"log_encoder" toml:"server" env:"LOG_ENCODER"`

	Server  serverConfig `field:"server" json:"server" yaml:"server" toml:"server"`
	Cache   cacheConfig  `field:"cache" json:"cache" yaml:"cache" toml:"cache"`
	Sources []dataSource `field:"sources" json:"sources" yaml:"sources" toml:"sources"`
}

// String implementation of Stringer interface
func (cfg *ConfigType) String() (res string) {
	if data, err := json.MarshalIndent(cfg, "", "  "); err != nil {
		res = `{"error":"` + err.Error() + `"}`
	} else {
		res = string(data)
	}
	return res
}

// IsDebug mode
func (cfg *ConfigType) IsDebug() bool {
	return strings.EqualFold(cfg.LogLevel, "debug")
}

func (cfg *ConfigType) ConfigFilepath() string {
	return cfg.ConfigPath
}
