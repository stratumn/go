// Copyright 2017 Stratumn SAS. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/stratumn/sdk/store"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	sc "github.com/hyperledger/fabric/protos/peer"

	"github.com/stratumn/sdk/cs"
)

// Pagination functionality (limit & skip) is implemented in CouchDB but not in Hyperledger Fabric (FAB-2809 and FAB-5369).
// Creating an index in CouchDB:
// curl -i -X POST -H "Content-Type: application/json" -d "{\"index\":{\"fields\":[\"chaincodeid\",\"data.docType\",\"data.id\"]},\"name\":\"indexOwner\",\"ddoc\":\"indexOwnerDoc\",\"type\":\"json\"}" http://localhost:5984/mychannel/_index

// SmartContract defines chaincode logic
type SmartContract struct {
}

// ObjectType used in CouchDB documents
const (
	ObjectTypeSegment = "segment"
	ObjectTypeMap     = "map"
	ObjectTypeValue   = "value"
)

// MapDoc is used to store maps in CouchDB
type MapDoc struct {
	ObjectType string `json:"docType"`
	ID         string `json:"id"`
	Process    string `json:"process"`
}

// SegmentDoc is used to store segments in CouchDB
type SegmentDoc struct {
	ObjectType string     `json:"docType"`
	ID         string     `json:"id"`
	Segment    cs.Segment `json:"segment"`
}

// MapSelector used in MapQuery
type MapSelector struct {
	ObjectType string `json:"docType"`
	Process    string `json:"process,omitempty"`
}

// MapQuery used in CouchDB rich queries
type MapQuery struct {
	Selector MapSelector `json:"selector,omitempty"`
	Limit    int         `json:"limit,omitempty"`
	Skip     int         `json:"skip,omitempty"`
}

func newMapQuery(filterBytes []byte) (string, error) {
	filter := &store.MapFilter{}
	if err := json.Unmarshal(filterBytes, filter); err != nil {
		return "", err
	}

	mapSelector := MapSelector{}
	mapSelector.ObjectType = ObjectTypeMap

	if filter.Process != "" {
		mapSelector.Process = filter.Process
	}

	mapQuery := MapQuery{
		Selector: mapSelector,
		Limit:    filter.Pagination.Limit,
		Skip:     filter.Pagination.Offset,
	}

	queryBytes, err := json.Marshal(mapQuery)
	if err != nil {
		return "", err
	}

	return string(queryBytes), nil
}

// SegmentSelector used in SegmentQuery
type SegmentSelector struct {
	ObjectType   string    `json:"docType"`
	LinkHash     string    `json:"id,omitempty"`
	PrevLinkHash string    `json:"segment.link.meta.prevLinkHash,omitempty"`
	Process      string    `json:"segment.link.meta.process,omitempty"`
	MapIds       *MapIdsIn `json:"segment.link.meta.mapId,omitempty"`
	Tags         *TagsAll  `json:"segment.link.meta.tags,omitempty"`
}

// MapIdsIn specifies that segment mapId should be in specified list
type MapIdsIn struct {
	MapIds []string `json:"$in,omitempty"`
}

// TagsAll specifies all tags in specified list should be in segment tags
type TagsAll struct {
	Tags []string `json:"$all,omitempty"`
}

// SegmentQuery used in CouchDB rich queries
type SegmentQuery struct {
	Selector SegmentSelector `json:"selector,omitempty"`
	Limit    int             `json:"limit,omitempty"`
	Skip     int             `json:"skip,omitempty"`
}

