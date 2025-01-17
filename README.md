# Sr_Olivaras_Store
Hyperledger Fabric blockchain Smart Contract : Experimental smart contract for a Hyperledger Fabric blockchain network in a fictional item manufacturing network, managing the origin, registration,  usage of raw materials in stock, inventory management of items, and tracking data of sold items


**Brief Introduction to the Challenge Proposal:**
The challenge involves creating a traceability system for the magic wands sold at Mr. Ollivander's shop. This system should enable wizards in the wand production chain to list the materials they have available. Consequently, Mr. Ollivander, using the system, can view the materials available for wand production as well as add newly created wands to the database. Furthermore, by securely and traceably storing data through blockchain technology, the system will provide customers with information regarding the wand's origin and the materials it is made from.

# Main Challenge Report

### Brief Introduction to the Challenge Proposal
The challenge involves creating a traceability system for the magic wands sold at Mr. Ollivander's shop. This system should enable wizards in the wand production chain to list the materials they have available. Consequently, Mr. Ollivander, using the system, can view the materials available for wand production as well as add newly created wands to the database. Furthermore, by securely and traceably storing data through blockchain technology, the system will provide customers with information regarding the wand's origin and the materials it is made from.

---

## 1. Approach Adopted as a Solution to the Proposed Problem

### 1.1 Data Organization Structure in the World State
To store material data in the World State of the Ledger, two structs were created: `Wand` and `Material`. Additionally, the World State contains two Type~ID key lists: `materialIndexList` and `wandsIndexList`.

#### a) Material Struct
```go
struct Material {
  ObjectType string `json:"docType"`
  ID string `json:"ID"` // Unique ID
  Type string `json:"type"` // Material type
  Supplier string `json:"supplier"` // Supplier
}
```
- **Material Identification**: Each material is uniquely identified in the World State by its ID. This ID must remain unique throughout the material's lifecycle. 
  - If the material is deleted or used as part of a wand, the ID cannot be reused unless the material is explicitly removed using the `deleteMaterial` function or if the wand it belongs to is deleted. This ensures that all materials, whether available or used in wands, can be traced within the system.
- **Attributes**:
  - `Type`: Identifies the material type.
  - `Supplier`: Tracks the wizard responsible for supplying the material, enabling traceability.

#### b) Wand Struct
```go
struct Wand {
  ObjectType string `json:"docType"`
  ID string `json:"ID"` // Unique ID
  Type string `json:"type"` // Wand type
  Color string `json:"color"` // Wand color
  Size int `json:"size"` // Wand size
  Materials []string `json:"Materials"` // List of material IDs used
}
```
- **Wand Identification**: Each wand is uniquely identified by its ID in the World State.
- **Attributes**:
  - `Type`, `Color`, `Size`: Key properties of the wand.
  - `Materials`: A list of material IDs used in the wand's construction, enabling customers to trace the origin of each component.

#### c) World State Lists
To organize and query materials and wands effectively, the World State includes:
- `materialIndexList`: Stores Type~ID pairs for all available materials.
- `wandsIndexList`: Stores Type~ID pairs for all wands in the system.

---

### 1.2 Available Functions for Materials and Wands Management

#### Basic Functions Required by the Challenge
- **Materials:**
  - `initMaterial(ID, Type, Supplier)`: Creates a material.
  - `readMaterial(ID)`: Retrieves material details.
  - `getAllMaterials()`: Returns all materials available for wand production.
  - `getTotalNumberOfMaterials()`: Returns the total number of available materials.
  - `deleteMaterial(ID)`: Deletes a material from the World State.

- **Wands:**
  - `initWand(ID, Type, Color, Size, MaterialCount, Material1, Material2, ...)`: Creates a wand.
  - `readWand(ID)`: Retrieves wand details.
  - `getAllWands()`: Returns all wands in the system.
  - `getTotalNumberOfWands()`: Returns the total number of wands.
  - `deleteWand(ID)`: Deletes a wand from the World State.

#### Additional Functions
- `getMaterialIndexList()`: Returns the Type~ID list for materials.
- `getMaterialsByType(Type)`: Retrieves materials of a specific type.
- `getNumberMaterialsByType(Type)`: Returns the number of materials of a specific type.
- `getAllMaterialsAndIndexList()`: Returns data and the index list for all materials.
- `getWandsByType(Type)`: Retrieves wands of a specific type.
- `getNumberWandsByType(Type)`: Returns the count of wands of a specific type.

---

## 2. System Execution Instructions

### 2.1 Initializing Minifabric
1. Ensure Docker is installed.
2. Download the Minifabric tool and execute the following command to initialize the default network:
   ```bash
   sudo ./minifab up
   ```

### 2.2 Installing Chaincode Studio on Minifab
1. Create a directory for chaincode files:
   ```bash
   sudo mkdir -p $(pwd)/vars/chaincode/Studio/go
   ```
2. Manually move all files from the "Studio" folder (provided with this report) to the above directory using `sudo`.
3. Install the chaincode:
   ```bash
   sudo ./minifab ccup -n Studio -l go -v 1.0
   ```
   If an error occurs, try:
   ```bash
   sudo ./minifab ccup -v 2.0 -n Studio -l go
   ```

### 2.3 Running Functions for Wands and Materials Management
Refer to the Minifabric documentation for invoking chaincode methods: [Minifabric Docs](https://github.com/hyperledger-labs/minifabric/blob/main/docs/README.md#invoke-chaincode-methods)

Command format:
```bash
sudo ./minifab invoke -n Studio -p '"methodname","p1","p2",...'
```
Replace:
- `methodname` with the function name.
- `p1`, `p2`, etc., with the required arguments.

---

## 3. Execution Example
- **Example:** `initMaterial`
  - After creation, the system displays the current list of available materials.
- **Example:** `getAllWands`
  - The system returns JSON-formatted wand data, including IDs and attributes.

---

## 4. References
The following references were essential for implementing the solution and understanding the theoretical foundations of the system:

- **[Hyperledger Fabric Documentation](https://hyperledger-fabric.readthedocs.io):** Provided comprehensive details on the core concepts of Hyperledger Fabric, enabling the design and integration of blockchain technology in the system.

- **Videos from Hyperledger Foundation's YouTube Channel:** Demonstrated the practical usage of Minifabric and its functionalities, which were crucial for network initialization and chaincode deployment.

- **"Supply Chain Using Hyperledger Fabric" Thesis by Maria-Isavella G. Manolaki-Sempagios:** Offered insights into designing supply chain systems using Hyperledger Fabric, which informed the design of the World State and traceability structures.

- **Article: "Writing Chaincode in Golang - the OOP way" by Vishal:** Provided practical guidance on implementing chaincode with object-oriented programming principles in Golang, which was applied to struct and function development.

- **Minifabric's "privatemarbles" Code Example:** Served as a reference for implementing private data collections and chaincode functionality, facilitating the development of secure and efficient data storage mechanisms.

These resources collectively ensured the robust design and functionality of the traceability system while adhering to best practices in blockchain development.

