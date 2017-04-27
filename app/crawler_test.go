package service

import (
	"testing"
	//"fmt"
	//"github.com/Pallinder/go-randomdata"
	"github.com/stretchr/testify/suite"
	"log"
	"net/http/httptest"
	//"net/http"
	//"net/http/httptest"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
)

type CrawlerTestSuite struct {
	suite.Suite
	server                        httptest.Server
	VariableThatShouldStartAtFive int
	testDataLocation              string
}

func (suite *CrawlerTestSuite) SetupSuite() {
	log.Print("Setting up suite")
	suite.testDataLocation = `c:\test\images`
	suite.server = *httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			suite.handleClient(w, r)
		}))
}

var (
	testImages = []string{
		"image1.gif",
		"image2.gif",
	}
)

func (suite *CrawlerTestSuite) TestDownloadImageFromUrl() {
	for _, testImage := range testImages {
		err := downloadImageFromUrl(suite.server.URL + testImage,testImage,
			suite.testDataLocation)
		suite.Assert().Nil(err)
	}
	suite.Assert().True(true, true, "z")
}

func TestModelTestSuite(t *testing.T) {
	suite.Run(t, new(CrawlerTestSuite))
}
func (suite *CrawlerTestSuite) TearDownSuite() {
	log.Print("Closing test server")
	//suite.server.Close()
}
func (suite *CrawlerTestSuite) handleClient(writer http.ResponseWriter, request *http.Request) {
	//First of check if Get is set in the URL
	Filename := request.URL.Query().Get("file")
	if Filename == "" {
		//Get not set, send a 400 bad request
		http.Error(writer, "Get 'file' not specified in url.", 400)
		return
	}
	fmt.Println("Client requests: " + Filename)

	//Check if file exists and open
	Openfile, err := os.Open(suite.testDataLocation + Filename)
	defer Openfile.Close() //Close after function return
	if err != nil {
		//File not found, send 404
		http.Error(writer, "File not found.", 404)
		return
	}

	//File is found, create and send the correct headers

	//Get the Content-Type of the file
	//Create a buffer to store the header of the file in
	FileHeader := make([]byte, 512)
	//Copy the headers into the FileHeader buffer
	Openfile.Read(FileHeader)
	//Get content type of file
	FileContentType := http.DetectContentType(FileHeader)

	//Get the file size
	FileStat, _ := Openfile.Stat()                     //Get info from file
	FileSize := strconv.FormatInt(FileStat.Size(), 10) //Get file size as a string

	//Send the headers
	writer.Header().Set("Content-Disposition", "attachment; filename="+Filename)
	writer.Header().Set("Content-Type", FileContentType)
	writer.Header().Set("Content-Length", FileSize)

	//Send the file
	//We read 512 bytes from the file already so we reset the offset back to 0
	Openfile.Seek(0, 0)
	io.Copy(writer, Openfile) //'Copy' the file to the client
	return
}
