package persistence

import (
	"eco/services/persistence/db"
)

/**
 * User: Santiago Vidal
 * Date: 22/05/17
 * Time: 17:56
 */

const TagsSet = "eco:tags:" //por domain
const TagsFeatureSet = "eco:feature:tags:" //por feature

func Init() error {
	if err := db.Init(); err != nil {
		return err
	}
	return nil
}
