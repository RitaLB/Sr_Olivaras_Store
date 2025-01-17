/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	pb "github.com/hyperledger/fabric-protos-go/peer"
)

// MaterialsPrivateChaincode example Chaincode implementation
type Studio struct {
}

type Material struct {
	ObjectType string `json:"docType"` //docType is used to distinguish the various types of objects in state database
	ID         string `json:"ID"`      //the fieldtags are needed to keep case from bouncing around
	Type       string `json:"type"`    //the fieldtags are needed to keep case from bouncing around
	Supplier   string `json:"supplier"`
}

type Wand struct {
	ObjectType string   `json:"docType"` //docType is used to distinguish the various types of objects in state database
	ID         string   `json:"ID"`      //the fieldtags are needed to keep case from bouncing around
	Type       string   `json:"type"`    //the fieldtags are needed to keep case from bouncing around
	Color      string   `json:"color"`
	Size       int      `json:"size"`
	Materials  []string `json:"Materials"`
}

// Init initializes chaincode
// ===========================
func (t *Studio) Init(stub shim.ChaincodeStubInterface) pb.Response {
	// Verifica se a lista de materiais já existe no estado do mundo
	materialIndexListBytes, err := stub.GetState("materialIndexList")
	if err != nil {
		return shim.Error(fmt.Sprintf("Falha ao verificar a existência da lista de materiais: %s", err))
	}
	if materialIndexListBytes == nil {
		// A lista de materiais não existe, então a inicializamos
		var emptyMaterialIndexList []string
		emptyMaterialIndexListBytes, _ := json.Marshal(emptyMaterialIndexList)
		stub.PutState("materialIndexList", emptyMaterialIndexListBytes)
	}

	// Verifica se a lista de varinhas já existe no estado do mundo
	wandsIndexListBytes, err := stub.GetState("wandsIndexList")
	if err != nil {
		return shim.Error(fmt.Sprintf("Falha ao verificar a existência da lista de varinhas: %s", err))
	}
	if wandsIndexListBytes == nil {
		// A lista de varinhas não existe, então a inicializamos
		var emptyWandsIndexList []string
		emptyWandsIndexListBytes, _ := json.Marshal(emptyWandsIndexList)
		stub.PutState("wandsIndexList", emptyWandsIndexListBytes)
	}

	return shim.Success(nil)
}

// Invoke - Our entry point for Invocations
// ========================================
func (t *Studio) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()
	fmt.Println("invoke is running " + function)

	// Handle different functions
	switch function {
	case "initMaterial":
		//create a new material
		return t.initMaterial(stub, args)
	case "readMaterial":
		//read a marble
		return t.readMaterial(stub, args)
	case "getMaterialsByType":
		// read all materials of som specificy type
		return t.getMaterialsByType(stub, args)
	case "getAllMaterials":
		// returns all materials at MaterialsIndexList
		return t.getAllMaterials(stub)
	case "getNumberMaterialsByType":
		// returns number of materials of given type at the world state
		return t.getNumberMaterialsByType(stub, args)
	case "getTotalNumberOfMaterials":
		// returns total number of avaible materials on MaterialIndexList
		return t.getTotalNumberOfMaterials(stub)
	case "deleteMaterial":
		//delete the given ID material
		return t.deleteMaterial(stub, args)
	case "getAllMaterialsAndIndexList":
		// returns all materials and index list
		return t.getAllMaterialsAndIndexList(stub)
	case "getMaterialIndexList":
		// returns all materials index list
		return t.getMaterialIndexList(stub)
	case "initWand":
		//create a new wand
		return t.initWand(stub, args)
	case "readWand":
		// read the ID given wand from chaincode state
		return t.readWand(stub, args)
	case "getWandsByType":
		// returns all Wands of given type at wandsIndexList
		return t.getWandsByType(stub, args)
	case "getAllWands":
		// returns all wands available at wandsIndexList
		return t.getAllWands(stub)
	case "getNumberWandsByType":
		// returns number of wands of given type at wandsIndexList
		return t.getNumberwandsByType(stub, args)
	case "getTotalNumberOfWands":
		// returns total number of avaible wands on wandsIndexList
		return t.getTotalNumberOfWands(stub)
	case "deleteWand":
		// delete the given ID wand
		return t.deletewand(stub, args)
	default:
		//error
		fmt.Println("invoke did not find func: " + function)
		return shim.Error("Received unknown function invocation")
	}
}

