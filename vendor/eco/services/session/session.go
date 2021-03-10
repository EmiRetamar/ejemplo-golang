package session

import (
	"net/http"
	"github.com/adam-hanna/sessions"
	"github.com/adam-hanna/sessions/store"
	"github.com/adam-hanna/sessions/auth"
	"github.com/adam-hanna/sessions/transport"
	"time"
	"encoding/json"
	"eco/services/config"
	"strings"
	"io"
	"eco/services/persistence/db"
	"eco/services/jobs"
	"eco/services/halt"
	"fmt"
)

/**
 * User: Santiago Vidal
 * Date: 11/05/17
 * Time: 14:41
 */

type EcoSessionAuth interface {
	Validate() error
	User() string
	Domain() string
}

type EcoSession struct {
	EcoUser struct {
		Email           string
		IsRoot          bool
		IsAdmin         bool
	}
	Domain string
	Context  struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}

	jobID  string
	accessToken string
	contextCreation *bool
	request *http.Request
	response http.ResponseWriter
}
func (s *EcoSession) GetHttp() (http.ResponseWriter, *http.Request) {
	return s.response, s.request
}
func (s *EcoSession) Logout() {
	DeleteSession(s.response, s.request)
}
func (s *EcoSession) AccessToken() string {
	return s.accessToken
}
func (s *EcoSession) JobID() string {
	return s.jobID
}
func (s *EcoSession) RunningAsJob() bool {
	return len(s.jobID) > 0
}
func (s *EcoSession) ContextCreation() bool {

	if s.contextCreation != nil {
		return *s.contextCreation
	}

	type profileData struct {
		ContextCreation bool `json:"contextCreation"`
	}
	sess := db.Redis.Session()
	defer sess.Close()

	var profile profileData
	sess.ReadEntity(fmt.Sprintf("eco:users:profiles::%s:%s", s.Domain, s.EcoUser.Email), &profile)
	s.contextCreation = &profile.ContextCreation

	return profile.ContextCreation
}

func CreateJobSession(mainSession *EcoSession, input io.ReadCloser) *jobs.CreationInfo {

	req := mainSession.request
	req.Body = input

	resp := jobResponseWriter{}
	jobInfo := jobs.New(resp)
	resp.jobID = jobInfo.JobID


	ecoSess := &EcoSession{
		jobID: jobInfo.JobID,
		accessToken: mainSession.accessToken,
		request: req,
		response: resp,
		Domain: mainSession.Domain,
		EcoUser: mainSession.EcoUser,
	}

	jobInfo.SetJobData(ecoSess)
	return jobInfo
}

func NewEmptyEcoSession(token string, w http.ResponseWriter, r *http.Request) *EcoSession {
	return &EcoSession{
		accessToken: token,
		request: r,
		response: w,
	}
}


type ecoSessionData struct {
	Email           string
	IsRoot          bool
	IsAdmin         bool
	ContextCreation bool
	Active          bool
	ActiveDomain    string
}

var sesh *sessions.Service

func Init() {

	host, _ := config.EcoCfg.String("database.redis.host")
	port, _ := config.EcoCfg.String("database.redis.port")
	sessStore := store.New(store.Options{
		ConnectionAddress: host + ":" + port,
	})

	sessAuth,_ := auth.New(auth.Options{
		Key: []byte("EHqweI/7j2tNE39xU0jHMwNr7yFDtc3Mn1sAyQqDy7D91koGppNPtfflOf7WqBHaI4A45VXXFmTRbuQ2dTQ4WA=="),
	})

	sessTransport := transport.New(transport.Options{
		HTTPOnly: true,
		Secure:   false, // note: can't use secure cookies in development!
	})

	var expiration time.Duration
	debug, _ := config.EcoCfg.Bool("debug")
	if debug {
		expiration = time.Hour * 8
	} else {
		expiration = time.Minute * 10
	}
	sesh = sessions.New(sessStore, sessAuth, sessTransport, sessions.Options{ ExpirationDuration: expiration })

}

//noinspection GoUnusedExportedFunction
func NewSession(email string, domain string, root bool) (*EcoSession, error) {

	return &EcoSession{
		EcoUser: struct {
			Email   string
			IsRoot bool
			IsAdmin bool
		}{
			Email:  email,
			IsRoot: root,
			IsAdmin: false,
		},
		Domain: domain,

	}, nil

}

func GetSession(w http.ResponseWriter, r *http.Request) (ret *EcoSession, err error) {

	s, err := sesh.GetUserSession(r)
	if err != nil {
		if s == nil {
			return nil, halt.Errorf(http.StatusText(401), 401)
		} else {
			return nil, err
		}
	}
	if s == nil {
		return nil, halt.Errorf(http.StatusText(401), 401)
	}

	var data ecoSessionData
	json.NewDecoder(strings.NewReader(s.JSON)).Decode(&data)

	return &EcoSession{
		Domain: data.ActiveDomain,
		request: r,
		response: w,
		EcoUser: struct {
			Email   string
			IsRoot bool
			IsAdmin bool
		}{
			Email: data.Email,
			IsRoot: data.IsRoot,
			IsAdmin: data.IsAdmin,
		},

	}, nil
}

func DeleteSession(w http.ResponseWriter, r *http.Request) (err error) {

	s, err := sesh.GetUserSession(r)
	if err == nil {
		sesh.ClearUserSession(s, w)
	}
	return err

}

func ExtendSession(w http.ResponseWriter, r *http.Request) (err error) {

	s, err := sesh.GetUserSession(r)
	if err == nil {
		sesh.ExtendUserSession(s, r, w)
	}
	return err
}


/*** job response ****/

type jobResponseWriter struct {
	jobID string
}

func (jrw jobResponseWriter) Header() http.Header {
	return http.Header{}
}
func (jrw jobResponseWriter) Write(data []byte) (int, error) {

	s := db.Redis.Session()
	defer s.Close()

	var ji jobs.ExecutionInfo
	s.ReadEntity(jobs.Key + jrw.jobID, &ji)

	err := s.Store(jobs.Key + jrw.jobID, &jobs.ExecutionInfo{
		ID:     ji.ID,
		Status: "finished",
		Result: ji.Result + string(data[:]),
	})
	return len(data), err
}
func (jrw jobResponseWriter) WriteHeader(int) {

}
