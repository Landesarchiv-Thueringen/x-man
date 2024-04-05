package db

import (
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var db *gorm.DB

func Init() {
	dsn := `host=database
		user=` + os.Getenv("POSTGRES_USER") + `
		password=` + os.Getenv("POSTGRES_PASSWORD") + `
		dbname=` + os.Getenv("POSTGRES_DB") + `
		port=5432
		sslmode=disable 
		TimeZone=Europe/Berlin`

	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(fmt.Sprintf("failed to connect to database: %v", err))
	}
	db = database
	db.AutoMigrate(&ServerState{})
}

// GetXManVersion returns the x-man version that the database was migrated to.
//
// Returns (0,0,0) when starting x-man with a fresh database.
func GetXManVersion() (uint, uint, uint) {
	var serverState ServerState
	result := db.Limit(1).Find(&serverState)
	if result.Error != nil {
		panic(result.Error)
	}
	return serverState.XManMajorVersion, serverState.XManMinorVersion, serverState.XManPatchVersion
}

func SetXManVersion(major, minor, patch uint) {
	var serverState ServerState
	result := db.Limit(1).Find(&serverState)
	if result.Error != nil {
		panic(result.Error)
	}
	serverState.XManMajorVersion = major
	serverState.XManMinorVersion = minor
	serverState.XManPatchVersion = patch
	result = db.Save(&serverState)
	if result.Error != nil {
		panic(result.Error)
	}
}

// Migrate migrates all database tables and relations.
func Migrate() {
	if db == nil {
		panic("database wasn't initialized")
	}
	// Migrate the complete schema.
	err := db.AutoMigrate(
		&Agency{},
		&Appraisal{},
		&XdomeaVersion{},
		&Process{},
		&ProcessState{},
		&ProcessStep{},
		&Message{},
		&MessageType{},
		&MessageHead{},
		&Contact{},
		&AgencyIdentification{},
		&Institution{},
		&FileRecordObject{},
		&ProcessRecordObject{},
		&DocumentRecordObject{},
		&GeneralMetadata{},
		&FilePlan{},
		&Lifetime{},
		&ArchiveMetadata{},
		&RecordObjectAppraisal{},
		&ConfidentialityLevel{},
		&Version{},
		&Format{},
		&PrimaryDocument{},
		&FormatVerification{},
		&ToolResponse{},
		&ExtractedFeature{},
		&Feature{},
		&FeatureValue{},
		&ToolConfidence{},
		&ProcessingError{},
		&Collection{},
		&Task{},
		&User{},
		&UserPreferences{},
	)
	if err != nil {
		panic(fmt.Sprintf("failed to migrate database: %v", err))
	}
}

func InitMessageTypes(messageTypes []*MessageType) {
	result := db.Create(messageTypes)
	if result.Error != nil {
		panic(fmt.Sprintf("failed to initialize message types: %v", result.Error))
	}
}

func InitXdomeaVersions(versions []*XdomeaVersion) {
	result := db.Create(versions)
	if result.Error != nil {
		panic(fmt.Sprintf("failed to initialize xdomea versions: %v", result.Error))
	}
}

func InitRecordObjectAppraisals(appraisals []*RecordObjectAppraisal) {
	result := db.Create(appraisals)
	if result.Error != nil {
		panic(fmt.Sprintf("failed to initialize record object appraisal values: %v", result.Error))
	}
}

func InitConfidentialityLevelCodelist(codelist []*ConfidentialityLevel) {
	result := db.Create(codelist)
	if result.Error != nil {
		panic(fmt.Sprintf("failed to initialize confidentiality level codelist: %v", result.Error))
	}
}

func InitMediumCodelist(mediumCodelist []*Medium) {
	result := db.Create(mediumCodelist)
	if result.Error != nil {
		panic(fmt.Sprintf("failed to initialize medium code list: %v", result.Error))
	}
}

func InitAgencies(agencies []Agency) {
	result := db.Create(agencies)
	if result.Error != nil {
		panic(fmt.Sprintf("failed to initialize agency configuration: %v", result.Error))
	}
}

func AddProcess(
	agency Agency,
	processID string,
	processStoreDir string,
	institution *string,
) Process {
	var process Process
	processState := AddProcessState()
	process = Process{
		Agency:       agency,
		ID:           processID,
		StoreDir:     processStoreDir,
		Institution:  institution,
		ProcessState: processState,
	}
	result := db.Save(&process)
	if result.Error != nil {
		panic(result.Error)
	}
	return process
}