// --------------------------------------------------------------------------------------------------------------------------------------------------------
// Funções Material

// ============================================================
// initMaterial - create a new material, store into chaincode state
// ============================================================
func (t *Studio) initMaterial(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error

	// ==== Input sanitation ====
	fmt.Println("- start init material")

	// Checks the correct number of arguments
	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments. Expecting ID, Type and Supplier")
	}

	// Extracting the arguments
	materialID := args[0]
	materialType := args[1]
	materialSupplier := args[2]

	// Checks that the material ID is not empty
	if materialID == "" {
		return shim.Error("Material ID cannot be empty")
	}

	// Checks if material with given ID already exists
	if materialAsBytes, _ := stub.GetState(materialID); materialAsBytes != nil {
		fmt.Println("This material already exists: " + materialID)
		return shim.Error("This material already exists: " + materialID)
	}

	// Creates a material
	material := &Material{
		ObjectType: "Material",
		ID:         materialID,
		Type:       materialType,
		Supplier:   materialSupplier,
	}

	// Marshal the material to JSON
	materialJSONasBytes, err := json.Marshal(material)
	if err != nil {
		return shim.Error(err.Error())
	}

	// === Save material to state ===
	err = stub.PutState(material.ID, materialJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	// ==== Index the material to enable type-based range queries ====
	// The composite key is based on indexName~type~ID.
	// This will enable very efficient range queries based on composite keys matching indexName~type~*
	indexName := "type~ID"
	typeIndexKey, err := stub.CreateCompositeKey(indexName, []string{material.Type, material.ID})
	if err != nil {
		return shim.Error(err.Error())
	}

	// Get the current material index list
	indexListBytes, err := stub.GetState("materialIndexList")
	if err != nil {
		return shim.Error(err.Error())
	}

	var indexList []string
	if indexListBytes != nil {
		// Unmarshal the index list
		err = json.Unmarshal(indexListBytes, &indexList)
		if err != nil {
			return shim.Error(err.Error())
		}
	}

	// Append the new index to the index list
	indexList = append(indexList, typeIndexKey)

	// Marshal the updated index list to JSON
	updatedIndexListBytes, err := json.Marshal(indexList)
	if err != nil {
		return shim.Error(err.Error())
	}

	// Save the updated index list to state
	err = stub.PutState("materialIndexList", updatedIndexListBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	// Save the index entry to the state.
	// Only the key name is needed, no need to store a duplicate copy of the material.
	// Note - passing a 'nil' value will effectively delete the key from state, therefore we pass a null character as value
	err = stub.PutState(typeIndexKey, []byte{0x00})
	if err != nil {
		return shim.Error(err.Error())
	}

	// ==== Material saved. Return success ====
	fmt.Println("- end init material")
	return shim.Success(updatedIndexListBytes)
}

// ===============================================
// readMaterial - read the ID given material from chaincode state
// ===============================================
func (t *Studio) readMaterial(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var ID string
	var err error

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting ID of the material to query")
	}

	ID = args[0]
	valAsbytes, err := stub.GetState(ID) //get the material from chaincode state
	if err != nil {
		return shim.Error("Failed to get state for " + ID + ": " + err.Error())
	} else if valAsbytes == nil {
		return shim.Error("Material does not exist: " + ID)
	}

	return shim.Success(valAsbytes)
}

// ============================================================
// getMaterialIndexList returns the materialIndexList from world state
// ============================================================
func (t *Studio) getMaterialIndexList(stub shim.ChaincodeStubInterface) pb.Response {
	materialIndexListBytes, err := stub.GetState("materialIndexList")
	if err != nil {
		return shim.Error("Failed to get material index list: " + err.Error())
	}
	return shim.Success(materialIndexListBytes)
}

