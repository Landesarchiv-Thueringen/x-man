package verification

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"lath/xman/internal/db"
	"lath/xman/internal/tasks"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"time"
)

func init() {
	tasks.RegisterTaskHandler(
		db.ProcessStepFormatVerification,
		initVerificationHandler,
		tasks.Options{ConcurrentItems: 3, RetrySafe: true},
	)
}

var borgURL = os.Getenv("BORG_URL")
var tr http.Transport = http.Transport{}
var client http.Client = http.Client{
	Timeout:   time.Second * 60,
	Transport: &tr,
}

func VerifyFileFormats(process db.SubmissionProcess, message db.Message) {
	rootRecords := db.FindRootRecords(context.Background(), process.ProcessID, db.MessageType0503)
	var items []db.TaskItem
	for _, d := range db.GetPrimaryDocuments(&rootRecords) {
		items = append(items, db.TaskItem{
			Label: d.Filename,
			State: db.TaskStatePending,
			Data:  d.Filename,
		})
	}
	task := db.Task{
		ProcessID: process.ProcessID,
		Type:      db.ProcessStepFormatVerification,
		Items:     items,
	}
	task = db.InsertTask(task)
	tasks.Run(&task)
}

type VerificationHandler struct {
	message db.Message
}

func (h *VerificationHandler) HandleItem(itemData interface{}) error {
	return verifyDocument(h.message, itemData.(string))
}
func (h *VerificationHandler) Finish()    {}
func (h *VerificationHandler) AfterDone() {}

func initVerificationHandler(t *db.Task) (tasks.ItemHandler, error) {
	message, ok := db.FindMessage(context.Background(), t.ProcessID, db.MessageType0503)
	if !ok {
		return nil, fmt.Errorf("failed to find 0503 message for process %v", t.ProcessID)
	}
	resp, err := http.Head(borgURL)
	if err != nil {
		return nil, fmt.Errorf("failed to reach BORG: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to reach BORG: HEAD \"%s\": %d", borgURL, resp.StatusCode)
	}
	return &VerificationHandler{message: message}, nil
}

// verifyDocument runs format verification on the given document using the
// remote BORG service and saves the result to the document object.
func verifyDocument(message db.Message, filename string) error {
	filePath := path.Join(message.StoreDir, filename)
	_, err := os.Stat(filePath)
	if err != nil {
		return err
	}
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	fw, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return err
	}
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	_, err = io.Copy(fw, file)
	if err != nil {
		return err
	}
	writer.Close()
	url := borgURL + "/analyze-file"
	request, err := http.NewRequest("POST", url, bytes.NewReader(body.Bytes()))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", writer.FormDataContentType())
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("POST \"%s\": %d", url, response.StatusCode)
	}
	var parsedResponse db.FormatVerification
	err = json.NewDecoder(response.Body).Decode(&parsedResponse)
	if err != nil {
		return err
	}
	db.UpdatePrimaryDocumentFormatVerification(
		message.MessageHead.ProcessID,
		filename,
		&parsedResponse,
	)
	return nil
}