func newSegmentQuery(filterBytes []byte) (string, error) {
	filter := &store.SegmentFilter{}
	if err := json.Unmarshal(filterBytes, filter); err != nil {
		return "", err
	}

	segmentSelector := SegmentSelector{}
	segmentSelector.ObjectType = ObjectTypeSegment

	if filter.PrevLinkHash != nil {
		segmentSelector.PrevLinkHash = filter.PrevLinkHash.String()
	}
	if filter.Process != "" {
		segmentSelector.Process = filter.Process
	}
	if len(filter.MapIDs) > 0 {
		segmentSelector.MapIds = &MapIdsIn{filter.MapIDs}
	} else {
		segmentSelector.Tags = nil
	}
	if len(filter.Tags) > 0 {
		segmentSelector.Tags = &TagsAll{filter.Tags}
	} else {
		segmentSelector.Tags = nil
	}

	segmentQuery := SegmentQuery{
		Selector: segmentSelector,
		Limit:    filter.Pagination.Limit,
		Skip:     filter.Pagination.Offset,
	}

	queryBytes, err := json.Marshal(segmentQuery)
	if err != nil {
		return "", err
	}

	return string(queryBytes), nil
}

// Init method is called when the Smart Contract "pop" is instantiated by the blockchain network
func (s *SmartContract) Init(APIstub shim.ChaincodeStubInterface) sc.Response {
	return shim.Success(nil)
}

// Invoke method is called as a result of an application request to run the Smart Contract "pop"
func (s *SmartContract) Invoke(APIstub shim.ChaincodeStubInterface) sc.Response {
	// Retrieve the requested Smart Contract function and arguments
	function, args := APIstub.GetFunctionAndParameters()

	switch function {
	case "GetSegment":
		return s.GetSegment(APIstub, args)
	case "FindSegments":
		return s.FindSegments(APIstub, args)
	case "GetMapIDs":
		return s.GetMapIDs(APIstub, args)
	case "SaveSegment":
		return s.SaveSegment(APIstub, args)
	case "DeleteSegment":
		return s.DeleteSegment(APIstub, args)
	case "SaveValue":
		return s.SaveValue(APIstub, args)
	case "GetValue":
		return s.GetValue(APIstub, args)
	case "DeleteValue":
		return s.DeleteValue(APIstub, args)
	default:
		return shim.Error("Invalid Smart Contract function name: " + function)
	}
}

// SaveMap saves map into CouchDB using map document
func (s *SmartContract) SaveMap(stub shim.ChaincodeStubInterface, segment *cs.Segment) error {
	mapDoc := MapDoc{
		ObjectTypeMap,
		segment.Link.GetMapID(),
		segment.Link.GetProcess(),
	}
	mapDocBytes, err := json.Marshal(mapDoc)
	if err != nil {
		return err
	}

	return stub.PutState(segment.Link.GetMapID(), mapDocBytes)
}