// ===============================================
// getMaterialsByType - returns all materials of given type at materialIndexList
// ===============================================
func (t *Studio) getMaterialsByType(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	fmt.Println("- start query material by type")

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting type of the material to query")
	}

	typeIndex := args[0]

	// Retrieve the material index list from the world state
	materialIndexListBytes, err := stub.GetState("materialIndexList")
	if err != nil {
		return shim.Error("Failed to get material index list: " + err.Error())
	}

	// Unmarshal the material index list from bytes
	var materialIndexList []string
	err = json.Unmarshal(materialIndexListBytes, &materialIndexList)
	if err != nil {
		return shim.Error("Failed to unmarshal material index list: " + err.Error())
	}

	// Process each material ID in the index list
	var materials []Material
	var cont int = 0
	for _, compositeKey := range materialIndexList {
		_, composite_parts, err := stub.SplitCompositeKey(compositeKey)
		if err != nil {
			return shim.Error("Failed to split composite key: " + err.Error())
		}
		materialID := composite_parts[1]
		materialBytes, err := stub.GetState(materialID)
		if err != nil {
			return shim.Error("Failed to get material details: " + err.Error())
		}
		if materialBytes == nil {
			return shim.Error("Material not found: " + materialID)
		}

		// Unmarshal the material details
		var material Material
		err = json.Unmarshal(materialBytes, &material)
		if err != nil {
			return shim.Error("Failed to unmarshal material details: " + strconv.Itoa(cont) + err.Error())
		}

		// Check if the material type matches the requested type
		if material.Type == typeIndex {
			// Append the material to the list
			materials = append(materials, material)
		}
	}

	// Marshal the materials list to JSON
	materialsJSON, err := json.Marshal(materials)
	if err != nil {
		return shim.Error("Failed to marshal materials to JSON: " + err.Error())
	}

	fmt.Println("- end query material by type")
	return shim.Success(materialsJSON)
}

// ===============================================
// getAllMaterials - returns all materials avaible at MaterialsIndexList
// ===============================================
func (t *Studio) getAllMaterials(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println("- start query all materials")

	// Retrieve the material index list from the world state
	materialIndexListBytes, err := stub.GetState("materialIndexList")
	if err != nil {
		return shim.Error("Failed to get material index list: " + err.Error())
	}

	// Unmarshal the material index list from bytes
	var materialIndexList []string
	err = json.Unmarshal(materialIndexListBytes, &materialIndexList)
	if err != nil {
		return shim.Error("Failed to unmarshal material index list: " + err.Error())
	}

	// Process each material ID in the index list
	var materials []Material
	//var materials []string
	for _, compositeKey := range materialIndexList {
		_, composite_parts, err := stub.SplitCompositeKey(compositeKey)
		if err != nil {
			return shim.Error("Failed to split composite key: " + err.Error())
		}
		materialID := composite_parts[1]
		materialBytes, err := stub.GetState(materialID)
		if err != nil {
			return shim.Error("Failed to get material details: " + err.Error())
		}
		if materialBytes == nil {
			// Ignore if materialBytes is nil
			continue
		}

		// Unmarshal the material details
		var material Material
		err = json.Unmarshal(materialBytes, &material)
		if err != nil {
			return shim.Error("Failed to unmarshal material details: " + err.Error())
		}
		// Append the material to the list
		materials = append(materials, material)

	}

	// Marshal the materials list to JSON
	materialsJSON, err := json.Marshal(materials)
	if err != nil {
		return shim.Error("Failed to marshal materials to JSON: " + err.Error())
	}

	fmt.Println("- end query all materials")
	//return shim.Success(materialsJSON)
	return shim.Success(materialsJSON)
}

// ===============================================
// getNumberMaterialsByType - returns number of materials of given type at materialIndexList
// ===============================================
func (t *Studio) getNumberMaterialsByType(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	fmt.Println("- start query material by type")

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting type of the material to query")
	}

	typeIndex := args[0]

	// Retrieve the MaterialIndexList from the world state
	materialIndexListBytes, err := stub.GetState("materialIndexList")
	if err != nil {
		return shim.Error("Failed to get material index list: " + err.Error())
	}

	// Unmarshal the material index list from bytes
	var materialIndexList []string
	err = json.Unmarshal(materialIndexListBytes, &materialIndexList)
	if err != nil {
		return shim.Error("Failed to unmarshal material index list: " + err.Error())
	}

	// Initialize the count of materials
	numMaterials := 0

	for _, compositeKey := range materialIndexList {
		_, composite_parts, err := stub.SplitCompositeKey(compositeKey)
		if err != nil {
			return shim.Error("Failed to split composite key: " + err.Error())
		}
		materialType := composite_parts[0]
		// Check if the material type matches the requested type
		if materialType == typeIndex {
			numMaterials++
		}
	}

	// Marshal the number of materials to JSON
	numMaterialsJSON, err := json.Marshal(map[string]int{"num_materials": numMaterials})
	if err != nil {
		return shim.Error("Failed to marshal number of materials to JSON: " + err.Error())
	}

	fmt.Println("- end query material by type")
	return shim.Success(numMaterialsJSON)
}

