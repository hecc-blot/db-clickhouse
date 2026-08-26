package config

// ClickhouseConfig ClickHouse 连接配置
type ClickhouseConfig struct {
	Ip              string `mapstructure:"ip"`
	Port            int    `mapstructure:"port"`
	Username        string `mapstructure:"username"`
	Password        string `mapstructure:"password"`
	Database        string `mapstructure:"database"`
	ConnectTimeout  int    `mapstructure:"connect_timeout"`
	MaxIdleConn     int    `mapstructure:"max_idle_conn"`
	MaxOpenConn     int    `mapstructure:"max_open_conn"`
	ConnMaxLifetime int    `mapstructure:"conn_max_lifetime"`
	SlowThreshold   int    `mapstructure:"slow_threshold"`
}