// DeleteProcess deletes the given process and all its associations.
func DeleteProcess(id string) {
	if id == "" {
		panic("called DeleteProcess with empty string")
	}
	// Note that we don't use inline (`Delete(&Process{}, id)`) or explicit
	// (`Where("...")`) conditions. `BeforeDelete` and `AfterDelete` hooks only
	// see the primary value that was passed to `Delete`. If we don't include
	// the ID in this value, we cannot delete associations using these hooks.
	result := db.Delete(&Process{ID: id})
	if result.Error != nil {
		panic(result.Error)
	} else if result.RowsAffected != 1 {
		panic(fmt.Sprintf("failed to delete process %v: not found", id))
	}
}

// DeleteMessage deletes the given message and all its associations.
//
// Panics if the message cannot be found.
func DeleteMessage(message Message) {
	if message.ID == uuid.Nil {
		panic("called DeleteMessage with nil ID")
	}
	processID := message.MessageHead.ProcessID
	process, found := GetProcess(processID)
	if !found {
		panic("process not found: " + processID)
	}
	if process.Message0501ID != nil && *process.Message0501ID == message.ID {
		process.Message0501ID = nil
		process.ProcessState.Receive0501.CompletionTime = nil
		process.ProcessState.Receive0501.Complete = false
		UpdateProcessStep(process.ProcessState.Receive0501)
	} else if process.Message0503ID != nil && *process.Message0503ID == message.ID {
		process.Message0503ID = nil
		process.ProcessState.Receive0503.CompletionTime = nil
		process.ProcessState.Receive0503.Complete = false
		UpdateProcessStep(process.ProcessState.Receive0503)
	} else if process.Message0505ID != nil && *process.Message0505ID == message.ID {
		process.Message0505ID = nil
		process.ProcessState.Receive0505.CompletionTime = nil
		process.ProcessState.Receive0505.Complete = false
		UpdateProcessStep(process.ProcessState.Receive0505)
	} else {
		panic(fmt.Errorf("could not find message reference of message %v in process %v",
			message.ID, process.ID))
	}
	result := db.Delete(&message)
	if result.Error != nil {
		panic(result.Error)
	} else if result.RowsAffected != 1 {
		panic(fmt.Sprintf("failed to delete message %v: not found", message.ID))
	}
}

func SetProcessNote(
	process Process,
	note string,
) {
	process.Note = &note
	result := db.Save(&process)
	if result.Error != nil {
		panic(result.Error)
	}
}

func AddProcessState() ProcessState {
	var processState ProcessState
	Receive0501 := ProcessStep{}
	result := db.Save(&Receive0501)
	if result.Error != nil {
		panic(result.Error)
	}
	Appraisal := ProcessStep{}
	result = db.Save(&Appraisal)
	if result.Error != nil {
		panic(result.Error)
	}
	Receive0505 := ProcessStep{}
	result = db.Save(&Receive0505)
	if result.Error != nil {
		panic(result.Error)
	}
	Receive0503 := ProcessStep{}
	result = db.Save(&Receive0503)
	if result.Error != nil {
		panic(result.Error)
	}
	FormatVerification := ProcessStep{}
	result = db.Save(&FormatVerification)
	if result.Error != nil {
		panic(result.Error)
	}
	Archiving := ProcessStep{}
	result = db.Save(&Archiving)
	if result.Error != nil {
		panic(result.Error)
	}
	processState = ProcessState{
		Receive0501:        Receive0501,
		Appraisal:          Appraisal,
		Receive0505:        Receive0505,
		Receive0503:        Receive0503,
		FormatVerification: FormatVerification,
		Archiving:          Archiving,
	}
	result = db.Save(&processState)
	if result.Error != nil {
		panic(result.Error)
	}
	return processState
}

