package oauth

import (
	"time"
	"encoding/json"
	"io"

	"eco/services/session"
	"net/http"
	"eco/services/persistence/db"
	"eco/services/halt"
)

/**
 * User: Santiago Vidal
 * Date: 23/05/17
 * Time: 14:20
 */
type EcoToken struct {
	db.EcoEntity
	Token     string        `json:"token"`
	User      string        `json:"email"`
	Domain    string        `json:"domain"`
	CreatedAt time.Time     `json:"createdAt"`
	ExpiresAt time.Time     `json:"expiresAt"`
	Admin     bool		    `json:"admin"`
	Root      bool          `json:"root"`
}

func (t *EcoToken) String() string {
	return t.Token
}
func (t *EcoToken) StreamJson(w io.Writer) {
	json.NewEncoder(w).Encode(t)
}

func GetSession(tokenStr string, w http.ResponseWriter, r *http.Request) (*session.EcoSession, error) {

	//var data string
	var tokenData EcoToken
	{
		s := db.Redis.Session()
		defer s.Close()

		if err := s.ReadEntity("eco:oauth:token:" + tokenStr, &tokenData); err != nil {
			return nil, err
		}
	}

	if len(tokenData.Token) > 0 {

		s := session.NewEmptyEcoSession(tokenData.Token, w, r)
		s.Domain = tokenData.Domain
		s.EcoUser.Email = tokenData.User
		s.EcoUser.IsRoot = tokenData.Root
		s.EcoUser.IsAdmin = tokenData.Admin
		return s, nil

	} else {
		return nil, halt.Errorf("invalid token", http.StatusUnauthorized)
	}

}
