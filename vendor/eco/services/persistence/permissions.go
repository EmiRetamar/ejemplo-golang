package persistence

import (
	"eco/services/persistence/db"
	"github.com/pborman/uuid"
)

/**
 * User: Santiago Vidal
 * Date: 20/06/17
 * Time: 22:19
 */

//var Permissions

type Permission struct {
	Id   string            `json:"id"`
	Name string            `json:"name"`
	Urls []permissionUrl   `json:"urls,omitempty"`
	db.AuditData
}

type permissionUrl struct {
	Path   string        `json:"path"`
	Method string        `json:"method"`
}

func (p Permission) RedisFillFields(ent interface{}, data *[]interface{}) {
	*data = append(*data, "name", p.Name)
	p.AuditData.RedisFillFields(ent, data)
}

func(p *Permission) Create() {

	s := db.Redis.Session()
	defer s.Close()

	p.Id = uuid.New()
	s.Store("eco:permission:" + p.Name + ":" + p.Id, p)
}

func(p *Permission) Get(id string) {

}
