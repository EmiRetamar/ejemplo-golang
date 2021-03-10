package net

/**
 * User: Santiago Vidal
 * Date: 01/05/17
 * Time: 14:12
 */

import (
	"eco/services/session"
	"encoding/json"
	"io"
	"net/http"
	"time"
	"eco/services/halt"
)

func HTTPError(w http.ResponseWriter, err error) {
	if ecoErr, ok := err.(halt.EcoError); ok {
		ecoErr.HTTPError(w)
	} else {
		halt.Error(err, 500).HTTPError(w)
	}
}

type EcoApiHttp interface {
	ServeHTTP(s *session.EcoSession) error
}

type EcoApiHandler func(s *session.EcoSession) error

func (f EcoApiHandler) ServeHTTP(s *session.EcoSession) error {
	return f(s)
}

func DoGet(apiUrl string, method string, callback func(response *http.Response) error) (duration time.Duration, outErr error) {

	startTime := time.Now()
	req, err := http.NewRequest(method, apiUrl, nil)
	if err != nil {
		outErr = err
		return
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		outErr = err
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {

		type ResponseJson struct {
			Status int    `json:"status"`
			Error  string `json:"error"`
		}
		var responseData ResponseJson
		json.NewDecoder(resp.Body).Decode(&responseData)

		return time.Since(startTime), halt.Errorf(responseData.Error, resp.StatusCode)
	}

	outErr = callback(resp)

	duration = time.Since(startTime)
	return

}

func DoPost(apiUrl string, method string, body io.Reader, callback func(response *http.Response) error) (duration time.Duration, outErr error) {

	startTime := time.Now()
	req, err := http.NewRequest(method, apiUrl, body)
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		outErr = err
		return
	}
	defer resp.Body.Close()
	outErr = callback(resp)

	duration = time.Since(startTime)
	return

}

func ReadJsonRequest(r *http.Request, out interface{}) error {

	body := r.Body
	defer body.Close()

	in := SafePipeReader(body)
	defer in.Close()

	err := json.NewDecoder(in).Decode(out)
	if err != nil {
		return halt.Error(err, http.StatusBadRequest)
	}
	return nil
}

func SafePipeReader(reader io.Reader) (*io.PipeReader) {

	pipeIn, pipeOut := io.Pipe()
	go func() {
		const mbLimit = 5
		const bytesLimit = mbLimit * 1024 * 1024 + 1//no levanto mas de 5Mb de lo que el cliente este queriendo postear

		safeReader := io.LimitReader(reader, bytesLimit)
		written, _ := io.Copy(pipeOut, safeReader)

		if written == bytesLimit {
			//aborto el envio y escupo error
			pipeOut.CloseWithError(halt.Errorf("posted body is too large. Limit is %vMb", http.StatusNotAcceptable, mbLimit))
		} else {
			pipeOut.Close()
		}

	}()
	return pipeIn
}

type ecoService struct {
	name string
}
func (e *ecoService) Call(url string, method string, body interface{}, output interface{}) error {
	return nil
}

func Service(serviceName string) *ecoService {
	return &ecoService{
		name: serviceName,
	}
}
