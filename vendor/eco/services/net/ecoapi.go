package net

import (
	"bytes"
	"eco/services/config"
	"eco/services/halt"
	"eco/services/session"
	"eco/services/session/oauth"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"eco/services/crypto"
	"github.com/google/logger"
)

/**
 * User: Santiago Vidal
 * Date: 25/09/17
 * Time: 15:56
 */

type apiResult struct {
	Status int    `json:"status"`
	Error  string `json:"error"`
}

type ecoApi struct {
	sess            *session.EcoSession
	service         string
	api             string
	jobMode         bool
	useID			bool
	queryParameters map[string]string
}

func NewEcoApi(s *session.EcoSession, service string, api string) *ecoApi {
	return &ecoApi{
		sess:    s,
		service: service,
		api:     api,
		useID:   true,
	}
}

func RootAuth(w http.ResponseWriter, r *http.Request) (*session.EcoSession, error) {

	baseUrl, err := getBaseUrl(r, "oauth")
	if err != nil {
		return nil, err
	}

	id, _ := config.EcoCfg.String("root.clientId")
	secret, _ := config.EcoCfg.String("root.secretKey")
	aesCryp, err := crypto.NewCrypto(crypto.AES)

	if err != nil {
		logger.Error(err)
	}

	isInternal, err := aesCryp.Encode("true")

	if err != nil {
		logger.Error(err)
	}

	var sess *session.EcoSession

	url := fmt.Sprintf("%s/auth/token?client_id=%s&secret_key=%s&access=root&isInternal=%s", baseUrl, id, secret, isInternal)
	_, err = DoGet(url, "GET", func(response *http.Response) error {

		var t oauth.EcoToken
		json.NewDecoder(response.Body).Decode(&t)

		sess = session.NewEmptyEcoSession(t.Token, w, r)
		return nil
	})

	if err != nil {
		return nil, err
	}
	return sess, nil
}

func (e *ecoApi) SwitchToJobMode() {
	e.jobMode = true
}

func (e *ecoApi) NotUseID()  {
	e.useID = false
}

func (e *ecoApi) QueryParameters(parameters map[string]string) {
	e.queryParameters = parameters
}

func (e *ecoApi) Get(id string, out interface{}) error {

	_, r := e.sess.GetHttp()
	url, err := getBaseUrl(r, e.service)
	if err != nil {
		return err
	}
	url += e.api + "/" + id + "?access_token=" + e.sess.AccessToken()

	if e.queryParameters != nil {
		for k, v := range e.queryParameters {
			url += "&" + k + "=" + v
		}
	}

	if e.jobMode {
		url += "&mode=job"
	}

	if _, err := DoGet(url, "GET", func(response *http.Response) error {

		if response.StatusCode != 200 {
			return e.processResponse(response)
		}

		if out == nil {
			return nil
		}

		defer response.Body.Close()
		if w, ok := out.(io.Writer); ok {
			_, err := io.Copy(w, response.Body)
			return err
		} else {
			return json.NewDecoder(response.Body).Decode(out)
		}

	}); err != nil {
		return err
	}

	return nil
}

func (e *ecoApi) Post(in interface{}, out interface{}) error {

	var data, err = json.Marshal(in)
	if err != nil {
		return err
	}

	_, r := e.sess.GetHttp()
	url, err := getBaseUrl(r, e.service)

	if err != nil {
		return err
	}
	url += e.api + "?access_token=" + e.sess.AccessToken()

	if e.queryParameters != nil {
		for k, v := range e.queryParameters {
			url += "&" + k + "=" + v
		}
	}

	if e.jobMode {
		url += "&mode=job"
	}

	if _, err := DoPost(url, "POST", bytes.NewReader(data), func(response *http.Response) error {

		if response.StatusCode != 200 {
			return e.processResponse(response)

		} else if out != nil {

			defer response.Body.Close()

			if w, ok := out.(io.Writer); ok {
				io.Copy(w, response.Body)
			} else {
				json.NewDecoder(response.Body).Decode(out)
			}
		}
		return nil

	}); err != nil {
		return err
	}
	return nil
}

func (e *ecoApi) Put(id string, in interface{}, out interface{}) error {

	var data, err = json.Marshal(in)
	if err != nil {
		return err
	}

	_, r := e.sess.GetHttp()
	url, err := getBaseUrl(r, e.service)

	if err != nil {
		return err
	}


	url += e.api

	//solo agrego ID si lo usa
	if e.useID {
		url += "/" + id
	}

	url += "?access_token=" + e.sess.AccessToken()

	if e.jobMode {
		url += "&mode=job"
	}

	if _, err := DoPost(url, "PUT", bytes.NewReader(data), func(response *http.Response) error {

		if response.StatusCode != 200 {
			return e.processResponse(response)

		} else if out != nil {

			defer response.Body.Close()
			json.NewDecoder(response.Body).Decode(out)
		}
		return nil

	}); err != nil {
		return err
	}
	return nil

}

func (e *ecoApi) Delete(id string) error {

	_, r := e.sess.GetHttp()
	url, err := getBaseUrl(r, e.service)

	if err != nil {
		return err
	}
	url += e.api + "/" + id + "?access_token=" + e.sess.AccessToken()

	if e.jobMode {
		url += "&mode=job"
	}

	if _, err := DoGet(url, "DELETE", func(response *http.Response) error {

		return e.processResponse(response)

	}); err != nil {
		return err
	}

	return nil

}

func (e *ecoApi) processResponse(r *http.Response) error {

	defer r.Body.Close()
	if r.StatusCode != 200 {

		var result apiResult
		if err := json.NewDecoder(r.Body).Decode(&result); err != nil {
			return err
		}
		return halt.Errorf(e.service+" service error: "+result.Error, 500)
	}
	return nil

}

func getBaseUrl(r *http.Request, service string) (string, error) {

	if len(r.Host) > 8 && r.Host[:9] == "localhost" {

		//en modo localhost hay que obtener en que puerto esta corriendo el servicio solicitado
		url := "http://localhost:" + getLocalServicePort(service)
		_, err := DoGet(url+"/who", "GET", func(response *http.Response) error {

			type result struct {
				Who string `json:"who"`
			}

			var r result
			json.NewDecoder(response.Body).Decode(&r)

			if r.Who != service {
				return halt.Errorf("couldn't find '%s' service. service response was '%s'", 500, service, r.Who)
			}
			return nil

		})
		if err != nil {

			if _, ok := err.(halt.EcoError); ok {
				return "", err
			} else {
				return "", halt.Errorf("couldn't find %s service: %s", 500, service, err.Error())
			}

		}
		return url, nil

	} else {
		return "https://" + r.Host, nil
	}

}

func getLocalServicePort(service string) string {
	content, _ := ioutil.ReadFile("/tmp/eco-services/" + service)
	return string(content)
}
