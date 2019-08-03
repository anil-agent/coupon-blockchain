package main

import (
    "time"
    "encoding/json"
    "fmt"
	"strings"
	"strconv"
    "bytes"
    "github.com/hyperledger/fabric/core/chaincode/shim"
     sc "github.com/hyperledger/fabric/protos/peer"
)

type CouponChaincode struct {

}

type QueryRecord struct{
	RecordType 			string 					 `json:"recordType"`
}

type QueryKey struct{
	Key 				string 					 `json:"key"`
}

type Coupon struct {	
    Key                 string           		 `json:"key"`
    Name                string              	 `json:"name"`
    CreatedDateTime     string            		 `json:"createdDateTime"`
    ExpiresOn           string           		 `json:"expiresOn"`   	
    DiscountAmount      float64              	 `json:"discountAmount,string"`
    RevenueSharePercent float64              	 `json:"revenueSharePercent,string"`
    Status              string               	 `json:"status"` 
    CustomerKey         string               	 `json:"customerKey"` 
}

type CouponResponse struct {
	Key 				string 					  `json:"key"`
	Coupon 				Coupon 					  `json:"record"`
}

type Customer struct {
    Key                 string               	 `json:"key"`
    Name                string               	 `json:"name"`
    Email               string              	 `json:"email"`
}

type Partner struct {
    Key                 string                   `json:"key"`
    Name                string                   `json:"name"`
    AddressKey          string                   `json:"addressKey"`
}

type Address struct {
    Key                 string              	 `json:"key"`
    Street              string              	 `json:"street"`
    ZipCode             string              	 `json:"zipCode"`
    State               string              	 `json:"state"`
    Country             string              	 `json:"country"`
}

type ValidateCouponRequest struct {
    CouponKey          	string                	 `json:"couponKey"`
    CustomerKey         string               	 `json:"customerKey"`
}

type ValidateCouponResponse struct {
	IsValid          	bool                	 `json:"isValid"`
	Message				string			    	 `json:"message"`
}

type RedeemCouponRequest struct { 
    AssetOriginalPrice  float64             	 `json:"assetOriginalPrice,string"`
    CouponKey           string              	 `json:"couponKey"`
    PartnerKey          string               	 `json:"partnerKey"`
}

type RedeemCouponResponse struct { 
    SalesTransaction 	SalesTransaction 	 	`json:"salesTransaction"`
}

type CustomerCoupon struct {
    Customer            Customer             	`json:"customer"`
    Coupons             []Coupon             	`json:"coupons"`
}

type SalesTransaction struct { 
    Key                 string                	 `json:"key,omitempty"`
    PartnerKey          string                	 `json:"partnerKey"`
    CouponKey           string               	 `json:"couponKey"`
    AssetOriginalPrice  float64              	 `json:"assetOriginalPrice,string"`
    SalesAmount         float64              	 `json:"salesAmount,string"`
    RevenueShareAmount  float64              	 `json:"revenueShareAmount,string"`
    SettlementAmount    float64              	 `json:"settlementAmount,string"`
}

var (
    errResponse     string
    shimResponse    string 
)

const (
	couponKeyPrefix = "coupon"
	salesTransactionKeyPrefix = "salestransaction"
	couponRangeStartKey = "couponrangestartkey"
	couponRangeEndKey = "couponrangeendkey"
	customerRangeStartKey = "customerrangestartkey"
	customerRangeEndKey = "customerrangeendkey"
	partnerRangeStartKey = "partnerrangestartkey"
    partnerRangeEndKey = "partnerrangeendkey"
    salesTransactionRangeStartKey = "salesTransactionrangestartkey"
    salesTransactionRangeEndKey = "salesTransactionrangendkey"
	dateFormat = "02-01-2006"
	couponStatusIssued = "ISSUED"
	couponStatusRedeemed = "REDEEMED"
)

// Init is called during the smart contract instantiation .
func (t *CouponChaincode) Init(stub shim.ChaincodeStubInterface) sc.Response  {
    //initiating the ledger 
    t.initCustomers(stub)
    t.initPartners(stub)
    t.initAddresses(stub) 
	t.initRangeKeys(stub)
	return shim.Success(nil)
}

