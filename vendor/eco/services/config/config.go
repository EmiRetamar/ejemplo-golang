package config

import (
	"github.com/olebedev/config"
	"github.com/ian-kent/go-log/log"
)

/**
 * User: Santiago Vidal
 * Date: 16/05/17
 * Time: 09:24
 */

var EcoCfg *config.Config
 
func LoadConfig() (err error) {

	log.Info("Parsing configuration...")
	cfg, err := load()
	if err != nil {
		return err
	}
	EcoCfg = cfg
	return nil
}

func Info() (ret string) {

	//hay ciertos valores que no quiero que se impriman en el log
	c, _ := EcoCfg.Copy("")
	c.Set("root.clientId", nil)
	c.Set("root.secretKey", nil)
	c.Set("passwordsKey", nil)

	ret, _ = config.RenderYaml(c)
	return ret
}

func load() (*config.Config, error) {

	cfg, err := config.ParseYamlFile("server.yml")
	if err != nil {
		return nil, err
	}

	/*
	reemplazo por configuraciones especificas en variables de enviroment
	ejemplo:

	   para el config de server.yaml
	   'database.redis.host'

	   puede ser reemplazado por la variable:
	   ECO_DATABASE_REDIS_HOST=192.168.100.1

	*/
	cfg = cfg.EnvPrefix("ECO")

	//reemplazo por configuraciones especificas definidas en parametros
	cfg = cfg.Flag()

	return cfg, nil
}