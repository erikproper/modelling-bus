/*
 *
 * Module:    BIG Modelling Bus
 * Package:   Generic
 * Component: Layer 3 - Observation
 *
 * ..... ... .. .
 *
 * Creator: Henderik A. Proper (e.proper@acm.org), TU Wien, Austria
 *
 * Version of: 05.12.2025
 *
 */

package connect

import (
	"github.com/erikproper/big-modelling-bus.go.v1/generics"
)

const (
	rawObservationFilePathElement  = "observation/raw"
	jsonObservationFilePathElement = "observation/json"
)

/*
 * Defining topic paths
 */

func (b *TModellingBusConnector) rawObservationsTopicPath(observationID string) string {
	return rawArtefactsPathElement +
		"/" + observationID
}

func (b *TModellingBusConnector) jsonObservationsTopicPath(observationID string) string {
	return jsonArtefactsPathElement +
		"/" + observationID
}

/*
 *
 * Externally visible functionality
 *
 */

/*
 * Posting artefacts
 */

func (b *TModellingBusConnector) PostRawObservation(observationID, localFilePath string) {
	b.postFile(b.rawObservationsTopicPath(observationID), localFilePath, generics.GetTimestamp())
}

func (b *TModellingBusConnector) PostJSONObservation(observationID string, json []byte) {
	b.postJSON(b.jsonObservationsTopicPath(observationID), json, generics.GetTimestamp())
}

/*
 * Listening to observations related postings
 */

func (b *TModellingBusConnector) ListenForRawObservationPostings(agentID, observationID string, postingHandler func(string)) {
	b.listenForFilePostings(agentID, b.jsonObservationsTopicPath(observationID), generics.JSONFileName, func(localFilePath, _ string) {
		postingHandler(localFilePath)
	})
}

// HERE!
func (b *TModellingBusConnector) ListenForJSONObservationPostings(agentID, topicPath string, postingHandler func([]byte, string)) {
	b.listenForFilePostings(agentID, topicPath, generics.JSONFileName, func(localFilePath, timestamp string) {
		postingHandler(b.getJSONFromTemporaryFile(localFilePath, timestamp))
	})

	//	b.listenForJSONPostings(agentID, b.jsonArtefactsUpdateTopicPath(artefactID), func(json []byte, _ string) {
	//		if b.updateUpdatedJSONArtefact(json) {
	//			handler()
	//		}
	//	})

	//	func (b *TModellingBusConnector) listenForJSONPostings(agentID, topicPath string, postingHandler func([]byte, string)) {
	//	b.modellingBusEventsConnector.listenForEvents(agentID, topicPath, func(message []byte) {
	//		postingHandler(b.getJSONFromTemporaryFile(b.getLinkedFileFromRepository(message, generics.JSONFileName)))
	//	})

	// func (b *TModellingBusConnector) getJSONFromTemporaryFile(tempFilePath, timestamp string) ([]byte, string) {
	//}

}

//
// func (b *TModellingBusConnector) GetRawObservation(agentID, topicPath, localFileName string) string {
// 	localFilePath, _ := b.getFileFromPosting(agentID, topicPath, localFileName)
// 	return localFilePath
// }
//
// func (b *TModellingBusConnector) DeleteRawObservation(topicPath string) {
// 	b.deletePosting(topicPath)
// }