// Invoke is called to update or query the ledger in a  transaction proposal.
func (c *CouponChaincode) Invoke(stub shim.ChaincodeStubInterface) sc.Response {
    fnc, args := stub.GetFunctionAndParameters()
    // Route to the appropriate handler function to interact with the ledger appropriately
    switch(strings.ToLower(fnc)) {
    case "createcoupon" :
		return c.CreateCoupon(stub, args)
	case "createsalestransaction" :
		return c.CreateSalesTransaction(stub, args)
    case "querybykey" :
        return c.QueryByKey(stub, args)
    case "querybyrange" :
		return c.QueryByRange(stub, args)
	case "validatecoupon" :
        return c.ValidateCoupon(stub, args)
	case "redeemcoupon" :
		return c.RedeemCoupon(stub, args)
	case "deleterecord" :
		return c.DeleteRecord(stub, args)
	case "queryhistorybykey" :
		return c.QueryHistoryByKey(stub, args)
	case "querycouponsbycustomer" :
		return c.QueryCouponsByCustomer(stub, args)
    default: 
        return shim.Error(fmt.Sprintf("Invalid ChainCode Function : %s", fnc))
    }
}

//Function to get Record by Key
func (c *CouponChaincode) QueryByKey(stub shim.ChaincodeStubInterface, args []string) sc.Response {
	var queryKey QueryKey
	json.Unmarshal([]byte (args[0]), &queryKey)
	key := strings.ToLower(queryKey.Key)
	resultAsBytes , err := stub.GetState(key)
	if err != nil || resultAsBytes == nil {
		return shim.Error(fmt.Sprintf("QueryByKey failed for Key : %s error : %s",key ,err.Error()))   
	} 
	return shim.Success(resultAsBytes)
}

//Get Result by Query
func (c *CouponChaincode) QueryByRange(stub shim.ChaincodeStubInterface, args []string) sc.Response {
	var startKeyAsBytes []byte
	var endKeyAsBytes []byte
	var record QueryRecord
	json.Unmarshal([]byte (args[0]), &record)
    switch(strings.ToLower(record.RecordType)) {
    case "coupon" :
		startKeyAsBytes, _  = stub.GetState(couponRangeStartKey)
		endKeyAsBytes, _    = stub.GetState(couponRangeEndKey)
	case "customer" : 
		startKeyAsBytes, _  = stub.GetState(customerRangeStartKey)
		endKeyAsBytes, _    = stub.GetState(customerRangeEndKey)
    case "salestransaction" :
        startKeyAsBytes, _  = stub.GetState(salesTransactionRangeStartKey)
		endKeyAsBytes, _    = stub.GetState(salesTransactionRangeEndKey)
	case "partner" :
        startKeyAsBytes, _  = stub.GetState(partnerRangeStartKey)
		endKeyAsBytes, _    = stub.GetState(partnerRangeEndKey)
    default: 
		return shim.Error(fmt.Sprintf("Invalid Entity Type : %s  ARGS: %s", record.RecordType, args[0]))
	}
	startRangeKey := string(startKeyAsBytes)
	endRangeKey := string(endKeyAsBytes)
	outboundEndKey  := strings.Split(endRangeKey, ":")
	outboundEndKeyNumber , _ :=  strconv.Atoi(outboundEndKey[1])
	resultByte, err := getStatebyRangeResult(stub,startRangeKey, getKeyByRecordType(outboundEndKey[0], outboundEndKeyNumber + 1))
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(resultByte)
}

// Function to create record
func (c *CouponChaincode) CreateCoupon(stub shim.ChaincodeStubInterface, args []string) sc.Response {
    
    resultAsBytes, err := stub.GetState(couponRangeEndKey);
    if err != nil {
        return shim.Error(fmt.Sprintf("CouponRangeEndKey %s GetState failed : %s", couponRangeEndKey , err.Error()))
	}
	newRecordKey , keyNumber := generateKey(string(resultAsBytes))
    writeErr := stub.PutState(newRecordKey, []byte (args[0]))
    if writeErr != nil {
        return shim.Error(fmt.Sprintf("Coupon %s PutState failed: %s", newRecordKey, writeErr.Error()))
	}
	writeErr = stub.PutState(couponRangeEndKey, []byte(getKeyByRecordType(couponKeyPrefix,keyNumber)))
	if writeErr != nil {
	   return shim.Error(fmt.Sprintf("CouponRangeEndKey %s PutState failed : %s", couponRangeEndKey, writeErr.Error()))
    }
	return shim.Success([]byte (fmt.Sprintf("%s created successfully", newRecordKey)))
}

