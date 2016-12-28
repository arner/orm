package orm

import (
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"testing"
	"fmt"
)

// Need a chaincode to start stub
type MockChaincode struct{}
func (t *MockChaincode) Init(stub shim.ChaincodeStubInterface) ([]byte, error) {return nil,nil}
func (t *MockChaincode) Invoke(stub shim.ChaincodeStubInterface) ([]byte, error) {return nil, nil}
func (t *MockChaincode) Query(stub shim.ChaincodeStubInterface) ([]byte, error) {return nil, nil}

const STRUCT_NAME = "TestStruct"

type TestStruct struct {
	privateField string
	Str string
	I64 int64
	I32 int32
	UI32 uint32
	UI64 uint64
//	Bytes []uint8
	Bool bool
	Saveable
}

func getTestStruct() TestStruct {
	return TestStruct{
		Str: "isAString",
		I64: -9999999999,
		I32: -9999999,
		UI32: 99999999,
		UI64: 999999999,
	//	Bytes: []uint8("isAByteArray"),
		Bool: true,
	}
}

func checkCreateTable(t *testing.T, stub shim.ChaincodeStubInterface) {
	s := TestStruct{}
	if err := CreateTable(stub, &s); err != nil {
		fmt.Println(err)
		t.FailNow()
	}
	if _, err := stub.GetTable(STRUCT_NAME); err != nil {
		fmt.Println(err)
		t.FailNow()
	}
}

func TestCreateTable(t *testing.T) {
	stub := shim.NewMockStub("cc", new(MockChaincode))
	stub.MockTransactionStart("test")
	checkCreateTable(t, stub)
}

func checkCreate(t *testing.T, stub shim.ChaincodeStubInterface) {
	s := getTestStruct()
	if err := Create(stub, &s); err != nil {
		fail(t, err)
	} else {
		if s.Id == 0 {
			fmt.Println("Id is not set");
			t.FailNow()
		}
	}
}

func TestInsert(t *testing.T) {
	stub := shim.NewMockStub("cc", new(MockChaincode))
	stub.MockTransactionStart("test")
	checkCreateTable(t, stub)
	checkCreate(t, stub)
}

func checkGet(t *testing.T, stub shim.ChaincodeStubInterface) TestStruct {
	var s TestStruct
	if err := Get(stub, &s, 1); err != nil {
		fail(t, err)
	}
	return s
}

func checkEqual(t *testing.T, a TestStruct, b TestStruct) {
	if a.I32 != b.I32 {
		fail(t, "i32 not ok")
	}
	if a.I64 != b.I64 {
		fail(t, "i64 not ok")
	}
	if a.Str != b.Str {
		fail(t, "str not ok")
	}
	if a.UI32 != b.UI32 {
		fail(t, "ui32 not ok")
	}
	if a.UI64 != b.UI64 {
		fail(t, "ui64 not ok")
	}
	//if string(a.Bytes) != string(b.Bytes) {
	//	fail(t, "bytes not ok")
	//}
	if a.Bool != b.Bool {
		fail(t, "bool not ok")
	}
}

func TestGet(t *testing.T) {
	stub := shim.NewMockStub("cc", new(MockChaincode))
	stub.MockTransactionStart("test")
	checkCreateTable(t, stub)
	checkCreate(t, stub)
	a := checkGet(t, stub)
	b := getTestStruct()
	checkEqual(t, a, b)
}

func TestGetShouldFail(t *testing.T) {
	stub := shim.NewMockStub("cc", new(MockChaincode))
	stub.MockTransactionStart("test")
	checkCreateTable(t, stub)
	var s TestStruct
	if err := Get(stub, &s, 10000); err == nil {
		fail(t, "Get should fail with non existing Id")
	}
}


func checkGetAll(t *testing.T, stub shim.ChaincodeStubInterface) []TestStruct {
	var items []TestStruct
	if err := GetAll(stub, &items); err != nil {
		fail(t, err)
	}
	return items
}

func TestUpdate(t *testing.T) {
	stub := shim.NewMockStub("cc", new(MockChaincode))
	stub.MockTransactionStart("test")
	checkCreateTable(t, stub)
	checkCreate(t, stub)
	a := checkGet(t, stub)
	b := getTestStruct()
	checkEqual(t, a, b)

	a.Str = "Updated"
	a.I64 = -9
	a.I32 = -9
	a.UI32 = 9
	a.UI64 = 9
	a.Bool = true

	Update(stub, &a)

	b.Str = "Updated"
	b.I64 = -9
	b.I32 = -9
	b.UI32 = 9
	b.UI64 = 9
	b.Bool = true

	var c TestStruct
	Get(stub, &c, a.Id)

	checkEqual(t, b, c)
}


func TestDelete(t *testing.T) {
	stub := shim.NewMockStub("cc", new(MockChaincode))
	stub.MockTransactionStart("test")
	checkCreateTable(t, stub)
	checkCreate(t, stub)
	a := checkGet(t, stub)
	if err := Delete(stub, &a); err != nil {
		fail(t, err)
	}
	var s TestStruct
	if err := Get(stub, &s, 1); err == nil {
		fail(t, "After delete, Get should return error")
	}
}


//Mock not working correctly!
//func TestGetAll(t *testing.T) {
//	stub := shim.NewMockStub("cc", new(MockChaincode))
//	stub.MockTransactionStart("test")
//	checkCreateTable(t, stub)
//	checkInsert(t, stub)
//	checkInsert(t, stub)
//	items := checkGetAll(t, stub)
//	if len(items) != 2 {
//		fmt.Println(len(items))
//		fail(t, "Not the right amount of items returned (expected 2):" + string(len(items)))
//	}
//}



func fail(t *testing.T, arg interface{}) {
	fmt.Println(arg)
	t.FailNow()
}




