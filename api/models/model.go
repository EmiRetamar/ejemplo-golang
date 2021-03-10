package models

import (
	"eco/services/persistence"
)

type InfoToken struct {
	Admin bool `json:"admin"`
	Internal bool `json:"internal"`
}

type ContextList struct {
	persistence.ContextBase
	Color     string `json:"color"`
	Email     string `json:"email"`
	CreatedBy string `json:"createdBy,omitempty"`
}

type FeatureData struct {
	FeatureId    string `json:"featureId"`
	FeatureName  string `json:"featureName"`
	FeatureKind  string `json:"featureKind"`
	ContextId    string `json:"contextId"`
	ContextName  string `json:"contextName"`
	ContextColor string `json:"contextColor"`
}

type ClientData struct {
	Id             int    `json:"id"`
	Code           string `json:"CODIGO"`
	Identification string `json:"IDENTIFICACION"`
	Caption        string `json:"caption"`
}

type ProviderData struct {
	Id                int    `json:"id"`
	AccountProviderId string `json:"CUENTAIDPROVEEDOR"`
	Code              string `json:"CODIGO"`
	Identification    string `json:"IDENTIFICACION"`
	Caption           string `json:"caption"`
	AccountCode       string `json:"CUENTACODIGO"`
}

type UserData struct {
	persistence.User
}

type MenuOption struct {
	Caption      string `json:"caption"`
	Id           int    `json:"id"`
	ShortCaption string `json:"shortCaption"`
	Type         int    `json:"tipo"`
}

type SearchData struct {
	TeamplaceMenuOptions []MenuOption   `json:"teamplaceMenuOptions"`
	Features             []FeatureData  `json:"features"`
	Clients              []ClientData   `json:"clients"`
	Providers            []ProviderData `json:"providers"`
	Users                []UserData     `json:"users"`
}
