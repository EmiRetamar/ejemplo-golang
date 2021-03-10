package persistence

import (
	"eco/services/persistence/db"
)

/**
 * User: Santiago Vidal
 * Date: 16/05/17
 * Time: 15:40
 */

type Domain struct {
	Name        string `json:"name"`
	Server      string `json:"server"`
	Credentials struct {
		ClientID  string `json:"client_id"`
		SecretKey string `json:"secret_key"`
	} `json:"credentials"`
	db.AuditData
}
