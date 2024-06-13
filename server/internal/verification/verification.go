package verification

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"lath/xman/internal/db"
	"lath/xman/internal/errors"
	"lath/xman/internal/tasks"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"strings"
	"time"
)

func init() {
	tasks.RegisterTaskHandler(
		db.ProcessStepFormatVerification,
		initVerificationHandler,
		tasks.Options{ConcurrentItems: 10, SafeRepeat: true},
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
	message      db.Message
	invalidFiles []string
}

// HandleItem runs format verification on the given document using the
// remote BORG service and saves the result to the document object.
func (h *VerificationHandler) HandleItem(ctx context.Context, itemData interface{}) error {
	filename := itemData.(string)
	filePath := path.Join(h.message.StoreDir, filename)
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
	request, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body.Bytes()))
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
		h.message.MessageHead.ProcessID,
		filename,
		&parsedResponse,
	)
	if isInvalid(parsedResponse) || hasError(parsedResponse) {
		h.invalidFiles = append(h.invalidFiles, filename)
	}
	return nil
}

func (h *VerificationHandler) Finish() {}

func (h *VerificationHandler) AfterDone() {
	if len(h.invalidFiles) > 0 {
		errors.AddProcessingError(db.ProcessingError{
			Title:       "Die Formatverifikation hat Probleme mit Primärdateien festgestellt",
			Info:        strings.Join(h.invalidFiles, "\n"),
			ProcessID:   h.message.MessageHead.ProcessID,
			MessageType: h.message.MessageType,
			ProcessStep: db.ProcessStepFormatVerification,
		})
	}
}

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

func isInvalid(f db.FormatVerification) bool {
	valid, ok := f.Summary["valid"]
	return ok && valid.Values[0].Value == "false" &&
		valid.Values[0].Score > 0.75
}

func hasError(f db.FormatVerification) bool {
	for _, r := range f.FileIdentificationResults {
		if r.Error != "" {
			return true
		}
	}
	for _, r := range f.FileValidationResults {
		if r.Error != "" {
			return true
		}
	}
	return false
}
