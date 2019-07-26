package main

import (
	"time"
	"encoding/json"
	"fmt"
	"strconv"
    //"strings"
	"bytes"
	"github.com/pborman/uuid"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	 sc "github.com/hyperledger/fabric/protos/peer"
)

func (t *Coupon) Init(stub shim.ChaincodeStubInterface) sc.Response {
	return shim.Success(nil)
}

type Coupon struct {
	Id       			string     	  		`json:"id"`
	Name     			string  			 `json:"name"`
	CreatedDateTime     time.Time 			 `json:"created_date"`
	ExpiresOn    		time.Time   		 `json:"expiry_date"`	
	DiscountPercent 	float64    			 `json:"discount_amount_percent"`
	RevenueSharePercent float64    			`json:"revenue_share_percent"`
	Status        		string               `json:"is_valid"` 
	CustomerObj   		Customer 		     `json:"customer"`
}

type Customer struct {
	CustomerId  	    string   			 `json:"customer_id"`
	Name 				string    			 `json:"name"`
	EmailID		 		string     			 `json:"email_id"`
}

type Redeem struct{
	OriginalPrice       float64 	     `json:"original_price"`
	SalesAmount			float64 		 `json:"sales_amount"`
	CouponObj			Coupon			 `json:"coupon"`
	RevenueShareAmount  float64          `json:"revenue_share_amount"`
	SettlementAmount	float64			 `json:"settlement_amount"`
	//TransactionTime		float64 		  `json:"transaction_time"`  ??
}

type CustomerCoupon struct{

	Customer 
	coupons 	[]Coupon
}

type SalesTransaction struct {
	PartnerId string  					`json:"partner_id"`
	PartnerName string 					`json:"partner_name"`
	// City string  						`json:"city"`
	// State string 						`json:"state"`
	// Country string 						`json:"country"`
	// Zipcode string 						`json:"zipcode"`
	// RetailCommissionPercentage float64 	`json:"retail_commission_percentage"`
}
type CouponChaincode struct {
}

// Init is called during Instantiate transaction.
func (t *CouponChaincode) Init(APIstub shim.ChaincodeStubInterface) sc.Response {
	return shim.Success(nil)
}

// Invoke is called to update or query the ledger in a proposal transaction.
func (c *CouponChaincode) Invoke(APIstub shim.ChaincodeStubInterface) sc.Response {

	// Retrieve the requested Chaincode function and arguments
	function, args := APIstub.GetFunctionAndParameters()
	// Route to the appropriate handler function to interact with the ledger appropriately
	if function == "createCoupon" {
		return c.createCoupon(APIstub, args)
	} else if function == "initLedger" {
		return c.initLedger(APIstub)
	}	else if function == "getCouponById" {
		return c.getCouponById(APIstub,args)
	} else if function == "getAllCoupons" {
		return c.getAllCoupons(APIstub)
		// else if function == "getAllCoupons" {
		// 	return c.getAllCoupons(APIstub,args)
	}  
	// else if function == "redeemCoupon" {
	// 	return c.redeemCoupon(APIstub, args)
	// } else if function == "issueCoupon" {
	// 	return c.issueCoupon(APIstub, args)
	// }
	// } else if function == "validateCoupon" {
	// 	return c.validateCoupon(APIstub, args)
	// }
	return shim.Error("No such chain code function available.")
}

// Function to create a coupon in the  ledger during a proposal transaction.
func (c *CouponChaincode) createCoupon(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
 
	if len(args) != 6 {
		return shim.Error("Incorrect number of arguments. Expecting 7 arguments")
	}
// create unique coupon id 
couponId := uuid.New()
customerId := uuid.New()
//timeLayout  := "Mon, 01/02/06, 03:04PM"

dateFormat := "02-01-2006"
parsedCreatedDate, err := time.Parse(dateFormat,args[1])
if err!=nil{
	return shim.Error("Created date not in correct format. Unable to parse! Enter date in format dd-mm-yy")
}
parsedExpiryDate,err := time.Parse(dateFormat,args[2])
if err!=nil{
	return shim.Error("ExpiresOn date not  in correct format. Unable to parse! Enter date in format dd-mm-yy")
}

parsedDiscountAmountPercentage, err := strconv.ParseFloat(args[3], 64)
if err!=nil{
	return shim.Error("Discount Amount Percentage is invalid. Unable to parse !")
}
	
	var myCoupon = Coupon{Id:couponId, Name: args[0], CreatedDateTime: parsedCreatedDate, ExpiresOn: parsedExpiryDate, DiscountPercent: parsedDiscountAmountPercentage,
		Status: "Valid", CustomerObj: Customer{ Name: args[4], EmailID :args[5],CustomerId:customerId} }
	
	couponAsBytes, _ := json.Marshal(myCoupon)
	APIstub.PutState(myCoupon.Id, couponAsBytes)

	return shim.Success(nil)
}