// ===============================================
// getTotalNumberOfMaterials- returns total number of avaible materials on MaterialIndexList
// ===============================================
func (t *Studio) getTotalNumberOfMaterials(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println("- start query number of materials")

	// Retrieve the MaterialIndexList from the world state
	materialIndexListBytes, err := stub.GetState("materialIndexList")
	if err != nil {
		return shim.Error("Failed to get material index list: " + err.Error())
	}

	// Unmarshal the material index list from bytes
	var materialIndexList []string
	err = json.Unmarshal(materialIndexListBytes, &materialIndexList)
	if err != nil {
		return shim.Error("Failed to unmarshal material index list: " + err.Error())
	}

	// Get the total number of materials
	numMaterials := len(materialIndexList)

	// Marshal the number of materials to JSON
	numMaterialsJSON, err := json.Marshal(map[string]int{"num_materials": numMaterials})
	if err != nil {
		return shim.Error("Failed to marshal number of materials to JSON: " + err.Error())
	}

	fmt.Println("- end query number of materials")
	return shim.Success(numMaterialsJSON)
}

// ==================================================
// deleteMaterial - remove a material key/value pair from state
// ==================================================
func (t *Studio) deleteMaterial(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	fmt.Println("- start delete material")

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Material ID must be passed.")
	}

	materialID := args[0]

	// Get the material details from the world state
	valAsbytes, err := stub.GetState(materialID)
	if err != nil {
		return shim.Error("Failed to get state for " + materialID)
	} else if valAsbytes == nil {
		return shim.Error("Material does not exist: " + materialID)
	}

	var materialToDelete Material
	err = json.Unmarshal(valAsbytes, &materialToDelete)
	if err != nil {
		return shim.Error("Failed to decode JSON of: " + string(valAsbytes))
	}

	// Delete the material from state
	err = stub.DelState(materialID)
	if err != nil {
		return shim.Error("Failed to delete state:" + err.Error())
	}

	// Delete the material index from the index list
	indexName := "type~ID"
	typeIDIndexKey, err := stub.CreateCompositeKey(indexName, []string{materialToDelete.Type, materialToDelete.ID})
	if err != nil {
		return shim.Error(err.Error())
	}

	indexListBytes, err := stub.GetState("materialIndexList")
	if err != nil {
		return shim.Error("Failed to get material index list: " + err.Error())
	}

	var indexList []string
	err = json.Unmarshal(indexListBytes, &indexList)
	if err != nil {
		return shim.Error("Failed to unmarshal index list: " + err.Error())
	}

	for i, index := range indexList {
		if index == typeIDIndexKey {
			indexList = append(indexList[:i], indexList[i+1:]...)
			break
		}
	}

	updatedIndexListBytes, err := json.Marshal(indexList)
	if err != nil {
		return shim.Error("Failed to marshal updated index list: " + err.Error())
	}

	err = stub.PutState("materialIndexList", updatedIndexListBytes)
	if err != nil {
		return shim.Error("Failed to update index list: " + err.Error())
	}

	fmt.Println("- end delete material")
	return shim.Success(nil)
}

