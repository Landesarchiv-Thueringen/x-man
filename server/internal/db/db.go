package db

import (
	"errors"
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

// GetXManVersion returns the XMan version that the database was migrated to.
func GetXManVersion() (uint, error) {
	var serverState ServerState
	result := db.Limit(1).Find(&serverState)
	return serverState.XManVersion, result.Error
}

func SetXManVersion(version uint) error {
	var serverState ServerState
	result := db.Limit(1).Find(&serverState)
	if result.Error != nil {
		return result.Error
	}
	serverState.XManVersion = version
	result = db.Save(&serverState)
	return result.Error
}

// Migrate migrates all database tables and relations.
func Migrate() {
	if db == nil {
		panic("database wasn't initialized")
	}
	// Migrate the complete schema.
	err := db.AutoMigrate(
		&Agency{},
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
	xdomeaID string,
	processStoreDir string,
	institution *string,
) (Process, error) {
	var process Process
	processState, err := AddProcessState()
	if err != nil {
		return process, err
	}
	process = Process{
		Agency:       agency,
		XdomeaID:     xdomeaID,
		StoreDir:     processStoreDir,
		Institution:  institution,
		ProcessState: processState,
	}
	result := db.Save(&process)
	return process, result.Error
}

// DeleteProcess deletes the given process and all its associations.
func DeleteProcess(id uuid.UUID) (bool, error) {
	// Note that we don't use inline (`Delete(&Process{}, id)`) or explicit
	// (`Where("...")`) conditions. `BeforeDelete` and `AfterDelete` hooks only
	// see the primary value that was passed to `Delete`. If we don't include
	// the ID in this value, we cannot delete associations using these hooks.
	result := db.Delete(&Process{ID: id})
	return result.RowsAffected == 1, result.Error
}

// DeleteMessage deletes the given message and all its associations.
//
// It dereferences but keeps the process.
func DeleteMessage(message Message) (bool, error) {
	processID := message.MessageHead.ProcessID
	process, err := GetProcessByXdomeaID(processID)
	if err != nil {
		return false, err
	}
	if process.Message0501ID != nil && *process.Message0501ID == message.ID {
		process.Message0501ID = nil
		process.ProcessState.Receive0501.CompletionTime = nil
		process.ProcessState.Receive0501.Complete = false
		err = UpdateProcessStep(process.ProcessState.Receive0501)
	} else if process.Message0503ID != nil && *process.Message0503ID == message.ID {
		process.Message0503ID = nil
		process.ProcessState.Receive0503.CompletionTime = nil
		process.ProcessState.Receive0503.Complete = false
		err = UpdateProcessStep(process.ProcessState.Receive0503)
	} else if process.Message0505ID != nil && *process.Message0505ID == message.ID {
		process.Message0505ID = nil
		process.ProcessState.Receive0505.CompletionTime = nil
		process.ProcessState.Receive0505.Complete = false
		err = UpdateProcessStep(process.ProcessState.Receive0505)
	} else {
		return false,
			fmt.Errorf("could not find message reference of message %v in process %v",
				message.ID, process.ID)
	}
	if err != nil {
		return false, err
	}
	result := db.Delete(&message)
	if result.RowsAffected == 0 || result.Error != nil {
		return result.RowsAffected == 1, result.Error
	}
	return result.RowsAffected == 1, err
}

func SetProcessNote(
	xdomeaID string,
	note string,
) error {
	process := Process{XdomeaID: xdomeaID}
	result := db.Model(&Process{}).Where(&process).Limit(1).Find(&process)
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	process.Note = &note
	result = db.Save(&process)
	return result.Error
}

func AddProcessState() (ProcessState, error) {
	var processState ProcessState
	Receive0501 := ProcessStep{}
	result := db.Save(&Receive0501)
	if result.Error != nil {
		return processState, result.Error
	}
	Appraisal := ProcessStep{}
	result = db.Save(&Appraisal)
	if result.Error != nil {
		return processState, result.Error
	}
	Receive0505 := ProcessStep{}
	result = db.Save(&Receive0505)
	if result.Error != nil {
		return processState, result.Error
	}
	Receive0503 := ProcessStep{}
	result = db.Save(&Receive0503)
	if result.Error != nil {
		return processState, result.Error
	}
	FormatVerification := ProcessStep{}
	result = db.Save(&FormatVerification)
	if result.Error != nil {
		return processState, result.Error
	}
	Archiving := ProcessStep{}
	result = db.Save(&Archiving)
	if result.Error != nil {
		return processState, result.Error
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
	return processState, result.Error
}

func AddMessage(
	agency Agency,
	xdomeaID string,
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
	process, err := GetProcessByXdomeaID(xdomeaID)
	// The process was not found. Create a new process.
	if err != nil {
		var institution *string
		// set institution if possible
		if message.MessageHead.Sender.Institution != nil {
			institution = message.MessageHead.Sender.Institution.Name
		}
		process, err = AddProcess(agency, xdomeaID, processStoreDir, institution)
		// The Database failed to create the process for the message.
		if err != nil {
			return process, message, err
		}
	} else {
		// Check if the process has already a message with the type of the given message.
		_, err = GetMessageOfProcessByCode(process, message.MessageType.Code)
		if err == nil {
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
		err = UpdateProcessStep(processStep)
		if err != nil {
			panic(err)
		}
	case "0503":
		process.Message0503 = &message
		processStep := process.ProcessState.Receive0503
		processStep.Complete = true
		completionTime := time.Now()
		processStep.CompletionTime = &completionTime
		err = UpdateProcessStep(processStep)
		if err != nil {
			panic(err)
		}
	case "0505":
		process.Message0505 = &message
		processStep := process.ProcessState.Receive0505
		processStep.Complete = true
		completionTime := time.Now()
		processStep.CompletionTime = &completionTime
		err = UpdateProcessStep(processStep)
		if err != nil {
			panic(err)
		}
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

func UpdateProcess(process Process) error {
	result := db.Save(&process)
	return result.Error
}

func UpdateMessage(message Message) error {
	result := db.Save(&message)
	return result.Error
}

func UpdatePrimaryDocument(primaryDocument PrimaryDocument) error {
	result := db.Save(&primaryDocument)
	return result.Error
}

func UpdateProcessStep(processStep ProcessStep) error {
	result := db.Save(&processStep)
	return result.Error
}

func SetFileRecordObjectAppraisal(
	id uuid.UUID,
	appraisalCode string,
	recursive bool,
) (FileRecordObject, error) {
	fileRecordObject, err := GetFileRecordObjectByID(id)
	if err != nil {
		return fileRecordObject, err
	}
	// check if message appraisal is already completed, if true return error
	message, err := GetCompleteMessageByID(fileRecordObject.MessageID)
	if err != nil {
		return fileRecordObject, err
	}
	if message.AppraisalComplete {
		return fileRecordObject, errors.New("message appraisal already finished")
	}
	// set appraisal
	err = fileRecordObject.SetAppraisal(appraisalCode)
	if err != nil {
		return fileRecordObject, err
	}
	// set appraisal for child elements if recursive appraisal was chosen
	if recursive {
		for _, process := range fileRecordObject.ProcessRecordObjects {
			_, err = SetProcessRecordObjectAppraisal(process.ID, appraisalCode)
			if err != nil {
				return fileRecordObject, err
			}
		}
	}
	// return updated file record object
	return GetFileRecordObjectByID(id)
}

func SetFileRecordObjectAppraisalNote(
	id uuid.UUID,
	appraisalNote string,
) (FileRecordObject, error) {
	fileRecordObject, err := GetFileRecordObjectByID(id)
	if err != nil {
		return fileRecordObject, err
	}
	// check if message appraisal is already completed, if true return error
	message, err := GetCompleteMessageByID(fileRecordObject.MessageID)
	if err != nil {
		return fileRecordObject, err
	}
	if message.AppraisalComplete {
		return fileRecordObject, errors.New("message appraisal already finished")
	}
	// set note
	err = fileRecordObject.SetAppraisalNote(appraisalNote)
	if err != nil {
		return fileRecordObject, err
	}
	// return updated file record object
	return GetFileRecordObjectByID(id)
}

func SetProcessRecordObjectAppraisal(
	id uuid.UUID,
	appraisalCode string,
) (ProcessRecordObject, error) {
	processRecordObject, err := GetProcessRecordObjectByID(id)
	if err != nil {
		return processRecordObject, err
	}
	// check if message appraisal is already completed, if true return error
	message, err := GetCompleteMessageByID(processRecordObject.MessageID)
	if err != nil {
		return processRecordObject, err
	}
	if message.AppraisalComplete {
		return processRecordObject, errors.New("message appraisal already finished")
	}
	// set appraisal
	err = processRecordObject.SetAppraisal(appraisalCode)
	if err != nil {
		return processRecordObject, err
	}
	// return updated process record object
	return GetProcessRecordObjectByID(id)
}

func SetProcessRecordObjectAppraisalNote(
	id uuid.UUID,
	appraisalNote string,
) (ProcessRecordObject, error) {
	processRecordObject, err := GetProcessRecordObjectByID(id)
	if err != nil {
		return processRecordObject, err
	}
	// check if message appraisal is already completed, if true return error
	message, err := GetCompleteMessageByID(processRecordObject.MessageID)
	if err != nil {
		return processRecordObject, err
	}
	if message.AppraisalComplete {
		return processRecordObject, errors.New("message appraisal already finished")
	}
	// set note
	err = processRecordObject.SetAppraisalNote(appraisalNote)
	if err != nil {
		return processRecordObject, err
	}
	// return updated process record object
	return GetProcessRecordObjectByID(id)
}

// AddProcessingError saves a processing error to the database.
//
// Do not call directly. Instead use CreateProcessingError.
func addProcessingError(e ProcessingError) {
	result := db.Save(&e)
	if result.Error != nil {
		panic(result.Error)
	}
}

// CreateProcessingError adds a new processing error to the database.
//
// It fills some missing fields if sufficient information is provided.
func CreateProcessingError(e ProcessingError) {
	if e.Process == nil && e.ProcessID != nil {
		process, err := GetProcess(*e.ProcessID)
		if err != nil {
			e.Process = &process
		}
	}
	if e.AgencyID == nil && e.Agency == nil {
		if e.Process != nil {
			e.AgencyID = &e.Process.AgencyID
			e.Agency = &e.Process.Agency
		}
	}
	if e.Message == nil && e.MessageID != nil {
		message, err := GetMessageByID(*e.MessageID)
		if err != nil {
			e.Message = &message
		}
	}
	if e.TransferPath == nil && e.Message != nil {
		e.TransferPath = &e.Message.TransferDirMessagePath
	}
	if e.Message != nil && e.Process != nil && e.ProcessStep == nil && e.ProcessStepID == nil {
		switch e.Message.MessageType.Code {
		case "0501":
			e.ProcessStep = &e.Process.ProcessState.Receive0501
		case "0503":
			e.ProcessStep = &e.Process.ProcessState.Receive0503
		case "0505":
			e.ProcessStep = &e.Process.ProcessState.Receive0505
		}
	}
	addProcessingError(e)
}

func GetProcessingError(id uint) (ProcessingError, error) {
	processingError := ProcessingError{ID: id}
	result := db.Preload(clause.Associations).First(&processingError)
	return processingError, result.Error
}

func UpdateProcessingError(processingError ProcessingError) error {
	result := db.Save(&processingError)
	return result.Error
}

func CreateAgency(agency Agency) (uint, error) {
	result := db.Create(&agency)
	return agency.ID, result.Error
}

func UpdateAgency(id uint, agency Agency) error {
	agency.ID = id
	result := db.Save(&agency)
	return result.Error
}

func DeleteAgency(id uint) (bool, error) {
	result := db.Delete(&Agency{}, id)
	return result.RowsAffected == 1, result.Error
}

func CreateCollection(Collection Collection) (uint, error) {
	result := db.Create(&Collection)
	return Collection.ID, result.Error
}

func UpdateCollection(id uint, collection Collection) error {
	collection.ID = id
	result := db.Save(&collection)
	return result.Error
}

func DeleteCollection(id uint) (bool, error) {
	result := db.Delete(&Collection{}, id)
	return result.RowsAffected == 1, result.Error
}

func CreateTask(task Task) (Task, error) {
	result := db.Create(&task)
	return task, result.Error
}

func UpdateTask(task Task) error {
	result := db.Save(&task)
	return result.Error
}

func DeleteTask(task Task) (bool, error) {
	result := db.Delete(&task)
	return result.RowsAffected == 1, result.Error
}
