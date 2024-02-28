package report

import "lath/xman/internal/db"

type ContentStats struct {
	Files        uint
	SubFiles     uint
	Processes    uint
	SubProcesses uint
	Documents    uint
	Attachments  uint
}

func (c *ContentStats) processFiles(files []db.FileRecordObject, isSubLevel bool) {
	for _, file := range files {
		if isSubLevel {
			c.SubFiles += 1
		} else {
			c.Files += 1
		}
		c.processFiles(file.SubFileRecordObjects, true)
		c.processProcesses(file.ProcessRecordObjects, false)
		c.processDocuments(file.DocumentRecordObjects, false)
	}
}

func (c *ContentStats) processProcesses(processes []db.ProcessRecordObject, isSubLevel bool) {
	for _, process := range processes {
		if isSubLevel {
			c.SubProcesses += 1
		} else {
			c.Processes += 1
		}
		c.processProcesses(process.SubProcessRecordObjects, false)
		c.processDocuments(process.DocumentRecordObjects, false)
	}
}

func (c *ContentStats) processDocuments(documents []db.DocumentRecordObject, isSubLevel bool) {
	for _, document := range documents {
		if isSubLevel {
			c.Attachments += 1
		} else {
			c.Documents += 1
		}
		c.processDocuments(document.Attachments, true)
	}
}

func getMessageContentStats(message db.Message) (c ContentStats) {
	c.processFiles(message.FileRecordObjects, false)
	c.processProcesses(message.ProcessRecordObjects, false)
	c.processDocuments(message.DocumentRecordObjects, false)
	return
}

func getFileContentStats(file db.FileRecordObject) (c ContentStats) {
	c.processFiles(file.SubFileRecordObjects, true)
	c.processProcesses(file.ProcessRecordObjects, false)
	c.processDocuments(file.DocumentRecordObjects, false)
	return
}

func getProcessContentStats(process db.ProcessRecordObject) (c ContentStats) {
	c.processProcesses(process.SubProcessRecordObjects, false)
	c.processDocuments(process.DocumentRecordObjects, false)
	return
}