func AddMessage(
	agency Agency,
	processID string,
	processStoreDir string,
	message Message,
) (Process, Message, error) {
	var process Process
	// generate ID for message, propagate the ID to record object children
	// must be done before saving the message in database
	message.ID = uuid.New()
	setRecordObjectsMessageID(&message)
	result := db.Create(&message)
	// The Database failed to create the message.
	if result.Error != nil {
		return process, message, result.Error
	}
	process, found := GetProcess(processID)
	// The process was not found. Create a new process.
	if !found {
		var institution *string
		// set institution if possible
		if message.MessageHead.Sender.Institution != nil {
			institution = message.MessageHead.Sender.Institution.Name
		}
		process = AddProcess(agency, processID, processStoreDir, institution)
	} else {
		// Check if the process has already a message with the type of the given message.
		_, found := GetMessageOfProcessByCode(process, message.MessageType.Code)
		if found {
			panic("process already has message with type " + message.MessageType.Code)
		}
	}
	switch message.MessageType.Code {
	case "0501":
		process.Message0501 = &message
		processStep := process.ProcessState.Receive0501
		processStep.Complete = true
		completionTime := time.Now()
		processStep.CompletionTime = &completionTime
		UpdateProcessStep(processStep)
	case "0503":
		process.Message0503 = &message
		processStep := process.ProcessState.Receive0503
		processStep.Complete = true
		completionTime := time.Now()
		processStep.CompletionTime = &completionTime
		UpdateProcessStep(processStep)
	case "0505":
		process.Message0505 = &message
		processStep := process.ProcessState.Receive0505
		processStep.Complete = true
		completionTime := time.Now()
		processStep.CompletionTime = &completionTime
		UpdateProcessStep(processStep)
	default:
		panic("unhandled message type: " + message.MessageType.Code)
	}
	result = db.Save(&process)
	return process, message, result.Error
}

// setRecordObjectsMessageID sets the message ID for all record objects of the message.
// This information helps to retrieve the message if only the record object is known.
func setRecordObjectsMessageID(message *Message) {
	for _, recordObject := range message.GetRecordObjects() {
		recordObject.SetMessageID(message.ID)
		for _, childRecordObject := range recordObject.GetChildren() {
			childRecordObject.SetMessageID(message.ID)
		}
	}
}

func UpdateProcess(process Process) {
	result := db.Save(&process)
	if result.Error != nil {
		panic(result.Error)
	}
}

func UpdateMessage(message Message) {
	result := db.Save(&message)
	if result.Error != nil {
		panic(result.Error)
	}
}

func UpdatePrimaryDocument(primaryDocument PrimaryDocument) {
	result := db.Save(&primaryDocument)
	if result.Error != nil {
		panic(result.Error)
	}
}

func UpdateProcessStep(processStep ProcessStep) {
	result := db.Save(&processStep)
	if result.Error != nil {
		panic(result.Error)
	}
}

func GetAppraisableRecordObject(messageID uuid.UUID, recordObjectID uuid.UUID) AppraisableRecordObject {
	if messageID == uuid.Nil {
		panic("called GetAppraisableRecordObject with nil messageID")
	} else if recordObjectID == uuid.Nil {
		panic("called GetAppraisableRecordObject with nil recordObjectID")
	}
	fileRecordObject := FileRecordObject{MessageID: messageID, XdomeaID: recordObjectID}
	result := db.Limit(1).Where(&fileRecordObject).Find(&fileRecordObject)
	if result.Error != nil {
		panic(result.Error)
	} else if result.RowsAffected > 0 {
		return &fileRecordObject
	}
	processRecordObject := ProcessRecordObject{MessageID: messageID, XdomeaID: recordObjectID}
	result = db.Limit(1).Where(&processRecordObject).Find(&processRecordObject)
	if result.Error != nil {
		panic(result.Error)
	} else if result.RowsAffected > 0 {
		return &processRecordObject
	}
	return nil
}

func GetAppraisalsForProcess(processID string) (appraisals []Appraisal) {
	if processID == "" {
		panic("called GetAppraisalsForProcess with empty processID")
	}
	result := db.Where(&Appraisal{ProcessID: processID}).Find(&appraisals)
	if result.Error != nil {
		panic(result.Error)
	}
	return
}

func GetAppraisal(processID string, recordObjectID uuid.UUID) (a Appraisal) {
	if processID == "" {
		panic("called GetAppraisal with empty processID")
	} else if recordObjectID == uuid.Nil {
		panic("called GetAppraisal with nil recordObjectID")
	}
	a.ProcessID = processID
	a.RecordObjectID = recordObjectID
	result := db.Limit(1).Where(&a).Find(&a)
	if result.Error != nil {
		panic(result.Error)
	}
	return
}

func SetAppraisal(processID string, recordObjectID uuid.UUID, decision AppraisalDecisionOption, internalNote string) {
	patchAppraisal(processID, recordObjectID, &decision, &internalNote)
}

func SetAppraisalDecision(processID string, recordObjectID uuid.UUID, decision AppraisalDecisionOption) {
	patchAppraisal(processID, recordObjectID, &decision, nil)
}

func SetAppraisalInternalNote(processID string, recordObjectID uuid.UUID, internalNote string) {
	patchAppraisal(processID, recordObjectID, nil, &internalNote)
}

