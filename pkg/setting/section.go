package setting

type Config struct {
	Server       ServerSetting       `mapstructure:"server"`
	Logger       LoggerSetting       `mapstructure:"logger"`
	Postgresql   PostgresqlSetting   `mapstructure:"postgresql"`
	Mail         MailSetting         `mapstructure:"mail"`
	Auth         AuthSetting         `mapstructure:"auth"`
	R2           R2Setting           `mapstructure:"r2"`
	Notification NotificationSetting `mapstructure:"notification"`
	SES          SESSetting          `mapstructure:"ses"`
}

type AuthSetting struct {
	JwtSecret           string `mapstructure:"jwt_secret"`
	JwtExpiresIn        string `mapstructure:"jwt_expires_in"`         // Duration string like "5m", "1h", "24h"
	JwtRefreshExpiresIn string `mapstructure:"jwt_refresh_expires_in"` // Duration string like "24h", "7d"
}

type R2Setting struct {
	Bucket          string `mapstructure:"bucket"`
	Token           string `mapstructure:"token"`
	AccessKeyId     string `mapstructure:"access_key_id"`
	SecretAccessKey string `mapstructure:"secret_access_key"`
	Url             string `mapstructure:"url"`
	PublicUrl       string `mapstructure:"public"` // Public URL for the bucket (e.g., https://pub-xxx.r2.dev)
}

type ServerSetting struct {
	Port int    `mapstructure:"port"`
	Mode string `mapstructure:"mode"`
}

type LoggerSetting struct {
	LogLevel    string `mapstructure:"log_level"`
	LogFileName string `mapstructure:"log_file_name"`
	MaxSize     int    `mapstructure:"max_size"`
	MaxBackups  int    `mapstructure:"max_backups"`
	MaxAge      int    `mapstructure:"max_age"`
	Compress    bool   `mapstructure:"compress"`
}

type PostgresqlSetting struct {
	Host        string `mapstructure:"host"`
	Port        int    `mapstructure:"port"`
	Username    string `mapstructure:"username"`
	Password    string `mapstructure:"password"`
	DbName      string `mapstructure:"dbname"`
	SslMode     string `mapstructure:"sslmode"`
	TimeZone    string `mapstructure:"timezone"`
	MaxConn     int    `mapstructure:"max_conn"`
	IdleTimeOut int    `mapstructure:"idle_timeout"`
	MaxLifeTime int    `mapstructure:"max_lifetime"`
}

type MailSetting struct {
	ServiceURL      string `mapstructure:"service_url"`
	Username        string `mapstructure:"username"`
	Password        string `mapstructure:"password"`
	TrackingBaseURL string `mapstructure:"tracking_base_url"` // Base URL for tracking pixel (e.g., https://your-domain.com)
}

// NotificationSetting holds configuration for the notification module
type NotificationSetting struct {
	TrackingBaseURL string      `mapstructure:"tracking_base_url"` // Base URL for tracking pixel
	Push            PushSetting `mapstructure:"push"`
	SMS             SMSSetting  `mapstructure:"sms"`
}

// PushSetting holds Firebase Cloud Messaging configuration
type PushSetting struct {
	Enabled         bool   `mapstructure:"enabled"`
	CredentialsFile string `mapstructure:"credentials_file"` // Path to Firebase service account JSON
}

// SMSSetting holds Twilio SMS configuration
type SMSSetting struct {
	Enabled    bool   `mapstructure:"enabled"`
	AccountSID string `mapstructure:"account_sid"`
	AuthToken  string `mapstructure:"auth_token"`
	FromNumber string `mapstructure:"from_number"`
}

type SESSetting struct {
	Region          string `mapstructure:"region"`
	AccessKeyId     string `mapstructure:"access_key_id"`
	SecretAccessKey string `mapstructure:"secret_access_key"`
	FromEmail       string `mapstructure:"from_email"`
	InviteBaseURL   string `mapstructure:"invite_base_url"`
}
