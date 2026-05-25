package config

import "github.com/zeromicro/go-zero/rest"

type Config struct {
	rest.RestConf

	AppName string `json:",optional"`

	Email EmailConf `json:",optional"`
	Redis RedisConf `json:",optional"`
	Mysql MysqlConf `json:",optional"`
	Jwt   JwtConf   `json:",optional"`
	Rabbitmq RabbitmqConf `json:",optional"`
	Rag   RagConf   `json:",optional"`
	OpenAI OpenAIConf `json:",optional"`
	Image ImageConf `json:",optional"`
	AiHelper AiHelperConf `json:",optional"`
}

type AiHelperConf struct {
	MaxSessionsPerUser int `json:",optional"`
}

type EmailConf struct {
	Authcode string `json:",optional"`
	Email    string `json:",optional"`
}

type RedisConf struct {
	Host     string `json:",optional"`
	Port     int    `json:",optional"`
	Password string `json:",optional"`
	Db       int    `json:",optional"`
}

type MysqlConf struct {
	Host         string `json:",optional"`
	Port         int    `json:",optional"`
	User         string `json:",optional"`
	Password     string `json:",optional"`
	DatabaseName string `json:",optional"`
	Charset      string `json:",optional"`
}

type JwtConf struct {
	ExpireDuration int    `json:",optional"`
	Issuer         string `json:",optional"`
	Subject        string `json:",optional"`
	Key            string `json:",optional"`
}

type RabbitmqConf struct {
	Host     string `json:",optional"`
	Port     int    `json:",optional"`
	Username string `json:",optional"`
	Password string `json:",optional"`
	Vhost    string `json:",optional"`
}

type RagConf struct {
	EmbeddingModel string `json:",optional"`
	ChatModelName  string `json:",optional"`
	DocDir         string `json:",optional"`
	BaseUrl        string `json:",optional"`
	Dimension      int    `json:",optional"`
}

type OpenAIConf struct {
	ApiKey  string `json:",optional"`
	Model   string `json:",optional"`
	BaseUrl string `json:",optional"`
}

type ImageConf struct {
	ModelPath string `json:",optional"`
	LabelPath string `json:",optional"`
	InputH    int    `json:",optional"`
	InputW    int    `json:",optional"`
}

type RedisKeyConf struct {
	CaptchaPrefix   string
	IndexName       string
	IndexNamePrefix string
}

var DefaultRedisKeys = RedisKeyConf{
	CaptchaPrefix:   "captcha:%s",
	IndexName:       "rag_docs:%s:idx",
	IndexNamePrefix: "rag_docs:%s:",
}
