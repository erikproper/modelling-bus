/*
 *
 * Package: mbconnect
 * Layer:   2
 * Module:  basic_modelling_bus
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
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// ///// GENERIC
const (
	modellingBusVersion = "bus-version-1.0"
	jsonFileExtension   = ".json"
)

type (
	TModellingBusConnector struct {
		configData *TConfigData

		modellingBusRepositoryConnector *tModellingBusRepositoryConnector
		modellingBusEventsConnector     *tModellingBusEventsConnector

		AgentID, // used "up" here?
		experimentID, // used "up" here?
		lastTimeTimestamp string

		timestampCounter int

		errorReporter TErrorReporter // used "up" here?
	}
)

/*
 *
 * Internal functionality
 *
 */

/*
 * MQTT connection
 */

// ///
func (b *TModellingBusConnector) listenForRawFileLinkPostingsOnMQTT(AgentID, topicPath string, postingHandler func(string, string, string, string)) {
	b.modellingBusEventsConnector.listenForEvents(AgentID, topicPath, func(message []byte) {
		var rawFileLink TRawFileLink
		/// Use a generic error checker for Unmarshal. Shouldreturn a bool
		err := json.Unmarshal(message, &rawFileLink)
		if err == nil {
			postingHandler(rawFileLink.Server, rawFileLink.Port, rawFileLink.Path, rawFileLink.Timestamp)
		}
	})
}

type TJSONFileLink struct {
	Server      string `json:"server"`
	Port        string `json:"port"`
	Path        string `json:"path"`
	Timestamp   string `json:"timestamp"`
	JSONVersion string `json:"json version"`
}

type TRawFileLink struct {
	Server    string `json:"server"`
	Port      string `json:"port"`
	Path      string `json:"path"`
	Timestamp string `json:"timestamp"`
}

func (b *TModellingBusConnector) postJSONFileLinkToMQTT(topicPath, jsonFileName, jsonVersion, timestamp string) {
	var jsonFileLink TJSONFileLink

	jsonFileLink.Server = b.modellingBusRepositoryConnector.ftpServer
	jsonFileLink.Port = b.modellingBusRepositoryConnector.ftpPort
	jsonFileLink.Path = b.modellingBusRepositoryConnector.ftpAgentRoot + "/" + topicPath + "/" + jsonFileName
	jsonFileLink.Timestamp = timestamp
	jsonFileLink.JSONVersion = jsonVersion

	jsonData, err := json.Marshal(jsonFileLink)
	if err != nil {
		b.errorReporter("Something went wrong JSONing the link data", err)
		return
	}

	b.modellingBusEventsConnector.postEvent(topicPath, jsonData)
}

func (b *TModellingBusConnector) listenForJSONFileLinkPostingsOnMQTT(AgentID, topicPath string, postingHandler func(string, string, string, string, string)) {
	b.modellingBusEventsConnector.listenForEvents(AgentID, topicPath, func(message []byte) {
		var jsonFileLink TJSONFileLink

		err := json.Unmarshal(message, &jsonFileLink)
		if err == nil {
			postingHandler(jsonFileLink.Server, jsonFileLink.Port, jsonFileLink.Path, jsonFileLink.Timestamp, jsonFileLink.JSONVersion)
		}
	})
}

func (b *TModellingBusConnector) postRawFileLinkToMQTT(topicPath, rawFileName, timestamp string) {
	var rawFileLink TRawFileLink

	rawFileLink.Server = b.modellingBusRepositoryConnector.ftpServer
	rawFileLink.Port = b.modellingBusRepositoryConnector.ftpPort
	rawFileLink.Path = b.modellingBusRepositoryConnector.ftpAgentRoot + "/" + topicPath + "/" + rawFileName
	rawFileLink.Timestamp = timestamp

	data, err := json.Marshal(rawFileLink)
	if err != nil {
		b.errorReporter("Something went wrong JSONing the link data", err)
		return
	}

	b.modellingBusEventsConnector.postEvent(topicPath, data)
}

/*
 * Combined FTP + MQTT connection
 */