// Function to create record
func (c *CouponChaincode) CreateSalesTransaction(stub shim.ChaincodeStubInterface, args []string) sc.Response {
     resultAsBytes, err := stub.GetState(salesTransactionRangeEndKey);
    if err != nil {
        return shim.Error(fmt.Sprintf("SalesTransactionRangeEndKey %s GetState failed : %s", salesTransactionRangeEndKey , err.Error()))
	}
    newRecordKey , keyNumber:= generateKey(string(resultAsBytes))
    writeErr := stub.PutState(newRecordKey, []byte (args[0]))
    if writeErr != nil {
        return shim.Error(fmt.Sprintf("SalesTransaction %s PutState failed: %s", newRecordKey, writeErr.Error()))
	}
	writeErr = stub.PutState(salesTransactionRangeEndKey, []byte(getKeyByRecordType(salesTransactionKeyPrefix , keyNumber)))
	if writeErr != nil {
	   return shim.Error(fmt.Sprintf("SalesTransactionRangeEndKey %s PutState failed : %s", salesTransactionRangeEndKey, writeErr.Error()))
    }
	return shim.Success([]byte (fmt.Sprintf("%s created successfully", newRecordKey)))
}

//Function to delete record
func (c *CouponChaincode) DeleteRecord(stub shim.ChaincodeStubInterface, args []string) sc.Response {
	var queryKey QueryKey
	json.Unmarshal([]byte (args[0]), &queryKey)
	deleteKey := strings.ToLower(queryKey.Key)
	keyValue := strings.Split(deleteKey, ":")
	keyTextValue := keyValue[0]
	keyNumberValue, _ := strconv.Atoi(keyValue[1])
	// Delete the key
	delErr := stub.DelState(deleteKey)
	if delErr != nil {
		return shim.Error(fmt.Sprintf("Failed to delete record %s error: %s", args[0], delErr.Error()))
	} else {
		isRecordTypeFound, key :=  getRecordRangeEndKey(keyTextValue)
		if isRecordTypeFound {
			//If the deleted record is the last record, then update the end range keys
		    latestRecordRangeEndKey := key
			resultAsBytes, err := stub.GetState(latestRecordRangeEndKey);
			if err != nil {
				return shim.Error(fmt.Sprintf("Failed to get Range End Keys for %s error: %s", keyTextValue, err.Error()))
			}
			latestRecordRangeEndKeyValue := string(resultAsBytes)
			value := strings.Split(latestRecordRangeEndKeyValue, ":")
			latestRecordRangeEndKeyNumberValue, err := strconv.Atoi(value[1])
			if keyNumberValue == latestRecordRangeEndKeyNumberValue {
				latestRecordRangeEndKeyNumberValue -= 1 
			}
			revisedEndRangeKey := getKeyByRecordType(keyTextValue , latestRecordRangeEndKeyNumberValue)
			writeErr := stub.PutState(key, []byte(revisedEndRangeKey))
			if writeErr != nil {
				return shim.Error(fmt.Sprintf("Unable to write %s end key to the ledger. error : %s" , revisedEndRangeKey, writeErr.Error()))
			}
		} 
	}
	return shim.Success([]byte ("Deleted record "+ deleteKey))
}