// ===============================================
// getAllMaterialsAndIndexList - returns all materials and index list
// ===============================================
func (t *Studio) getAllMaterialsAndIndexList(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println("- start get all materials and index list")

	// Retrieve the MaterialIndexList from the world state
	materialIndexListBytes, err := stub.GetState("materialIndexList")
	if err != nil {
		return shim.Error("Failed to get material index list: " + err.Error())
	}

	var materialIndexList []string
	err = json.Unmarshal(materialIndexListBytes, &materialIndexList)
	if err != nil {
		return shim.Error("Failed to unmarshal material index list: " + err.Error())
	}

	// Retrieve all materials
	var allMaterials []Material
	for _, compositeKey := range materialIndexList {
		_, composite_parts, err := stub.SplitCompositeKey(compositeKey)
		if err != nil {
			return shim.Error("Failed to split composite key: " + err.Error())
		}
		materialID := composite_parts[1]

		materialBytes, err := stub.GetState(materialID)
		if err != nil {
			return shim.Error("Failed to get material details: " + err.Error())
		}
		if materialBytes == nil {
			// Ignore if materialBytes is nil
			continue
		}

		// Unmarshal the material details
		var material Material
		err = json.Unmarshal(materialBytes, &material)
		if err != nil {
			return shim.Error("Failed to unmarshal material details: " + err.Error())
		}
		// Append the material to the list
		allMaterials = append(allMaterials, material)
	}

	// Combine materials list and index list into a single JSON response
	responseData := struct {
		Materials []Material `json:"materials"`
		IndexList []string   `json:"index_list"`
	}{
		Materials: allMaterials,
		IndexList: materialIndexList,
	}
	responseDataJSON, err := json.Marshal(responseData)
	if err != nil {
		return shim.Error("Failed to marshal response data to JSON: " + err.Error())
	}

	fmt.Println("- end get all materials and index list")
	return shim.Success(responseDataJSON)
}

//--------------------------------------------------------------------------------------------
// Funções wands

// ============================================================
// initWand - create a new wand, store into chaincode state
// ============================================================
func (t *Studio) initWand(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error

	// ==== Input sanitation ====
	fmt.Println("- start init wand")

	// Checks the correct number of arguments
	if len(args) < 6 {
		return shim.Error("Incorrect number of arguments. Expecting ID, Type, Color, Size and a list of Materials.")
	}

	// Extracting the arguments
	wandID := args[0]
	wandType := args[1]
	wandColor := args[2]
	wandSize, err := strconv.Atoi(args[3])
	if err != nil {
		return shim.Error("Size must be an integer.")
	}

	num_materials, err := strconv.Atoi(args[4])
	var materials []string
	if err != nil {
		return shim.Error("Erro ao converter num_materials para inteiro: " + err.Error())
	}

	// Loop que vai do 5 até 5 + num_materials
	for i := 5; i < 5+num_materials; i++ {
		materials = append(materials, args[i])
	}

	// Checks that the wand ID is not empty
	if wandID == "" {
		return shim.Error("Wand ID cannot be empty")
	}

	// Checks if wand with given ID already exists
	wandAsBytes, err := stub.GetState(wandID)
	if err != nil {
		return shim.Error("Failed to get wand: " + err.Error())
	} else if wandAsBytes != nil {
		fmt.Println("This wand already exists: " + wandID)
		return shim.Error("This wand already exists: " + wandID)
	}

	// Creates a Wand
	wand := &Wand{
		ObjectType: "Wand",
		ID:         wandID,
		Type:       wandType,
		Color:      wandColor,
		Size:       wandSize,
		Materials:  materials,
	}

	// Convert Wand structure to JSON
	wandJSONasBytes, err := json.Marshal(wand)
	if err != nil {
		return shim.Error("Error converting wand to JSON: " + err.Error())
	}

	// Save the wand in the world state
	err = stub.PutState(wandID, wandJSONasBytes)
	if err != nil {
		return shim.Error("Erro ao salvar a varinha no estado do world state: " + err.Error())
	}

	// Adds the wand ID to the wand index list

	// ==== Index the material to enable type-based range queries ====
	// The composite key is based on indexName~type~ID.
	// This will enable very efficient range queries based on composite keys matching indexName~type~*
	indexName := "type~ID"
	typeIndexKey, err := stub.CreateCompositeKey(indexName, []string{wand.Type, wand.ID})
	if err != nil {
		return shim.Error(err.Error())
	}

	wandsIndexListBytes, err := stub.GetState("wandsIndexList")
	if err != nil {
		return shim.Error("Error saving wand in world state: " + err.Error())
	}

	var wandsIndexList []string
	if wandsIndexListBytes != nil {
		err = json.Unmarshal(wandsIndexListBytes, &wandsIndexList)
		if err != nil {
			return shim.Error("Error decoding wand index list: " + err.Error())
		}
	}

	// Adds the wand ID to the wand index list
	wandsIndexList = append(wandsIndexList, typeIndexKey)
	wandsIndexListBytes, err = json.Marshal(wandsIndexList)
	if err != nil {
		return shim.Error("Error encoding wand index list: " + err.Error())
	}

	err = stub.PutState("wandsIndexList", wandsIndexListBytes)
	if err != nil {
		return shim.Error("Error when saving the list of wand indexes in the world state: " + err.Error())
	}

	// Deleting Materials ID from MaterialIndexList of world State

	// Retrieve the MaterialIndexList from the world state
	materialIndexListBytes, err := stub.GetState("materialIndexList")
	if err != nil {
		return shim.Error("Failed to get material index list: " + err.Error())
	}

	var materialIndexList []string
	err = json.Unmarshal(materialIndexListBytes, &materialIndexList)
	if err != nil {
		return shim.Error("Failed to unmarshal material index list: " + err.Error())
	}

	// Loop through each material in wand.Materials
	for _, materialID := range wand.Materials {
		// Check if the material ID exists in the material index list
		found := false
		for i, compositeKey := range materialIndexList {
			_, composite_parts, err := stub.SplitCompositeKey(compositeKey)
			if err != nil {
				return shim.Error("Failed to split composite key: " + err.Error())
			}
			listID := composite_parts[1]
			if listID == materialID {
				// Delete the material ID from the material index list
				materialIndexList = append(materialIndexList[:i], materialIndexList[i+1:]...)
				found = true
				break
			}
		}
		if !found {
			// If the material ID is not found in the material index list
			return shim.Error("Material ID not found in the material index list: " + materialID)
		}
	}

	// Marshal the updated material index list and save it to the world state
	updatedMaterialIndexListBytes, err := json.Marshal(materialIndexList)
	if err != nil {
		return shim.Error("Failed to marshal updated material index list: " + err.Error())
	}

	err = stub.PutState("materialIndexList", updatedMaterialIndexListBytes)
	if err != nil {
		return shim.Error("Failed to update material index list: " + err.Error())
	}

	// Returns a success message
	fmt.Println("- end init wand")
	return shim.Success(nil)
}

