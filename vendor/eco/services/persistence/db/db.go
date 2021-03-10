package db

import (
	"time"
	"bytes"
	"encoding/json"
)

/**
 * User: Santiago Vidal
 * Date: 16/05/17
 * Time: 15:44
 */

//Init inicializa los motores de base de datos
func Init() error {

	if err := initRedis(); err != nil {
		return err
	}
	initMongo()
	return nil
}

func initMongo() {

	//	host, _ := config.EcoCfg.String("database.mongo.host")
	//	port, _ := config.EcoCfg.String("database.mongo.port")

}

type RedisAware interface {
	RedisFillFields(ent interface{}, data *[]interface{})
}

type EcoEntity struct {}
func (ai EcoEntity) RedisFillFields(ent interface{}, data *[]interface{}) {

	buf := new(bytes.Buffer)
	json.NewEncoder(buf).Encode(ent)

	*data = append(*data, "entity", buf)
}

type AuditEntity interface {
	AuditInfo() (*AuditData, string)
}
type AuditData struct {
	EcoEntity
	CreatedBy  string    `json:"createdBy,omitempty"`
	CreatedAt  time.Time `json:"createdAt,omitempty"`
	ModifiedBy string    `json:"modifiedBy,omitempty"`
	ModifiedAt time.Time `json:"modifiedAt,omitempty"`

	user string
}

func (ai *AuditData) AuditInfo() (*AuditData, string) {
	return ai, ai.user
}
func (ai *AuditData) Audit(email string) {
	ai.user = email
}
func (ai AuditData) RedisFillFields(ent interface{}, data *[]interface{}) {
	*data = append(*data, "created_at", ai.CreatedAt)
	ai.EcoEntity.RedisFillFields(ent, data)
}
