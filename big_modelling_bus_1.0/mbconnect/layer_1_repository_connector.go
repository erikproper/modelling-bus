/*
 *
 * Package: mbconnect
 * Layer:   1
 * Module:  repository_connector
 *
 * ..... ... .. .
 *
 * Creator: Henderik A. Proper (e.proper@acm.org), TU Wien, Austria
 *
 * Version of: XX.10.2025
 *
 */

package mbconnect

import (
	"github.com/secsy/goftp"
	"os"
	"path/filepath"
	"strings"
)

type (
	tModellingBusRepositoryConnector struct {
		ftpPort,
		ftpUser,
		ftpAgentRoot,
		ftpServer,
		ftpPassword,
		ftpPathPrefix,
		ftpLocalWorkDirectory string

		createdPaths map[string]bool

		errorReporter TErrorReporter
	}
)

func (r *tModellingBusRepositoryConnector) ftpConnect() (*goftp.Client, error) {
	config := goftp.Config{}
	config.User = r.ftpUser
	config.Password = r.ftpPassword

	ftpServerDefinition := r.ftpServer + ":" + r.ftpPort
	client, err := goftp.DialConfig(config, ftpServerDefinition)
	if err != nil {
		r.errorReporter("Error connecting to the FTP server:", err)
		return client, err
	}

	return client, err
}

func (r *tModellingBusRepositoryConnector) mkRepositoryDirectoryPath(remoteDirectoryPath string) {
	if !r.createdPaths[remoteDirectoryPath] {
		// Connect to the FTP server
		client, err := r.ftpConnect()
		if err != nil {
			r.errorReporter("Couldn't open an FTP connection:", err)
			return
		}

		pathCovered := ""
		for _, Directory := range strings.Split(remoteDirectoryPath, "/") {
			pathCovered = pathCovered + Directory + "/"
			client.Mkdir(pathCovered)
		}

		client.Close()

		r.createdPaths[remoteDirectoryPath] = true
	}
}

func (r *tModellingBusRepositoryConnector) pushFileToRepository(topicPath, fileName, localFilePath string) {
	remoteDirectoryPath := r.ftpAgentRoot + "/" + topicPath
	remoteFilePath := remoteDirectoryPath + "/" + fileName

	r.mkRepositoryDirectoryPath(remoteDirectoryPath)

	// Connect to the FTP server
	client, err := r.ftpConnect()
	if err != nil {
		r.errorReporter("Couldn't open an FTP connection:", err)
		return
	}

	file, err := os.Open(localFilePath)
	if err != nil {
		r.errorReporter("Error opening File for reading:", err)
		return
	}

	err = client.Store(remoteFilePath, file)
	if err != nil {
		r.errorReporter("Error uploading File to ftp server:", err)
		return
	}

	client.Close()
}

func (r *tModellingBusRepositoryConnector) pushJSONAsFileToRepository(topicPath, fileName string, json []byte, timestamp string) {
	// Define the file paths
	localFilePath := r.ftpLocalWorkDirectory + "/" + fileName

	// Create a temporary local file with the JSON record
	err := os.WriteFile(localFilePath, json, 0644)
	if err != nil {
		r.errorReporter("Error writing to temporary file:", err)
	}

	r.pushFileToRepository(topicPath, fileName, localFilePath)

	// Cleanup the temporary file aftewards
	os.Remove(localFilePath)
}

func (r *tModellingBusRepositoryConnector) cleanRepositoryPath(topicPath, timestamp string) {
	// Connect to the FTP server
	client, err := r.ftpConnect()
	if err != nil {
		r.errorReporter("Couldn't open an FTP connection:", err)
		return
	}

	fileInfos, _ := client.ReadDir(r.ftpAgentRoot + "/" + topicPath)

	// Remove older Files from the FTP server within the topicPath Directory
	for _, fileInfo := range fileInfos {
		if timestamp == "" {
			err = client.Delete(fileInfo.Name())
			if err != nil {
				r.errorReporter("Couldn't delete File:", err)
				return
			}
		} else {
			filePath := fileInfo.Name()
			fileName := strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))

			if fileName < timestamp {
				client.Delete(filePath)
			}
		}
	}
}

func (r *tModellingBusRepositoryConnector) getFileFromRepository(server, port, remoteFilePath, localFileName string) {
	client, err := goftp.DialConfig(goftp.Config{}, server+":"+port)
	if err != nil {
		r.errorReporter("Something went wrong connecting to the FTP server", err)
		return
	}

	// Download an File to disk
	// ====> CHECK need for OS (Dos, Linux, ...) independent "/"
	File, err := os.Create(localFileName)
	if err != nil {
		r.errorReporter("Something went wrong creating local file", err)
		return
	}

	err = client.Retrieve(remoteFilePath, File)
	if err != nil {
		r.errorReporter("Something went wrong retrieving file", err)
		return
	}
}

func createModellingBusRepositoryConnector(topicBase, agentID string, configData *TConfigData, errorReporter TErrorReporter) *tModellingBusRepositoryConnector {
	r := tModellingBusRepositoryConnector{}

	r.errorReporter = errorReporter

	// Get data from the config file
	r.ftpLocalWorkDirectory = configData.GetValue("", "work").String()
	r.ftpPort = configData.GetValue("ftp", "port").String()
	r.ftpUser = configData.GetValue("ftp", "user").String()
	r.ftpServer = configData.GetValue("ftp", "server").String()
	r.ftpPassword = configData.GetValue("ftp", "password").String()

	// Needed???
	r.ftpPathPrefix = configData.GetValue("ftp", "prefix").String()

	r.ftpAgentRoot = r.ftpPathPrefix + "/" + topicBase + "/" + agentID

	r.createdPaths = map[string]bool{}

	return &r
}
