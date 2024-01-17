package main

import (
	"io"
	"log"
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

func getDefaultResponse(context *gin.Context) {
	context.String(http.StatusOK, defaultResponse)
}

func render(context *gin.Context) {
	// Create a temporary directory
	dir, err := os.MkdirTemp("", "")
	if err != nil {
		context.String(http.StatusInternalServerError, "Internal Server Error")
		log.Println("Error creating temporary directory: ", err)
		return
	}
	defer os.RemoveAll(dir)
	// Create a hard link of the template file inside the temporary directory
	err = os.Link(templateFileName, dir+"/"+templateFileName)
	if err != nil {
		context.String(http.StatusInternalServerError, "Internal Server Error")
		log.Println("Error linking template file: ", err)
		return
	}
	// Write the received data to a JSON file inside the temporary directory
	jsonData, err := io.ReadAll(context.Request.Body)
	if err != nil {
		context.String(http.StatusInternalServerError, "Internal Server Error")
		log.Println("Error reading request body: ", err)
		return
	}
	dataFile, err := os.Create(dir + "/data.json")
	if err != nil {
		context.String(http.StatusInternalServerError, "Internal Server Error")
		log.Println("Error creating data file: ", err)
		return
	}
	defer dataFile.Close()
	_, err = dataFile.Write(jsonData)
	if err != nil {
		context.String(http.StatusInternalServerError, "Internal Server Error")
		log.Println("Error writing data to file: ", err)
		return
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
		context.String(http.StatusBadRequest, "Failed to compile template with the given data.\n\n"+string(output))
		return
	}
	// Return the compiled file
	context.FileAttachment(dir+"/"+outputFileName, outputFileName)
}