// ===============================================
// readWand - read the ID given wand from chaincode state
// ===============================================
func (t *Studio) readWand(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var ID string
	var err error

	fmt.Println("Checking the number of arguments...")
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting ID of the wand to query")
	}

	ID = args[0]
	valAsbytes, err := stub.GetState(ID) //get the wand from chaincode state
	if err != nil {
		return shim.Error("Failed to get state for " + ID + ": " + err.Error())
	} else if valAsbytes == nil {
		return shim.Error("Wand does not exist: " + ID)
	}

	return shim.Success(valAsbytes)
}

// ===============================================
// getWandsByType - returns all Wands of given type at wandsIndexList
// ===============================================
func (t *Studio) getWandsByType(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	fmt.Println("- Start query wands by type")

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting type of the wand to query")
	}

	typeIndex := args[0]

	// Retrieve the wandsIndexList from the world state
	fmt.Println("- Trying to get wandsIndexList")
	wandsIndexListBytes, err := stub.GetState("wandsIndexList")
	if err != nil {
		return shim.Error("Failed to get wands index list: " + err.Error())
	}

	var wandsIndexList []string
	err = json.Unmarshal(wandsIndexListBytes, &wandsIndexList)
	if err != nil {
		return shim.Error("Failed to unmarshal wands index list: " + err.Error())
	}

	// Process each wand ID in the index list
	var wands []Wand
	for _, compositeKey := range wandsIndexList {
		_, composite_parts, err := stub.SplitCompositeKey(compositeKey)
		if err != nil {
			return shim.Error("Failed to split composite key: " + err.Error())
		}
		wandID := composite_parts[1]
		wandBytes, err := stub.GetState(wandID)
		if err != nil {
			return shim.Error("Failed to get wand details: " + err.Error())
		}
		if wandBytes == nil {
			return shim.Error("Wand not found: " + wandID)
		}

		// Unmarshal the wand details
		var wand Wand
		err = json.Unmarshal(wandBytes, &wand)
		if err != nil {
			return shim.Error("Failed to unmarshal wand details: " + err.Error())
		}

		// Check if the wand type matches the requested type
		if wand.Type == typeIndex {
			// Append the wand to the list
			wands = append(wands, wand)
		}
	}
	// Marshal the wands list to JSON
	wandsJSON, err := json.Marshal(wands)
	if err != nil {
		return shim.Error("Failed to marshal wands to JSON: " + err.Error())
	}

	fmt.Println("- end query wand by type")
	return shim.Success(wandsJSON)
}

