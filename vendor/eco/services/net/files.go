package net

import (
	"net/http"
	"bytes"
	"mime/multipart"
	"io"
	"eco/services/halt"
	"encoding/json"
)

func NewFilesUploader() *filesUploader {
	return &filesUploader{}
}

type fileData struct {
	data io.Reader
	name string
}
type filesUploader struct {
	files []fileData
}
func (fu *filesUploader) AddFile(data io.Reader, fileName string) {
	fu.files = append(fu.files,
		fileData{
			data: data,
			name: fileName,
	})
}
func (fu *filesUploader) Post(apiUrl string, callback func(response *http.Response) error) error {

	//por ahora soporto un solo archivo
	fileName := fu.files[0].name
	data := fu.files[0].data

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", fileName)
	if err != nil {
		return err
	}
	_, err = io.Copy(part, data)

	err = writer.Close()
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", apiUrl, body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {

		type ResponseJson struct {
			Status int    `json:"status"`
			Error  string `json:"error"`
		}
		var responseData ResponseJson
		json.NewDecoder(resp.Body).Decode(&responseData)

		return halt.Errorf(responseData.Error, resp.StatusCode)
	}
	return callback(resp)
}