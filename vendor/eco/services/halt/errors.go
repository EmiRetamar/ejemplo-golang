package halt

import (
	"io"
	"encoding/json"
	"bytes"
	"net/http"
	"fmt"
	"errors"
	"runtime"
	"os"
	"strconv"
	"math/rand"
	"log"
)

/**
 * User: Santiago Vidal
 * Date: 22/06/17
 * Time: 17:45
 */

type EcoError struct {
	status int
	err    error
}

func (e *EcoError) marshal(w io.Writer) {

	type DisplayError struct {
		Status int    `json:"status"`
		Error  string `json:"error"`
	}

	json.NewEncoder(w).Encode(DisplayError{
		Status: e.status,
		Error: e.err.Error(),
	})
}

func (e EcoError) StatusCode() int {
	return e.status
}

func (e EcoError) Error() string {

	var buf = new(bytes.Buffer)
	e.marshal(buf)
	return buf.String()
}

func (e EcoError) Message() string {
	return e.err.Error()
}

func (e EcoError) HTTPError(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(e.status)
	e.marshal(w)
}

func Error(err error, status int) EcoError {

	if eco, ok := err.(EcoError); ok {
		return eco
	}

	return EcoError{
		status: status,
		err:    err,
	}
}
func Errorf(format string, status int,  args ...interface{}) EcoError {
	return EcoError{
		status: status,
		err:    errors.New(fmt.Sprintf(format, args...)),
	}
}

func ErrorEco(format string, err error, args ...interface{}) EcoError {

	if ecoErr, ok := err.(EcoError); ok {

		return EcoError{
			status: ecoErr.StatusCode(),
			err: errors.New(fmt.Sprintf(format + " " + ecoErr.Message(), args...)),
		}
	} else {
		return Errorf(format, 500, args)
	}

}

/*****/

func PanicHandler(err interface{}, out io.Writer) {

	if ecoErr, ok := err.(EcoError); ok {

		if rw, ok := out.(http.ResponseWriter); ok {
			ecoErr.HTTPError(rw)
		} else {
			ecoErr.marshal(out)
		}
		return
	}

	errorID := strconv.FormatUint(rand.Uint64(), 10)

	buf := make([]byte, 1<<16)
	stackSize := runtime.Stack(buf, false)
	log.Printf("** !PANIC ERROR ID=%s! **\n", errorID)
	log.Printf("Panic: %v\n", err)
	log.Printf("%s\n", string(buf[0:stackSize]))

	os.Stderr.Write(buf[0:stackSize])

	out.Write([]byte(Errorf("fatal error - ID:%s - %v", 500, errorID, err).Error()))

}

func HandlePanics(w http.ResponseWriter) {

	if r := recover(); r != nil {

		if ecoErr, ok := r.(EcoError); ok {
			ecoErr.HTTPError(w)
			return
		}

		w.WriteHeader(500)
		errorID := strconv.FormatUint(rand.Uint64(), 10)

		buf := make([]byte, 1<<16)
		stackSize := runtime.Stack(buf, false)
		log.Printf("** !PANIC ERROR ID=%s! **\n", errorID)
		log.Printf("%s\n", string(buf[0:stackSize]))

		os.Stderr.Write(buf[0:stackSize])

		w.Write([]byte(Errorf("fatal error - ID:%s - %v", 500, errorID, r).Error()))
	}
}
