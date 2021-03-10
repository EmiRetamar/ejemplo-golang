package persistence

import (
	"eco/services/persistence/db"
)

/**
 * User: Santiago Vidal
 * Date: 11/05/17
 * Time: 14:01
 */

type User struct {
	Email      string   `json:"email"`
	FirstName  string   `json:"firstName"`
	LastName   string   `json:"lastName"`
	Password   string   `json:"password,omitempty"`
	Domains    []string `json:"domains,omitempty"`
	Admin      bool     `json:"admin"`
	Active     bool     `json:"active"`
	LastDomain string   `json:"lastDomain,omitempty"`
	Tags       []string `json:"tags"`
	Organizations   []string `json:"organizations"`
	FromDB     bool     `json:"-"`
	Root       bool     `json:"-"`
	db.AuditData
}

func (u User) RedisFillFields(ent interface{}, data *[]interface{}) {
	*data = append(*data,
		"firstName", u.FirstName,
		"lastName", u.LastName,
		"admin", u.Admin,
		"active", u.Active,
		"tags", u.Tags)
	u.AuditData.RedisFillFields(ent, data)
}