func patchAppraisal(processID string, recordObjectID uuid.UUID, decision *AppraisalDecisionOption, internalNote *string) {
	if processID == "" {
		panic("called SetAppraisal with empty processID")
	} else if recordObjectID == uuid.Nil {
		panic("called SetAppraisal with nil recordObjectID")
	}
	appraisal := Appraisal{ProcessID: processID, RecordObjectID: recordObjectID}
	result := db.Limit(1).Where(&appraisal).Find(&appraisal)
	if result.Error != nil {
		panic(result.Error)
	}
	if decision != nil {
		appraisal.Decision = *decision
	}
	if internalNote != nil {
		appraisal.InternalNote = *internalNote
	}
	result = db.Save(&appraisal)
	if result.Error != nil {
		panic(result.Error)
	}
}

func UpdateAppraisal(a Appraisal) {
	if a.ProcessID == "" {
		panic("called SetAppraisal with empty processID")
	} else if a.RecordObjectID == uuid.Nil {
		panic("called SetAppraisal with nil recordObjectID")
	}
	result := db.Save(&a)
	if result.Error != nil {
		panic(result.Error)
	}
}

// AddProcessingError saves a processing error to the database.
//
// Do not call directly. Instead use clearing.HandleError.
func AddProcessingError(e ProcessingError) {
	result := db.Create(&e)
	if result.Error != nil {
		panic(result.Error)
	}
}

func GetProcessingError(id uint) (ProcessingError, bool) {
	if id == 0 {
		panic("called GetProcessingError with ID 0")
	}
	processingError := ProcessingError{ID: id}
	result := db.Preload(clause.Associations).Limit(1).Find(&processingError)
	if result.Error != nil {
		panic(result.Error)
	}
	return processingError, result.RowsAffected > 0
}

func UpdateProcessingError(processingError ProcessingError) {
	if processingError.ID == 0 {
		panic("called UpdateProcessingError with ID 0")
	}
	result := db.Save(&processingError)
	if result.Error != nil {
		panic(result.Error)
	}
}

func CreateAgency(agency Agency) (uint, error) {
	result := db.Create(&agency)
	return agency.ID, result.Error
}

func UpdateAgency(id uint, agency Agency) error {
	agency.ID = id
	err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&agency).Association("Users").Replace(agency.Users); err != nil {
			return err
		}
		tx.Save(&agency)
		return nil
	})
	return err
}

func DeleteAgency(id uint) bool {
	if id == 0 {
		panic("called DeleteAgency with ID 0")
	}
	result := db.Delete(&Agency{}, id)
	if result.Error != nil {
		panic(result.Error)
	}
	return result.RowsAffected == 1
}

func CreateCollection(Collection Collection) (uint, error) {
	result := db.Create(&Collection)
	return Collection.ID, result.Error
}

func UpdateCollection(id uint, collection Collection) error {
	if id == 0 {
		panic("called UpdateCollection with ID 0")
	}
	collection.ID = id
	result := db.Save(&collection)
	return result.Error
}

func DeleteCollection(id uint) bool {
	if id == 0 {
		panic("called DeleteCollection with ID 0")
	}
	result := db.Delete(&Collection{}, id)
	if result.Error != nil {
		panic(result.Error)
	}
	return result.RowsAffected == 1
}

func CreateTask(task Task) Task {
	result := db.Create(&task)
	if result.Error != nil {
		panic(result.Error)
	}
	return task
}

func UpdateTask(task Task) {
	if task.ID == 0 {
		panic("called UpdateTask with ID 0")
	}
	result := db.Save(&task)
	if result.Error != nil {
		panic(result.Error)
	}
}

func DeleteTask(task Task) {
	if task.ID == 0 {
		panic("called DeleteTask with ID 0")
	}
	result := db.Delete(&task)
	if result.Error != nil {
		panic(result.Error)
	} else if result.RowsAffected != 1 {
		panic(fmt.Sprintf("failed to delete task %v: not found", task.ID))
	}
}

// UpdateUserPreferences saves the preferences for the given user to the
// database.
//
// Both the entry for the user and the entry for the user preferences are
// created if they don't yet exist.
func UpdateUserPreferences(userID string, userPreferences UserPreferences) {
	if len(userID) == 0 {
		panic("called GetUserSettings with empty ID")
	}
	userPreferences.UserID = userID
	err := db.Transaction(func(tx *gorm.DB) error {
		tx.Save(&User{ID: userID})
		tx.Save(&userPreferences)
		return nil
	})
	if err != nil {
		panic(err)
	}
}