// SaveSegment saves segment into CouchDB using segment document
func (s *SmartContract) SaveSegment(stub shim.ChaincodeStubInterface, args []string) sc.Response {
	// Parse segment
	byteArgs := stub.GetArgs()
	segment := &cs.Segment{}
	if err := json.Unmarshal(byteArgs[1], segment); err != nil {
		return shim.Error("Could not parse segment")
	}

	// Validate segment
	if err := segment.Validate(); err != nil {
		return shim.Error(err.Error())
	}
	// Set pending evidence
	segment.SetEvidence(
		map[string]interface{}{
			"state":        cs.PendingEvidence,
			"transactions": map[string]string{"transactionID": stub.GetTxID()},
		})

	// Check has prevLinkHash if not create map else check prevLinkHash exists
	prevLinkHash := segment.Link.GetPrevLinkHashString()
	if prevLinkHash == "" {
		// Create map
		if err := s.SaveMap(stub, segment); err != nil {
			return shim.Error(err.Error())
		}
	}

	//  Save segment
	segmentDoc := SegmentDoc{
		ObjectTypeSegment,
		segment.GetLinkHashString(),
		*segment,
	}
	segmentDocBytes, err := json.Marshal(segmentDoc)
	if err != nil {
		return shim.Error(err.Error())
	}
	if err := stub.PutState(segment.GetLinkHashString(), segmentDocBytes); err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

// GetSegment gets segment for given linkHash
func (s *SmartContract) GetSegment(stub shim.ChaincodeStubInterface, args []string) sc.Response {
	segmentDocBytes, err := stub.GetState(args[0])
	if err != nil {
		return shim.Error(err.Error())
	}
	if segmentDocBytes == nil {
		return shim.Success(nil)
	}

	segmentBytes, err := extractSegment(segmentDocBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(segmentBytes)
}

// DeleteSegment deletes segment from CouchDB
func (s *SmartContract) DeleteSegment(stub shim.ChaincodeStubInterface, args []string) sc.Response {
	shimResponse := s.GetSegment(stub, args)
	if shimResponse.Status == shim.ERROR {
		return shimResponse
	}
	segmentBytes := shimResponse.Payload
	err := stub.DelState(args[0])
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(segmentBytes)
}

// FindSegments returns segments that match specified segment filter
func (s *SmartContract) FindSegments(stub shim.ChaincodeStubInterface, args []string) sc.Response {
	queryString, err := newSegmentQuery([]byte(args[0]))
	if err != nil {
		return shim.Error("Segment filter format incorrect")
	}

	resultsIterator, err := stub.GetQueryResult(queryString)
	if err != nil {
		return shim.Error(err.Error())
	}

	var segments cs.SegmentSlice

	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		segmentDoc := &SegmentDoc{}
		if err := json.Unmarshal(queryResponse.Value, segmentDoc); err != nil {
			return shim.Error(err.Error())
		}
		segments = append(segments, &segmentDoc.Segment)
	}
	sort.Sort(segments)

	resultBytes, err := json.Marshal(segments)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(resultBytes)
}

// GetMapIDs returns mapIDs for maps that match specified map filter
func (s *SmartContract) GetMapIDs(stub shim.ChaincodeStubInterface, args []string) sc.Response {
	queryString, err := newMapQuery([]byte(args[0]))
	if err != nil {
		return shim.Error("Map filter format incorrect")
	}

	resultsIterator, err := stub.GetQueryResult(queryString)
	if err != nil {
		return shim.Error(err.Error())
	}

	var mapIDs []string
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		mapIDs = append(mapIDs, queryResponse.Key)
	}

	sort.Strings(mapIDs)
	resultBytes, err := json.Marshal(mapIDs)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(resultBytes)
}

// SaveValue saves key, value in CouchDB
func (s *SmartContract) SaveValue(stub shim.ChaincodeStubInterface, args []string) sc.Response {
	compositeKey, err := getValueCompositeKey(args[0], stub)
	if err != nil {
		return shim.Error(err.Error())
	}
	err = stub.PutState(compositeKey, []byte(args[1]))
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(nil)
}

// GetValue gets value for specified key from CouchDB
func (s *SmartContract) GetValue(stub shim.ChaincodeStubInterface, args []string) sc.Response {
	compositeKey, err := getValueCompositeKey(args[0], stub)
	if err != nil {
		return shim.Error(err.Error())
	}
	value, err := stub.GetState(compositeKey)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(value)
}

// DeleteValue deletes key, value from CouchDB
func (s *SmartContract) DeleteValue(stub shim.ChaincodeStubInterface, args []string) sc.Response {
	compositeKey, err := getValueCompositeKey(args[0], stub)
	if err != nil {
		return shim.Error(err.Error())
	}
	value, err := stub.GetState(compositeKey)
	if err != nil {
		return shim.Error(err.Error())
	}

	err = stub.DelState(compositeKey)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(value)
}

func extractSegment(segmentDocBytes []byte) ([]byte, error) {
	segmentDoc := &SegmentDoc{}
	if err := json.Unmarshal(segmentDocBytes, segmentDoc); err != nil {
		return nil, err
	}
	segmentBytes, err := json.Marshal(segmentDoc.Segment)
	if err != nil {
		return nil, err
	}
	return segmentBytes, nil
}

func getValueCompositeKey(key string, stub shim.ChaincodeStubInterface) (compositeKey string, err error) {
	compositeKey, err = stub.CreateCompositeKey(ObjectTypeValue, []string{key})
	return
}

// main function starts up the chaincode in the container during instantiate
func main() {
	if err := shim.Start(new(SmartContract)); err != nil {
		fmt.Printf("Error starting SmartContract chaincode: %s", err)
	}
}