//Function to validate coupon
func (c *CouponChaincode) ValidateCoupon(stub shim.ChaincodeStubInterface,args []string) sc.Response {
	
	var  validateCouponRequest ValidateCouponRequest
	var  validateCouponResponse ValidateCouponResponse
	json.Unmarshal([]byte(args[0]), &validateCouponRequest)
	resultAsBytes, err := stub.GetState(validateCouponRequest.CouponKey);
	if err != nil || resultAsBytes == nil{
		return shim.Error(fmt.Sprintf("Unable to fetch coupon %s error : %s" , validateCouponRequest.CouponKey , err.Error()))
	}
	var coupon Coupon
	json.Unmarshal(resultAsBytes, &coupon)
	if coupon.CustomerKey != validateCouponRequest.CustomerKey {
		validateCouponResponse.IsValid = false
		validateCouponResponse.Message = fmt.Sprintf("Invalid Coupon : %s for Customer : %s" ,  validateCouponRequest.CouponKey , validateCouponRequest.CustomerKey)
		result , _ := json.Marshal(validateCouponResponse)
		return shim.Success(result)
	}
	if coupon.Status != couponStatusIssued {
		validateCouponResponse.IsValid = false
		validateCouponResponse.Message = fmt.Sprintf("Invalid Coupon status : %s", coupon.Status)
		result, _ := json.Marshal(validateCouponResponse)
		return shim.Success(result)
	}
	expiryDate, err := time.Parse(dateFormat, coupon.ExpiresOn)
	if err != nil {
		return shim.Error(fmt.Sprintf("Invalid Coupon expiry date : %s", expiryDate))
	}
	if hasCouponExpired(expiryDate) {
		validateCouponResponse.IsValid = false
		validateCouponResponse.Message = fmt.Sprintf("Coupon %s has expired!!! ",  validateCouponRequest.CouponKey )
		result, _ := json.Marshal(validateCouponResponse)
		return shim.Success(result)
	}
	validateCouponResponse.IsValid = true
	validateCouponResponse.Message = fmt.Sprintf("Valid Coupon %s!!!", validateCouponRequest.CouponKey)
	result, _ := json.Marshal(validateCouponResponse)
	return shim.Success(result)
}

func (c *CouponChaincode) RedeemCoupon(stub shim.ChaincodeStubInterface, args []string) sc.Response {
	
	var redeemCouponRequest RedeemCouponRequest
	json.Unmarshal([]byte(args[0]), &redeemCouponRequest)
	//Get Coupon Information based on CouponKey
	resultAsBytes, err := stub.GetState(redeemCouponRequest.CouponKey);
	if err != nil {
		return shim.Error(fmt.Sprintf("Unable to fetch coupon %s error : %s", redeemCouponRequest.CouponKey,  err.Error()))
	}
	coupon := Coupon{}
	json.Unmarshal(resultAsBytes, &coupon)
	//Get Partner Information based on PartnerKey
	resultAsBytes, err = stub.GetState(redeemCouponRequest.PartnerKey);
	if err != nil {
		return shim.Error(fmt.Sprintf("Unable to fetch partner %s error : %s", redeemCouponRequest.PartnerKey, err.Error()))
	}
	partner := Partner{}
	json.Unmarshal(resultAsBytes, &partner)
	salesTransaction := prepSalesTransaction(redeemCouponRequest, coupon)
	salesTransactionAsBytes, err := json.Marshal(salesTransaction)
	if err != nil {
		return shim.Error(err.Error())
	}
	salesTransactionArgs  := []string{ string(salesTransactionAsBytes) }
	response := c.CreateSalesTransaction(stub, salesTransactionArgs)
	if response.Status != 200 {
		return shim.Error(response.Message)
	}
	//update coupon status to redeemed
	coupon.Status = couponStatusRedeemed
	couponAsBytes, err := json.Marshal(coupon)
	writeErr := stub.PutState(redeemCouponRequest.CouponKey, couponAsBytes)
	if writeErr != nil {
		return shim.Error(fmt.Sprintf("Redeem Coupon %s save failed error : %s", redeemCouponRequest.CouponKey, writeErr.Error()))
	}
	return shim.Success([]byte ("Coupon Redeemed Sucessfully!!!"))
}

