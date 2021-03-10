package persistence

import (
	"eco/services/persistence/db"
	"eco/services/session"
	"eco/services/register"
)

type IFeature interface {
	OnLoaded() error
	Audit()
	Info() *Feature
	Validate() error
	GetEndPointURL(url register.FeatureFlags) string
	GetServiceName() string
	NeedsJobMode() bool
	db.RedisAware
	db.AuditEntity
}

type Feature struct {
	ID              string              `json:"id"`
	Name            string              `json:"name"`
	Description     string              `json:"description"`
	Kind            string              `json:"kind"`
	Domain          string              `json:"domain"`
	SecurityBased   bool                `json:"securityBased"`
	PublicFeatureID string              `json:"publicFeatureID,omitempty"`
	Active          bool                `json:"active"`
	URLAlias        string              `json:"urlAlias"`
	DevelopGroups   []string            `json:"developGroups"`
	Tags            []string            `json:"tags"`
	Session         *session.EcoSession `json:"-"`
	db.AuditData
}

func (f Feature) Info() *Feature {
	return &f
}

func (f Feature) Validate() error {
	return nil
}

func (f Feature) IsPublic() bool {
	return len(f.Domain) == 0
}

func (f Feature) AuditInfo() (*db.AuditData, string) {
	return f.AuditData.AuditInfo()
}
func (f Feature) Audit() {
	if f.Session != nil {
		f.AuditData.Audit(f.Session.EcoUser.Email)
	}
}

func (f Feature) OnLoaded() error {
	return nil
}

func (f Feature) GetEndPointURL(url register.FeatureFlags) string {
	return ""
}

func (f Feature) GetServiceName() string {
	return ""
}

func (f Feature) NeedsJobMode() bool {
	return false
}

func (f Feature) RedisFillFields(ent interface{}, data *[]interface{}) {

	*data = append(*data,
		"name", f.Name,
			"description", f.Description,
			"kind", f.Kind,
			"domain", f.Domain,
			"publicFeatureID", f.PublicFeatureID,
			"active", f.Active)

	f.AuditData.RedisFillFields(ent, data)
}