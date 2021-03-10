package check

import (
	"eco/services/net"
	"eco/services/session"
	"eco/services/session/oauth"
	"net/http"
	"eco/services/security"
	"eco/services/halt"
	"github.com/gorilla/mux"
)

/**
 * User: Santiago Vidal
 * Date: 22/05/17
 * Time: 15:54
 */

func LoginCheck(inner net.EcoApiHandler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s, err := session.GetSession(w, r)
		if err != nil {
			net.HTTPError(w, err)
			return
		}
		session.ExtendSession(w, r)
		err = inner.ServeHTTP(s)
		if err != nil {
			net.HTTPError(w, err)
		}
	})
}

func ApiCheck(inner net.EcoApiHandler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		defer halt.HandlePanics(w)

		w.Header().Set("Content-Type", "application/json; charset=UTF-8")

		var s *session.EcoSession
		var err error
		token := r.URL.Query().Get("access_token")

		if len(token) > 0 {
			s, err = oauth.GetSession(token, w, r)
		} else {
			//el front-end de eco tiene permitido usar la api sin token si ya se encuentra logueado
			s, err = session.GetSession(w, r)
			//err = errors.New("missing access token")
		}

		if err != nil {
			net.HTTPError(w, err)
			return
		}

		if len(token) == 0 { //solo se extiende por sessionID. AccesTokens expiran siempre.
			session.ExtendSession(w, r)
		}

		path := r.URL.Path
		if len(path) > 15 && path[0:16] == "/api/1/teamplace" {

			err := net.Service("domains").Call("/api/1/domains/" + s.Domain, "GET", nil, nil)
			if err != nil || (s.EcoUser.IsRoot && len(s.Domain) == 0) {
				net.HTTPError(w, halt.Errorf("no teamplace domain found for user %s", http.StatusForbidden, s.EcoUser.Email))
				return
			}

		}

		err = inner.ServeHTTP(s)
		if err != nil {
			net.HTTPError(w, err)
		}

	})
}

func AdminCheck(inner net.EcoApiHandler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		s, err := session.GetSession(w, r)
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
			return
		}

		if !s.EcoUser.IsRoot {
			ecoError := halt.Errorf("access denied", http.StatusForbidden)
			ecoError.HTTPError(w)
			return
		}

		session.ExtendSession(w, r)
		err = inner.ServeHTTP(s)

		if err != nil {
			net.HTTPError(w, err)
		}
	})
}

func SecurityCheck(inner net.EcoApiHttp) net.EcoApiHandler {

	return net.EcoApiHandler(func(sess *session.EcoSession) error {

		_, r := sess.GetHttp()
		if !sess.EcoUser.IsAdmin {

			if !security.CanAccess(sess, mux.Vars(r)["featureID"], sess.EcoUser.Email) {
				return halt.Errorf("access denied", http.StatusForbidden)
			}
		}
		return inner.ServeHTTP(sess)

	})
}
