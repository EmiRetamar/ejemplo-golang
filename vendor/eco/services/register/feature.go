package register

import (
	"eco/services/persistence/db"
	"eco/services/halt"
)

/**
 * User: Santiago Vidal
 * Date: 23/02/18
 * Time: 19:17
 */

const key = "eco:features:kind:"

type FeatureFlags int

const VALIDATE_URL = 1
const FETCH_URL = 2
const EXECUTE_URL = 4
const INSERT_URL = 8
const UPDATE_URL = 16
const DELETE_URL = 32
const JOB_MODE = 64

type featureRegistration struct {
	db.EcoEntity
	URL          string `json:"url"`
	NeedsJobMode bool   `json:"needsJobMode,omitempty"`
}

type FeaturesManager struct{}

var Features FeaturesManager

func (f FeaturesManager) NewKind(kind string, flags FeatureFlags, baseURL string) {

	if err := f.registerKind(kind, flags, baseURL); err != nil {
		panic(err)
	}

}

func (f FeaturesManager) GetURL(kind string, urlType FeatureFlags) (string, bool) {

	var field string
	switch urlType {

	case EXECUTE_URL:
		field = "execute"

	case VALIDATE_URL:
		field = "validate"

	case FETCH_URL:
		field = "fetch"

	case INSERT_URL:
		field = "insert"

	case UPDATE_URL:
		field = "update"

	case DELETE_URL:
		field = "delete"

	default:
		panic(halt.Errorf("feature registration: unknown URLType", 500))

	}

	s := db.Redis.Session()
	defer s.Close()

	var fr featureRegistration
	if err := s.ReadEntity(key + kind + ":" + field, &fr); err != nil {
		panic(halt.Errorf("feature kind '%s' redis errror: %s", 500, kind, err))
	}

	if len(fr.URL) > 0 {
		return fr.URL, fr.NeedsJobMode
	}

	//si ese tipo de url no estaba registrada verifico si "validate" existe, ya que esa es la unica obligatoria
	//si tampoco existe entonces esta tipo de feature no esta registrado!
	if exists, err := s.Exists(key + kind + ":" + "validate"); !exists {

		if err != nil {
			panic(halt.Errorf("feature kind '%s' redis errror: %s", 500, kind, err))
		}

		if !exists {
			panic(halt.Errorf("feature registration not found: " + kind, 500))
		}
	}
	return "", false

}

func (f FeaturesManager) registerKind(kind string, flags FeatureFlags, baseURL string) error {

	s := db.Redis.Session()
	defer s.Close()

	fr := featureRegistration{}

	if flags & VALIDATE_URL != 0 {
		fr.URL = baseURL + "/validate"
		if err := s.Store(key + kind + ":validate", fr); err != nil {
			return err
		}
	} else {
		s.Remove(key + kind + ":validate")
	}

	if flags & FETCH_URL != 0 {
		fr.URL = baseURL + "/fetch"
		if err := s.Store(key + kind + ":fetch", fr); err != nil {
			return err
		}
	} else {
		s.Remove(key + kind + ":fetch")
	}

	if flags & INSERT_URL != 0 {
		fr.URL = baseURL + "/insert"
		if err := s.Store(key + kind + ":insert", fr); err != nil {
			return err
		}
	} else {
		s.Remove(key + kind + ":insert")
	}

	if flags & UPDATE_URL != 0 {
		fr.URL = baseURL + "/update"
		if err := s.Store(key + kind + ":update", fr); err != nil {
			return err
		}
	} else {
		s.Remove(key + kind + ":update")
	}

	if flags & DELETE_URL != 0 {
		fr.URL = baseURL + "/delete"
		if err := s.Store(key + kind + ":delete", fr); err != nil {
			return err
		}
	} else {
		s.Remove(key + kind + ":delete")
	}

	if flags & EXECUTE_URL !=0 {
		fr.URL = baseURL + "/execute"
		fr.NeedsJobMode = (flags & JOB_MODE !=0)
		if err := s.Store(key + kind + ":execute", fr); err != nil {
			return err
		}
	} else {
		s.Remove(key + kind + ":execute")
	}

	return nil
}
