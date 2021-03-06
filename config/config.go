package config

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
)

var Config Conf

type Conf struct {
	Debug  bool
	AppKey string
	Db     Db
	Lvdb   string
	Sms    Sms
	Wechat Wechat
}

type Db struct {
	Url string
}

type Sms struct {
	Key          string
	Secret       string
	Sign         string
	RegTplCode   string
	LoginTplCode string
}

type Wechat struct {
	Id     string
	Secret string
}

func init() {

	data, err := ioutil.ReadFile("config.yml")
	if err != nil {
		log.Panicln(err.Error())
	}

	err = yaml.Unmarshal(data, &Config)
	if err != nil {
		log.Panicln(err.Error())
	}
}
