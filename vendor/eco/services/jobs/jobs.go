package jobs

import (
	"eco/services/persistence/db"
	"time"
	"github.com/pborman/uuid"
	"io"
	"eco/services/halt"
	"github.com/ian-kent/go-log/log"
	"errors"
	"fmt"
)

/**
 * User: Santiago Vidal
 * Date: 18/10/17
 * Time: 12:49
 */

const Key = "job:"

type CreationInfo struct {
	Mode      string `json:"mode"`
	Status    string `json:"status"`
	JobID     string `json:"jobID"`
	errStream io.Writer
	jobData   interface{}
}

func(ci *CreationInfo) SetJobData(data interface{}) {
	ci.jobData = data
}

func(ci *CreationInfo) Start(jobFunc func(jobData interface{}) error) {

	dbS := db.Redis.Session()
	defer dbS.Close()

	if err := dbS.Store(Key + ci.JobID, &ExecutionInfo{
		ID: ci.JobID,
		Status: "running",
	}); err != nil {
		panic(err)
	}
	dbS.Close()

	go func(jobInfo *CreationInfo) {

		defer func() { //PANIC recover

			if r := recover(); r != nil {

				if jobInfo.errStream != nil {
					halt.PanicHandler(r, jobInfo.errStream)
				} else {
					log.Error(r)
				}

				if err, ok := r.(error); ok {
					ci.FinishJob(errors.New(err.Error()))
				} else {
					ci.FinishJob(errors.New(fmt.Sprintf("%v", r)))
				}

			}

		}()
		ci.FinishJob(jobFunc(ci.jobData))

	}(ci)
}

func(ci *CreationInfo) FinishJob(err error) {

	defer func() { //PANIC recover

		if r := recover(); r != nil {
			log.Error(r)
		}
	}()

	redisKey := Key + ci.JobID

	s := db.Redis.Session()
	defer s.Close()

	if err != nil {

		s.Store(redisKey, &ExecutionInfo{
			ID: ci.JobID,
			Status: "error",
			Error: err.Error(),
		})
		s.ExpiresAt(redisKey, time.Now().Add(time.Duration(30 * time.Minute)))

	} else {

		var je ExecutionInfo
		s.ReadEntity(redisKey, &je)

		je.Status = "finished"
		je.Progress.Percentage = 0
		je.Progress.Text = ""
		s.Store(redisKey, &je)
		s.ExpiresAt(redisKey, time.Now().Add(time.Duration(30 * time.Minute)))
	}
}

type ExecutionProgress struct {
	Percentage byte    `json:"percentage"`
	Text       string  `json:"text"`
}

type ExecutionInfo struct {
	db.EcoEntity
	ID       string            `json:"id"`
	Status   string            `json:"status"`
	Error    string            `json:"error"`
	Progress ExecutionProgress `json:"progress"`
	Result   string            `json:"result"`
}

func (j ExecutionInfo) RedisFillFields(ent interface{}, data *[]interface{}) {
	*data = append(*data, "status", j.Status)
	j.EcoEntity.RedisFillFields(ent, data)
}

/*****/

func New(errStream io.Writer) *CreationInfo {

	jobID := uuid.New()
	return &CreationInfo{
		Mode: "job",
		Status: "started",
		JobID: jobID,
		errStream: errStream,
	}

}
