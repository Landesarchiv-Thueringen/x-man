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
	"regexp"
	"strconv"
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

func VerifyFileFormats(process db.SubmissionProcess, primaryDocuments []db.PrimaryDocumentData) {
	var items []db.TaskItem
	for _, d := range primaryDocuments {
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
func (h *VerificationHandler) HandleItem(
	ctx context.Context,
	itemData interface{},
	updateItemData func(data interface{}),
) error {
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
	url := borgURL + "/api/analyze-file"
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
	if parsedResponse.Summary.Invalid || parsedResponse.Summary.Error {
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
	err := TestConnection()
	if err != nil {
		return nil, fmt.Errorf("failed to reach BORG: %w", err)
	}

	return &VerificationHandler{message: message}, nil
}

func TestConnection() error {
	url := borgURL + "/api/version"
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("GET \"%s\": %d", url, resp.StatusCode)
	}
	contentType := resp.Header.Get("content-type")
	if !strings.HasPrefix(contentType, "text/plain") {
		return fmt.Errorf("GET \"%s\": expected content type text/plain, got: %s", url, contentType)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	return checkBorgVersion(string(body))
}

// checkBorgVersion compares the given Borg version to what is compatible with
// x-man and returns an error if it is not.
func checkBorgVersion(version string) error {
	r := regexp.MustCompile(`^([0-9])+\.([0-9])+\.([0-9])+$`)
	m := r.FindStringSubmatch(version)
	if len(m) != 4 {
		return fmt.Errorf("failed to parse version: %s", version)
	}
	major, err := strconv.Atoi(m[1])
	if err != nil {
		return fmt.Errorf("failed to parse version: %s", version)
	}
	minor, err := strconv.Atoi(m[2])
	if err != nil {
		return fmt.Errorf("failed to parse version: %s", version)
	}
	if major < 1 || major == 1 && minor < 4 { // >= 1.4.0
		return fmt.Errorf("require version >= 1.4.0, has: %s", version)
	}
	return nil
}
