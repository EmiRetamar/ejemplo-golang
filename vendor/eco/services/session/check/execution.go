package check

import (
	"eco/services/net"
	"eco/services/session"
	"encoding/json"
	"bytes"
	"io"
)

/**
 * User: Santiago Vidal
 * Date: 13/10/17
 * Time: 19:56
 */


//ExecutionCheck chequea si la api debe ejecutarse asincronicamente (en un job) y de ser asi devuelve inmediatamente un job-ID
//Lo hace verificando si se llamo con el parametro "?mode=job"
//Dicho job puede monitorearse haciendo GET a /api/1/jobs/monitor/{jobID} del servicio Eco-Backend-Srv-JobMonitor
func ExecutionCheck(inner net.EcoApiHttp) net.EcoApiHandler {

	return net.EcoApiHandler(func(sess *session.EcoSession) error {

		w, r := sess.GetHttp()
		param := r.URL.Query().Get("mode")

		if param == "job" {

			defer r.Body.Close()
			input := memoryReader{
				buffer: new(bytes.Buffer),
			}
			input.ReadFrom(r.Body)

			jobInfo := session.CreateJobSession(sess, input)
			jobInfo.Start(func(jobSession interface{}) error {
				return inner.ServeHTTP(jobSession.(*session.EcoSession))
			})

			return json.NewEncoder(w).Encode(jobInfo)

		} else {
			return inner.ServeHTTP(sess)
		}

	})
}

type memoryReader struct {
	buffer *bytes.Buffer
}
func(r memoryReader) ReadFrom(reader io.Reader) (n int64, err error) {
	return r.buffer.ReadFrom(reader)
}
func(r memoryReader) Read(p []byte) (n int, err error) {
	return r.buffer.Read(p)
}
func(r memoryReader) Close() error {
	//nada
	return nil
}