//Function to query coupons based on customer
func (c *CouponChaincode) QueryCouponsByCustomer(stub shim.ChaincodeStubInterface, args []string) sc.Response {
	var  couponResponse []CouponResponse
	var queryKey QueryKey
	customerCoupons := make([]Coupon, 0)
	json.Unmarshal([]byte (args[0]), &queryKey)
	customerKey := strings.ToLower(queryKey.Key)
	couponQuery := "{\"recordType\":\"coupon\"}";
	couponQueryResponse:= c.QueryByRange(stub, []string { couponQuery })
	if couponQueryResponse.Status != 200 {
		return shim.Error(couponQueryResponse.Message)
	}
	couponQueryResponseAsString := string(couponQueryResponse.Payload)
	json.Unmarshal([]byte(couponQueryResponseAsString), &couponResponse)

	for i := 0; i < len(couponResponse); i++ {
		if couponResponse[i].Coupon.CustomerKey == customerKey {
			couponResponse[i].Coupon.Key = couponResponse[i].Key
			customerCoupons = append(customerCoupons, couponResponse[i].Coupon)
		}
	}
	customerCouponsAsBytes, _ := json.Marshal(customerCoupons)
	return shim.Success(customerCouponsAsBytes)
}

//Function to get History for a key 
func (c *CouponChaincode) QueryHistoryByKey(stub shim.ChaincodeStubInterface, args []string) sc.Response {
	var queryKey QueryKey
	json.Unmarshal([]byte (args[0]), &queryKey)
	key := strings.ToLower(queryKey.Key)
	resultsIterator, err := stub.GetHistoryForKey(key)
	if err != nil {
		return shim.Error(fmt.Sprintf("Error %s while searching for key %s ", err.Error(), queryKey.Key))
	}
	historyForKey , err := generateHistoricalRecordsForKey(resultsIterator)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(historyForKey)
}


//Main Function
func main () {
	err := shim.Start(new(CouponChaincode))
    if err != nil {
        fmt.Printf("Instantiation of CouponChainCode Failed : %s", err)
    }
}

// Function to create the sales transaction
func prepSalesTransaction(redeemCouponRequest RedeemCouponRequest, coupon Coupon) SalesTransaction {
	salesAmount := redeemCouponRequest.AssetOriginalPrice - coupon.DiscountAmount
	revenueShareAmount := redeemCouponRequest.AssetOriginalPrice  * (coupon.RevenueSharePercent / 100)
	settlementAmount := salesAmount - revenueShareAmount
	salesTransaction := SalesTransaction { 
		PartnerKey: redeemCouponRequest.PartnerKey,
		CouponKey: redeemCouponRequest.CouponKey, 
		AssetOriginalPrice : redeemCouponRequest.AssetOriginalPrice, 
		SalesAmount : salesAmount, 
		RevenueShareAmount: revenueShareAmount, 
		SettlementAmount: settlementAmount,
	}
	return salesTransaction 
}

// Function to validate the coupon by date
func hasCouponExpired(expiryDate time.Time) bool {
	currentTime := time.Now()
	today := time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(),0,0,0,0, time.UTC)
	if today.After(expiryDate) || today.Equal(expiryDate) {
		return false
	}
	return true
}

//Function to get range end key
func getRecordRangeEndKey(recordType string) (bool , string) {
	 var isRecordTypeFound bool
	 var recordRangeEndKey string 
	 switch (recordType) {
	 case "coupon" :
		recordRangeEndKey = couponRangeEndKey
		isRecordTypeFound = true
	 case "salestransaction" :
		recordRangeEndKey = salesTransactionRangeEndKey
		isRecordTypeFound = true
	 case "partner" :
		recordRangeEndKey = partnerRangeEndKey
		isRecordTypeFound = true
	default :
		recordRangeEndKey = fmt.Sprintf("Unknown Record Type : %s", recordType)
		isRecordTypeFound = false
	 }
	 return isRecordTypeFound, recordRangeEndKey
}

//Function to get key based on record type
func getKeyByRecordType(recordType string, keyNumber int) string { 
	return  recordType + ":" + strconv.Itoa(keyNumber)
}

//Function to get key based on record type
func generateKey(endRangeKey string) ( string , int ) { 
    result := strings.Split(endRangeKey, ":")
    keyNumber, _ := strconv.Atoi(result[1])
    if  keyNumber == 0 {
        keyNumber = 101
    } else {
        keyNumber += 1
	}
	key := getKeyByRecordType(result[0] , keyNumber)
	return key , keyNumber
}

