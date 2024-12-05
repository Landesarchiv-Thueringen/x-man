package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"

	"github.com/gin-gonic/gin"
)

const defaultResponse = "X-Man report server is running"

func main() {
	router := gin.Default()
	router.SetTrustedProxies(nil)
	router.GET("", getDefaultResponse)
	router.POST("render/appraisal", renderAppraisalReport)
	router.POST("render/submission", renderSubmissionReport)
	router.Run()
}

func getDefaultResponse(c *gin.Context) {
	c.String(http.StatusOK, defaultResponse)
}

func renderAppraisalReport(c *gin.Context) {
	data, err := io.ReadAll(c.Request.Body)
	if err != nil {
		panic(fmt.Sprintf("failed to read request body: %v", err))
	}
	render(
		"appraisal-report.typ",
		"appraisal-data.json",
		data,
		func(path string, output string, err error) {
			if err != nil {
				c.String(
					http.StatusUnprocessableEntity,
					"Failed to compile template with the given data.\n\n"+string(output),
				)
				return
			}
			// Return the compiled file
			c.FileAttachment(path, "appraisal-report.pdf")
		},
	)
}

func renderSubmissionReport(c *gin.Context) {
	jsonData, err := io.ReadAll(c.Request.Body)
	if err != nil {
		panic(fmt.Sprintf("failed to read request body: %v", err))
	}
	render(
		"submission-report.typ",
		"submission-data.json",
		jsonData,
		func(path string, output string, err error) {
			if err != nil {
				c.String(
					http.StatusUnprocessableEntity,
					"Failed to compile template with the given data.\n\n"+string(output),
				)
				return
			}
			// Return the compiled file
			c.FileAttachment(path, "submission-report.pdf")
		},
	)
}

func render(
	templateFileName string,
	dataFileName string,
	data []byte,
	withResult func(path string, output string, err error),
) {
	const outputFileName = "report.pdf"
	// Create a temporary directory
	dir, err := os.MkdirTemp("", "")
	if err != nil {
		panic(fmt.Sprintf("failed to create temporary directory: %v", err))
	}
	defer os.RemoveAll(dir)
	// Create a hard link of the template file inside the temporary directory
	for _, filename := range []string{templateFileName, "shared.typ"} {
		err = os.Link(filename, dir+"/"+filename)
		if err != nil {
			panic(fmt.Sprintf("failed to link %s: %v", filename, err))
		}
	}
	// Write the received data to a JSON file inside the temporary directory
	dataFile, err := os.Create(dir + "/" + dataFileName)
	if err != nil {
		panic(fmt.Sprintf("failed to create data file: %v", err))
	}
	defer dataFile.Close()
	_, err = dataFile.Write(data)
	if err != nil {
		panic(fmt.Sprintf("failed to write data to file: %v", err))
	}
	// Compile Typst template with received data
	cmd := exec.Command(
		"typst",
		"compile",
		templateFileName,
		outputFileName,
	)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	withResult(dir+"/"+outputFileName, string(output), err)
}
