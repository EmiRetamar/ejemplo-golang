package security

import (
	"eco/services/session"
	"eco/services/persistence/db"
	"eco/services/persistence"
	"github.com/pborman/uuid"
	"eco/services/halt"
	"eco/services/net"
	"fmt"
)

/**
 * User: Santiago Vidal
 * Date: 20/06/17
 * Time: 22:54
 */


type permissionsManager struct{
	session *session.EcoSession
}
func (pm *permissionsManager) CanAccess() (bool, error) {
	//s := db.Redis.Session()
	//defer s.Close()

	//_, r := pm.session.GetHttp()
	//key := r.Method + ":" + r.URL.Path
	//return s.Sets("eco_permissions:url:" + pm.session.EcoUser.Email).IsMember(key)
	return true, nil
}

func Session(s *session.EcoSession) *permissionsManager {
	return &permissionsManager{
		session: s,
	}
}

func (pm *permissionsManager) Create(p *persistence.Permission, d *persistence.Domain) error {

	s := db.Redis.Session()
	defer s.Close()

	p.Audit(pm.session.EcoUser.Email)
	p.Id = uuid.New()
	if err := s.Store("eco:permission:" + p.Id, p); err != nil {
		return err
	}
	return pm.addToDomainsSet(p, d.Name)
}

func (pm *permissionsManager) Update(p *persistence.Permission, d *persistence.Domain) error {

	s := db.Redis.Session()
	defer s.Close()

	p.Audit(pm.session.EcoUser.Email)
	if exists, err := s.Sets("eco:permissions:list:" + d.Name).IsMember(p.Id); err != nil || !exists {

		if err != nil {
			return err

		}
		return halt.Errorf("permission not found", 404)
	}


	if err := s.Store("eco:permission:" + p.Id, p); err != nil {
		return err
	}
	return nil
}

func (pm *permissionsManager) Get(id string) (*persistence.Permission, error) {

	s := db.Redis.Session()
	defer s.Close()

	var p persistence.Permission
	if err := s.ReadEntity(id, p); err != nil {
		return nil, err
	}
	return &p, nil
}

func (pm *permissionsManager) List(domain string) ([]persistence.Permission, error) {

	s := db.Redis.Session()
	defer s.Close()

	rawList, err := s.SendRaw("SORT eco:permissions:list:" + domain + " by nosort get # get eco:permission:*->name")
	if err != nil {
		return nil, err
	}
	retList := make([]persistence.Permission, len(rawList)/2)

	r := 0
	for i := 0; i < len(retList); i++ {

		retList[i].Id = rawList[r]
		retList[i].Name = rawList[r+1]
		r += 2

	}
	return retList, nil
}

func (pm *permissionsManager) addToDomainsSet(p *persistence.Permission, domain string) error {

	s := db.Redis.Session()
	defer s.Close()

	return s.Sets("eco:permissions:list:" + domain).Add(p.Id)
}

func CanAccess(s *session.EcoSession, featureID string, email string) bool {

	type data struct {
		Access string
	}
	url := fmt.Sprintf("/api/1/security/features/%s/%s", s.Domain, email)
	api := net.NewEcoApi(s, "security", url)

	var d data
	if err := api.Get(featureID, &d); err != nil {
		return false
	}

	return d.Access == "granted"
}