//Function to get result based on range
func getStatebyRangeResult(stub shim.ChaincodeStubInterface, startRangeKey string, endRangeKey string) ([] byte, error) {
	//Get state by range
	resultsIterator, err := stub.GetStateByRange(startRangeKey, string(endRangeKey))
    defer resultsIterator.Close()
    if err != nil {
        return nil, err
	}
    // buffer is a JSON array containing QueryRecords
    var buffer bytes.Buffer
    buffer.WriteString("[")
    bArrayMemberAlreadyWritten := false
    for resultsIterator.HasNext() {
        queryResponse,
        err := resultsIterator.Next()
        if err != nil {
            return nil, err
        }
        // Add a comma before array members, suppress it for the first array member
        if bArrayMemberAlreadyWritten == true {
            buffer.WriteString(",")
        }
        buffer.WriteString("{\"Key\":")
        buffer.WriteString("\"")
        buffer.WriteString(queryResponse.Key)
        buffer.WriteString("\"")
        buffer.WriteString(", \"Record\":")
        // Record is a JSON object, so we write as-is
        buffer.WriteString(string(queryResponse.Value))
        buffer.WriteString("}")
        bArrayMemberAlreadyWritten = true
    }
    buffer.WriteString("]")
    return buffer.Bytes(), nil
}

//Function to generate the historical records for key 
func generateHistoricalRecordsForKey(resultsIterator shim.HistoryQueryIteratorInterface) ([]byte, error){
	defer resultsIterator.Close()

	// buffer is a JSON array containing the historic values 
	var buffer bytes.Buffer
	buffer.WriteString("[")
	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		response, err := resultsIterator.Next()
		if err != nil {
			return nil,err
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"TxId\":")
		buffer.WriteString("\"")
		buffer.WriteString(response.TxId)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Value\":")
		// if it was a delete operation on given key, then we need to set the
		//corresponding value null. Else, we will write the response.Value
		//as-is (as the Value itself a JSON marble)
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
		buffer.WriteString("\"")

		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")
	return buffer.Bytes(),nil
}
//Function to initiate ledger with sample Customers
func (t *CouponChaincode) initCustomers(stub shim.ChaincodeStubInterface) {
    //initiating the ledger with customers
    customers := []Customer {
        Customer{ Key : "customer:101", Name:"Louis", Email:"louis@gmail.com" }, 
        Customer{ Key : "customer:102", Name:"Elizabeth",  Email:"eizabeth@gmail.com" },
		Customer{ Key : "customer:103", Name:"Henry",  Email:"henry@outlook.com" },
	}
    for c := 0; c < len(customers); c++ {
        customerAsBytes, _ := json.Marshal(customers[c])
        stub.PutState(customers[c].Key, customerAsBytes)
    }
}

//Function to initiate ledger with sample Partners
func (t *CouponChaincode) initPartners(stub shim.ChaincodeStubInterface) {
    //initiating the ledger with partners
	partner := Partner{ Key: "partner:101",  Name: "Govberg Jewelers Suburban Square", AddressKey : "address:101"}
	partnerAsBytes, _ := json.Marshal(partner)
	stub.PutState(partner.Key,  partnerAsBytes)
}

//Function to initiate ledger with sample Addresses
func (t *CouponChaincode) initAddresses(stub shim.ChaincodeStubInterface) {
    //initiating the ledger with addresses
    address := Address{ Key : "address:101", Street:"65, St James Place", ZipCode:"19003", State:"Pennsylvania", Country: "USA"}
 	addressAsBytes, _ := json.Marshal(address)
   	stub.PutState(address.Key, addressAsBytes)
}

//Function to initiate ledger with Range Keys
func (t *CouponChaincode) initRangeKeys(stub shim.ChaincodeStubInterface) {
	stub.PutState(customerRangeStartKey, []byte("customer:101"))
	stub.PutState(customerRangeEndKey, []byte("customer:103"))
	stub.PutState(partnerRangeStartKey, []byte("partner:101"))
    stub.PutState(partnerRangeEndKey, []byte("partner:101"))
	stub.PutState(couponRangeStartKey, []byte("coupon:101"))
    stub.PutState(couponRangeEndKey, []byte("coupon:0"))
    stub.PutState(salesTransactionRangeStartKey,[]byte ("salestransaction:101"))
	stub.PutState(salesTransactionRangeEndKey, []byte ("salestransaction:0"))
}