func (b *TModellingBusConnector) postJSONFile(topicPath, jsonVersion, timestamp string, json []byte) {
	fileName := timestamp + jsonFileExtension

	b.modellingBusRepositoryConnector.pushJSONAsFileToRepository(topicPath, fileName, json, timestamp)
	b.postJSONFileLinkToMQTT(topicPath, fileName, jsonVersion, timestamp)
}

func (b *TModellingBusConnector) postRawFile(topicPath, fileName, localFilePath, timestamp string) {
	b.modellingBusRepositoryConnector.pushFileToRepository(topicPath, fileName, localFilePath)
	b.postRawFileLinkToMQTT(topicPath, fileName, timestamp)
}

// CLARIFY things above. Some of this may need to move up one layer

// func (b *TModellingBusConnector) postRawFile (...)
// func (b *TModellingBusConnector) listenForRawFileEvent(AgentID, topicPath string, postingHandler func(string, []byte)) {

// func (b *TModellingBusConnector) postJSONEventAsFile (...)
// func (b *TModellingBusConnector) postJSONEvent (...)
// func (b *TModellingBusConnector) listenForJSONEvents(AgentID, topicPath string, postingHandler func(string, []byte)) {

func (b *TModellingBusConnector) listenForJSONFilePostings(AgentID, topicPath string, postingHandler func(string, []byte)) {
	b.listenForJSONFileLinkPostingsOnMQTT(AgentID, topicPath, func(server, port, path, timestamp, jsonVersion string) {
		tempFilePath := b.modellingBusRepositoryConnector.ftpLocalWorkDirectory + "/" + b.GetTimestamp() + jsonFileExtension

		b.modellingBusRepositoryConnector.getFileFromRepository(server, port, path, tempFilePath)

		jsonPayload, err := os.ReadFile(tempFilePath)
		if err == nil {
			postingHandler(timestamp, jsonPayload)
		} else {
			b.errorReporter("Something went wrong retrieving file", err)
		}

		os.Remove(tempFilePath)
	})
}

/*
 *
 * Externally visible functionality
 *
 */

/*
 * Time stamping and unique IDs
 */

func (b *TModellingBusConnector) GetTimestamp() string {
	CurrenTime := time.Now()

	timeTimestamp := fmt.Sprintf(
		"%04d-%02d-%02d-%02d-%02d-%02d", //-%06d
		CurrenTime.Year(),
		CurrenTime.Month(),
		CurrenTime.Day(),
		CurrenTime.Hour(),
		CurrenTime.Minute(),
		CurrenTime.Second())
	//		CurrenTime.Nanosecond()/1000)

	if timeTimestamp == b.lastTimeTimestamp {
		b.timestampCounter++
	} else {
		b.lastTimeTimestamp = timeTimestamp
		b.timestampCounter = 0
	}

	return fmt.Sprintf("%s-%02d", b.lastTimeTimestamp, b.timestampCounter)
}

func (b *TModellingBusConnector) GetNewID() string {
	return fmt.Sprintf("%s-%s", b.AgentID, b.GetTimestamp())
}

/*
 * Initialisation & creation
 */

func (b *TModellingBusConnector) Initialise(configPath string, errorReporter TErrorReporter) {
	var ok bool

	// This needs to be done on the top level ...
	b.errorReporter = errorReporter
	b.configData, ok = LoadConfig(configPath, b.errorReporter)
	if !ok {
		fmt.Println("Config file not found ... need to fix this")
	}

	b.experimentID = b.configData.GetValue("", "experiment").String() // ever used beyond the calls below?
	b.AgentID = b.configData.GetValue("", "agent").String()           // ever used beyond the calls below?

	topicBase := modellingBusVersion + "/" + b.experimentID

	b.modellingBusRepositoryConnector = createModellingBusRepositoryConnector(topicBase, b.AgentID, b.configData, b.errorReporter)
	b.modellingBusEventsConnector = createModellingBusEventsConnector(topicBase, b.AgentID, b.configData, b.errorReporter)

	b.lastTimeTimestamp = ""
	b.timestampCounter = 0

}

func CreateModellingBusConnector(config string, errorReporter TErrorReporter) TModellingBusConnector {
	modellingBusConnector := TModellingBusConnector{}
	modellingBusConnector.Initialise(config, errorReporter)

	return modellingBusConnector
}