// ===============================================
// getAllWands - returns all wands available at wandsIndexList
// ===============================================
func (t *Studio) getAllWands(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println("- start query all available wands")

	fmt.Println("- Trying to get wandsIndexList")
	// Retrieve the wandsIndexList from the world state
	wandsIndexListBytes, err := stub.GetState("wandsIndexList")
	if err != nil {
		return shim.Error("Failed to get wands index list: " + err.Error())
	}

	var wandsIndexList []string
	err = json.Unmarshal(wandsIndexListBytes, &wandsIndexList)
	if err != nil {
		return shim.Error("Failed to unmarshal wands index list: " + err.Error())
	}

	// Process each wand ID in the index list
	var wands []Wand
	for _, compositeKey := range wandsIndexList {
		_, composite_parts, err := stub.SplitCompositeKey(compositeKey)
		if err != nil {
			return shim.Error("Failed to split composite key: " + err.Error())
		}
		wandID := composite_parts[1]
		wandBytes, err := stub.GetState(wandID)
		if err != nil {
			return shim.Error("Failed to get wand details: " + err.Error())
		}
		if wandBytes == nil {
			return shim.Error("Wand not found: " + wandID)
		}

		// Unmarshal the wand details
		var wand Wand
		err = json.Unmarshal(wandBytes, &wand)
		if err != nil {
			return shim.Error("Failed to unmarshal wand details: " + err.Error())
		}
		// Append the wand to the list
		wands = append(wands, wand)
	}

	// Marshal the wands list to JSON
	wandsJSON, err := json.Marshal(wands)
	if err != nil {
		return shim.Error("Failed to marshal wands to JSON: " + err.Error())
	}

	fmt.Println("- end query all available wands")
	return shim.Success(wandsJSON)
}

// ===============================================
// getNumberWandsByType - returns number of wands of given type at wandsIndexList
// ===============================================
func (t *Studio) getNumberwandsByType(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	fmt.Println("- start query number of wands by type")

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting type of the wands to query the number")
	}

	typeIndex := args[0]
	fmt.Println("- Trying to get wandsIndexList")
	// Retrieve the wandsIndexList from the world state
	wandsIndexListBytes, err := stub.GetState("wandsIndexList")
	if err != nil {
		return shim.Error("Failed to get wands index list: " + err.Error())
	}

	var wandsIndexList []string
	err = json.Unmarshal(wandsIndexListBytes, &wandsIndexList)
	if err != nil {
		return shim.Error("Failed to unmarshal wand index list: " + err.Error())
	}

	// Process each wand ID in the index list
	var num_wands int = 0
	for _, compositeKey := range wandsIndexList {
		_, composite_parts, err := stub.SplitCompositeKey(compositeKey)
		if err != nil {
			return shim.Error("Failed to split composite key: " + err.Error())
		}
		wandID := composite_parts[1]
		wandBytes, err := stub.GetState(wandID)
		if err != nil {
			return shim.Error("Failed to get wand details: " + err.Error())
		}
		if wandBytes == nil {
			return shim.Error("wand not found: " + wandID)
		}

		// Unmarshal the wand details
		var wand Wand
		err = json.Unmarshal(wandBytes, &wand)
		if err != nil {
			return shim.Error("Failed to unmarshal wand details: " + err.Error())
		}

		// Check if the wand type matches the requested type
		if wand.Type == typeIndex {
			// Increment the number of wands
			num_wands++
		}
	}

	// Marshal the number of wands to JSON
	numWandsJSON, err := json.Marshal(map[string]int{"num_wands": num_wands})
	if err != nil {
		return shim.Error("Failed to marshal number of wands to JSON: " + err.Error())
	}

	fmt.Println("- end query number of wands by type")
	return shim.Success(numWandsJSON)
}

