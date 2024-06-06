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
const templateFileName = "template.typ"
const outputFileName = "report.pdf"

func main() {
	router := gin.Default()
	router.SetTrustedProxies(nil)
	router.GET("", getDefaultResponse)
	router.POST("render", render)
	router.Run("0.0.0.0:80")
}

func getDefaultResponse(c *gin.Context) {
	c.String(http.StatusOK, defaultResponse)
}

func render(c *gin.Context) {
	// Create a temporary directory
	dir, err := os.MkdirTemp("", "")
	if err != nil {
		panic(fmt.Sprintf("failed to create temporary directory: %v", err))
	}
	defer os.RemoveAll(dir)
	// Create a hard link of the template file inside the temporary directory
	err = os.Link(templateFileName, dir+"/"+templateFileName)
	if err != nil {
		panic(fmt.Sprintf("failed to link template file: %v", err))
	}
	// Write the received data to a JSON file inside the temporary directory
	jsonData, err := io.ReadAll(c.Request.Body)
	if err != nil {
		panic(fmt.Sprintf("failed to read request body: %v", err))
	}
	dataFile, err := os.Create(dir + "/data.json")
	if err != nil {
		panic(fmt.Sprintf("failed to create data file: %v", err))
	}
	defer dataFile.Close()
	_, err = dataFile.Write(jsonData)
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
	if err != nil {
		c.String(http.StatusUnprocessableEntity, "Failed to compile template with the given data.\n\n"+string(output))
		return
	}
	// Return the compiled file
	c.FileAttachment(dir+"/"+outputFileName, outputFileName)
}
