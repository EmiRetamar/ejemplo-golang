package persistence

import (
	"eco/services/halt"
	"eco/services/modules"
	"eco/services/persistence/db"
	"encoding/json"
	"fmt"
	"net/http"
)

const ContextKey = "eco:context:hash:%s:%s"             //domain, id
const ContextList = "eco:context:list:%s"               //domain
const ContextListForUser = "eco:context:userlist:%s:%s" //domain, email
const ContextFeatureList = "eco:context:featureList:%s" //ctx ID
//const contextMembers = "eco:context:members:%s" //ctxID

type Context interface {
	ID() string
	Name() string
	IsMember(email string) bool
	HasFeature(featureID string) bool
	ValidateModules() error
}

//contexto default
type contextDefault struct {
	domain string
}

func (cd contextDefault) ID() string {
	return "{eco." + cd.domain + ".default.context}"
}

func (cd contextDefault) Name() string {
	return cd.domain + " default context"
}

func (cd contextDefault) IsMember(email string) bool {
	//	return cd.Domain.IsUserActive(email)
	return true
}

func (cd contextDefault) HasFeature(featureID string) bool {
	return false
}

func (cd contextDefault) ValidateModules() error {
	return nil //el default no valida modulos actualmente
}

func (cd *contextDefault) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
		Domain      string `json:"domain"`
	}{
		ID:     cd.ID(),
		Name:   cd.domain + " default context",
		Domain: cd.domain,
	})
}

func DefaultContext(domain string) Context {
	return &contextDefault{
		domain: domain,
	}
}

func IsDefaultContext(id string, domain string) bool {
	return id == DefaultContext(domain).ID()
}

func GetContext(domain string, id string) (Context, error) {
	defCtx := DefaultContext(domain)
	if len(id) == 0 || id == defCtx.ID() {
		return defCtx, nil
	} else {

		sess := db.Redis.Session()
		defer sess.Close()

		var ctx ContextData
		sess.ReadEntity(fmt.Sprintf("eco:context:hash:%s:%s", domain, id), &ctx)
		sess.Close()

		if ctx.ContextBase.ID != id {
			return nil, halt.Errorf("context %s not found", http.StatusNotFound, id)
		}
		return ctx, nil
	}
}

//contextos persistentes
type ModuleAccess struct {
	Key     string `json:"key"`
	Granted bool   `json:"granted"`
}

type ModuleAccesses struct {
	EcoChat           []ModuleAccess `json:"ecoChat"`
	EcoTasks          []ModuleAccess `json:"ecoTasks"`
	EcoWall           []ModuleAccess `json:"ecoWall"`
	EcoAutomanagement []ModuleAccess `json:"ecoAutomanagement"`
}

type ContextMember struct {
	Email     string          `json:"email"`
	FirstName string          `json:"firstName"`
	LastName  string          `json:"lastName"`
	Access    *ModuleAccesses `json:"access,omitempty"`
}

type ContextFeature struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Kind string `json:"kind"`
}

type ContextBase struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type Config struct {
	Key   string `json:"key"`
	Value interface{} `json:"value"`
}

type Widgets struct {
	Type   string   `json:"type"`
	Config []Config `json:"config"`
}

type MultiTypeConfig struct {
	Type    string    `json:"type"`
	Widgets []Widgets `json:"widgets"`
}

type ContextData struct {
	ContextBase
	Domain   string           `json:"domain"`
	Color    string           `json:"color,omitempty"`
	Email    string           `json:"email,omitempty"`
	Modules  []modules.Module `json:"modules,omitempty"`
	Members  []ContextMember  `json:"members,omitempty"`
	Features []ContextFeature `json:"features,omitempty"`
	Deals    MultiTypeConfig  `json:"deals,omitempty"`
	Invoices MultiTypeConfig  `json:"invoices,omitempty"`
	Portal   MultiTypeConfig  `json:"portal,omitempty"`
	db.AuditData
}

func (c ContextData) ID() string {
	return c.ContextBase.ID
}

func (c ContextData) Name() string {
	return c.Domain + " " + c.ContextBase.Name + " context"
}

func (c ContextData) RedisFillFields(ent interface{}, data *[]interface{}) {

	*data = append(*data,
		"name", c.ContextBase.Name,
		"description", c.Description,
		"color", c.Color,
		"email", c.Email)

	c.AuditData.RedisFillFields(ent, data)
}

func (c ContextData) IsMember(email string) bool {
	for _, m := range c.Members {
		if m.Email == email {
			return true
		}
	}
	return false
}

func (c ContextData) HasFeature(featureID string) bool {
	for _, f := range c.Features {
		if f.ID == featureID {
			return true
		}
	}
	return false
}

func (c ContextData) ValidateModules() error {

	for _, m := range c.Modules {
		if modules.Find(m.ID) == nil {
			return halt.Errorf("module '%s' not found", 500, m.ID)
		}
	}
	return nil
}