// Function to initialise the ledger with default coupon data.
func (c *CouponChaincode) initLedger(APIstub shim.ChaincodeStubInterface) sc.Response {
	//dummy customer and coupon
	
	couponID1 := uuid.New() // creates a new UUID 
	couponID2 := uuid.New()
	couponID3 := uuid.New()
	customerID1 := uuid.New()
	customerID2 := uuid.New()
	customerID3 := uuid.New()
	customerID4 := uuid.New()
	customerID5 := uuid.New()
	customerID6 := uuid.New()

	coupons := []Coupon{
		Coupon{Id: couponID1, Name: "Big Billion Day", CreatedDateTime:time.Date(
			2019, 07, 18, 20, 34, 58, 0, time.UTC), ExpiresOn: time.Date(
				2019, 12, 22, 20, 34, 58, 0, time.UTC),  DiscountPercent:7.5,Status:"Valid", CustomerObj:Customer{ Name:"Arun", EmailID:"arun@gmail.com", CustomerId:customerID1} },
		Coupon{Id: couponID2, Name: "Festival Offer", CreatedDateTime:time.Date(
			2019, 8, 18, 20, 34, 58, 0, time.UTC), ExpiresOn: time.Date(
			2019, 12, 22, 20, 34, 58, 0, time.UTC), DiscountPercent:10.5,Status:"Valid", CustomerObj:Customer{ Name:"Asha", EmailID:"asha@gmail.com", CustomerId:customerID2} },
		Coupon{Id: couponID3, Name: "Black Friday Sale", CreatedDateTime:time.Date(
			2019, 9, 18, 20, 34, 58, 0, time.UTC), ExpiresOn: time.Date(
			2019, 12, 22, 20, 34, 58, 0, time.UTC), DiscountPercent:20.5,Status:"Valid", CustomerObj:Customer{ Name:"Varun", EmailID:"varun@gmail.com", CustomerId:customerID3} },
			Coupon{Id: couponID3, Name: "Sale1", CreatedDateTime:time.Date(
				2019, 9, 18, 20, 34, 58, 0, time.UTC), ExpiresOn: time.Date(
				2019, 12, 22, 20, 34, 58, 0, time.UTC), DiscountPercent:20.5,Status:"Valid", CustomerObj:Customer{ Name:"Ramu", EmailID:"ramu@gmail.com", CustomerId:customerID4} },
				Coupon{Id: couponID3, Name: "Sale2", CreatedDateTime:time.Date(
					2019, 9, 18, 20, 34, 58, 0, time.UTC), ExpiresOn: time.Date(
					2019, 12, 22, 20, 34, 58, 0, time.UTC), DiscountPercent:20.5,Status:"Valid", CustomerObj:Customer{ Name:"Shyam", EmailID:"shyam@gmail.com", CustomerId:customerID5} },
					Coupon{Id: couponID3, Name: "Sale3", CreatedDateTime:time.Date(
						2019, 9, 18, 20, 34, 58, 0, time.UTC), ExpiresOn: time.Date(
						2019, 12, 22, 20, 34, 58, 0, time.UTC), DiscountPercent:20.5,Status:"Valid", CustomerObj:Customer{ Name:"Asha", EmailID:"asha@gmail.com", CustomerId:customerID6} },
	}

	
	couponCount := 0
	for couponCount < len(coupons) {
		couponAsBytes, _ := json.Marshal(coupons[couponCount])
		APIstub.PutState(coupons[couponCount].Id, couponAsBytes)
		couponCount = couponCount + 1
	}
	return shim.Success(nil)
}

//Function to get all the Coupons
func (c *CouponChaincode) getAllCoupons(APIstub shim.ChaincodeStubInterface) sc.Response {
  
	// if len(args) != 1 {
	// 	return shim.Error("Incorrect number of arguments. Expecting 1")
	// }

	// queryString :="Select * from "
	// resultsIterator, err := APIstub.GetQueryResult(queryString)
	//  if err != nil {
	// 	 return shim.Error(err.Error())
	//  }
	//  defer resultsIterator.Close()

	// return shim.Success(nil)
	// startKey := "Coupon0"
	// endKey := "Coupon100"

	// resultsIterator, err := APIstub.GetStateByRange(startKey, endKey)
	// if err != nil {
	// 	return shim.Error(err.Error())
	// }
	// defer resultsIterator.Close()

	// // buffer is a JSON array containing QueryResults
	 var buffer bytes.Buffer
	// buffer.WriteString("[")

	// bArrayMemberAlreadyWritten := false
	// for resultsIterator.HasNext() {
	// 	queryResponse, err := resultsIterator.Next()
	// 	if err != nil {
	// 		return shim.Error(err.Error())
	// 	}
	// 	if bArrayMemberAlreadyWritten == true {
	// 		buffer.WriteString(",")
	// 	}
	// 	buffer.WriteString("{\"Key\":")
	// 	buffer.WriteString("\"")
	// 	buffer.WriteString(queryResponse.Key)
	// 	buffer.WriteString("\"")

	// 	buffer.WriteString(", \"Record\":")
	// 	buffer.WriteString(string(queryResponse.Value))
	// 	buffer.WriteString("}")
	// 	bArrayMemberAlreadyWritten = true
	// }
	// buffer.WriteString("]")
	 return shim.Success(buffer.Bytes())
}

// Function to get the Coupon information based on Coupon Data 
 func (c *CouponChaincode) getCouponById(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
 
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}
	couponAsBytes, err := APIstub.GetState(args[0])
	if(err !=nil){
		return shim.Error(err.Error())
	}else{
		return shim.Success(couponAsBytes)
	}
}

//Main Function
func main (){
	 err := shim.Start(new(CouponChaincode))
	 if err != nil {
		 fmt.Printf("Error creating new Smart Contract: %s", err)
	 }
}