// ===============================================
// getTotalNumberOfWandss- returns total number of avaible wands on wandIndexList
// ===============================================
func (t *Studio) getTotalNumberOfWands(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println("- start query number of wands")

	/*
		// Checks caller permission -Only Sr. olivares (org0) should be able to query wand list data-
		fmt.Println("Checking permissions")
		mspid, err := cid.GetMSPID(stub)
		if err != nil {
			return shim.Error("Error getting MSP ID: " + err.Error())
		}
		fmt.Printf("- mspid: %s\n", mspid)
		// Checks if the MSP ID matches the allowed organization (org0)
		if mspid != "Org0MSP" {
			return shim.Error("Only members of organization 0 (Sr. Orlivaras) can perform this function")
		}
	*/

	// Retrieve the wandsIndexList from the world state
	fmt.Println("- Trying to get wandsIndexList")
	wandsIndexListBytes, err := stub.GetState("wandsIndexList")
	if err != nil {
		return shim.Error("Failed to get wand index list: " + err.Error())
	}

	var wandsIndexList []string
	err = json.Unmarshal(wandsIndexListBytes, &wandsIndexList)
	if err != nil {
		return shim.Error("Failed to unmarshal wand index list: " + err.Error())
	}

	// Count the number of wands in the index list
	num_wands := len(wandsIndexList)

	// Marshal the number of wands to JSON
	numWandsJSON, err := json.Marshal(map[string]int{"num_wands": num_wands})
	if err != nil {
		return shim.Error("Failed to marshal number of wands to JSON: " + err.Error())
	}

	fmt.Println("- end query number of wands")
	return shim.Success(numWandsJSON)
}

// ==================================================
// deleteWand - remove a wand key/value pair from state
// ==================================================
func (t *Studio) deletewand(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	fmt.Println("- start delete Wand")

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Wand ID must be passed.")
	}

	wandID := args[0]

	// Get the wand details from the world state
	valAsbytes, err := stub.GetState(wandID)
	if err != nil {
		return shim.Error("Failed to get state for " + wandID + ": " + err.Error())
	} else if valAsbytes == nil {
		return shim.Error("Wand does not exist: " + wandID)
	}

	var wandToDelete Wand
	err = json.Unmarshal(valAsbytes, &wandToDelete)
	if err != nil {
		return shim.Error("Failed to decode JSON of: " + string(valAsbytes) + ": " + err.Error())
	}

	// Delete the wand from state
	err = stub.DelState(wandID)
	if err != nil {
		return shim.Error("Failed to delete state:" + err.Error())
	}

	// Delete the wand index from the index list
	indexName := "type~ID"
	typeIDIndexKey, err := stub.CreateCompositeKey(indexName, []string{wandToDelete.Type, wandToDelete.ID})
	if err != nil {
		return shim.Error(err.Error())
	}
	indexListBytes, err := stub.GetState("wandsIndexList")
	if err != nil {
		return shim.Error("Failed to get wand index list: " + err.Error())
	}

	var indexList []string
	err = json.Unmarshal(indexListBytes, &indexList)
	if err != nil {
		return shim.Error("Failed to unmarshal index list: " + err.Error())
	}

	// Remove the wand ID from the index list
	var updatedIndexList []string
	for i, index := range indexList {
		if index == typeIDIndexKey {
			updatedIndexList = append(indexList[:i], indexList[i+1:]...)
			break
		}
	}

	// Marshal the updated index list
	updatedIndexListBytes, err := json.Marshal(updatedIndexList)
	if err != nil {
		return shim.Error("Failed to marshal updated index list: " + err.Error())
	}

	// Update the index list in the world state
	err = stub.PutState("wandsIndexList", updatedIndexListBytes)
	if err != nil {
		return shim.Error("Failed to update index list: " + err.Error())
	}

	// Deleting each material of Materials list from Worldstate
	for _, materialID := range wandToDelete.Materials {
		// Delete the material from state
		err = stub.DelState(materialID)
		if err != nil {
			return shim.Error("Failed to delete state:" + err.Error())
		}
	}

	fmt.Println("- end delete Wand")
	return shim.Success(nil)
}

func main() {
	err := shim.Start(&Studio{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Exiting Simple chaincode: %s", err)
		os.Exit(2)
	}
}
