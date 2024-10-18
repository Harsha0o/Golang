package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	pb "github.com/hyperledger/fabric-protos-go/peer"
)

// DealerAsset defines the structure of an asset (renamed to avoid conflict)
type DealerAsset struct {
	DEALERID    string `json:"DEALERID"`
	MSISDN      string `json:"MSISDN"`
	MPIN        string `json:"MPIN"`
	BALANCE     string `json:"BALANCE"`
	STATUS      string `json:"STATUS"`
	TRANSAMOUNT string `json:"TRANSAMOUNT"`
	TRANSTYPE   string `json:"TRANSTYPE"`
	REMARKS     string `json:"REMARKS"`
}

// SimpleChaincode implements the chaincode interface
type SimpleChaincode struct {
}

// Init initializes the chaincode
func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

// Invoke handles the different functions of the chaincode
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()
	if len(args) < 1 {
		return shim.Error("Invalid number of arguments")
	}

	switch function {
	case "createAsset":
		return t.createAsset(stub, args)
	case "updateAsset":
		return t.updateAsset(stub, args)
	case "queryAsset":
		return t.queryAsset(stub, args)
	case "getAssetHistory":
		return t.getAssetHistory(stub, args)
	default:
		return shim.Error("Invalid function name")
	}
}

// createAsset creates a new asset
func (t *SimpleChaincode) createAsset(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 8 {
		return shim.Error("Incorrect number of arguments. Expecting 8")
	}

	// Check if the asset already exists
	assetBytes, err := stub.GetState(args[0])
	if err != nil {
		return shim.Error("Failed to get asset: " + err.Error())
	} else if assetBytes != nil {
		return shim.Error("This asset already exists: " + args[0])
	}

	// Create the asset object
	asset := DealerAsset{
		DEALERID:    args[0],
		MSISDN:      args[1],
		MPIN:        args[2],
		BALANCE:     args[3],
		STATUS:      args[4],
		TRANSAMOUNT: args[5],
		TRANSTYPE:   args[6],
		REMARKS:     args[7],
	}

	// Convert the asset to JSON
	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return shim.Error(err.Error())
	}

	// Save the asset on the ledger
	err = stub.PutState(asset.DEALERID, assetJSON)
	if err != nil {
		return shim.Error(err.Error())
	}

	// Emit an event
	err = stub.SetEvent("AssetCreated", []byte(asset.DEALERID))
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

// updateAsset updates an existing asset
func (t *SimpleChaincode) updateAsset(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 8 {
		return shim.Error("Incorrect number of arguments. Expecting 8")
	}

	asset := DealerAsset{
		DEALERID:    args[0],
		MSISDN:      args[1],
		MPIN:        args[2],
		BALANCE:     args[3],
		STATUS:      args[4],
		TRANSAMOUNT: args[5],
		TRANSTYPE:   args[6],
		REMARKS:     args[7],
	}

	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return shim.Error(err.Error())
	}

	err = stub.PutState(asset.DEALERID, assetJSON)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

// queryAsset queries an asset by its DEALERID
func (t *SimpleChaincode) queryAsset(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	assetJSON, err := stub.GetState(args[0])
	if err != nil {
		return shim.Error(err.Error())
	}
	if assetJSON == nil {
		return shim.Error("Asset not found")
	}

	return shim.Success(assetJSON)
}

// getAssetHistory retrieves the history of an asset by its DEALERID
func (t *SimpleChaincode) getAssetHistory(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	resultsIterator, err := stub.GetHistoryForKey(args[0])
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		response, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}

		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"TxId\":")
		buffer.WriteString("\"")
		buffer.WriteString(response.TxId)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Value\":")
		if response.IsDelete {
			buffer.WriteString("null")
		} else {
			buffer.WriteString(string(response.Value))
		}

		buffer.WriteString(", \"Timestamp\":")
		buffer.WriteString("\"")
		buffer.WriteString(time.Unix(response.Timestamp.Seconds, int64(response.Timestamp.Nanos)).String())
		buffer.WriteString("\"")

		buffer.WriteString(", \"IsDelete\":")
		buffer.WriteString("\"")
		buffer.WriteString(strconv.FormatBool(response.IsDelete))
		buffer.WriteString("\"}")

		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	return shim.Success(buffer.Bytes())
}

func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting SimpleChaincode: %s", err)
	}
}
