// Package orm provides a simple wrapper for Hyperledger Fabric tables.
//
// Add an anonymous field orm.Saveable to your entities. At initialization, create a table as follows:
//
// user := new(User)
// if err := orm.CreateTable(stub, user); err != nil {
//   return nil, errors.Wrap(err, "Failed creating table.")
// }
//
// For more info, see the README.md on github
package orm

import (
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/pkg/errors"
	"reflect"
)

// Items need to implement this interface to use ORM. You can use an anonymous Saveable in your struct.
type BlockchainItemizer interface {
	GetId() int64
	SetId(int64)
}

// Place an anonymous Saveable in your struct to use ORM.
type Saveable struct {
	Id int64 `json:"id" key:"true"`
}
func (s *Saveable) GetId() int64   { return s.Id }
func (s *Saveable) SetId(id int64) { s.Id = id }

//
var columnDefinitions = map[string]shim.ColumnDefinition_Type {
	"bool": shim.ColumnDefinition_BOOL,
	//"[]uint8": shim.ColumnDefinition_BYTES, // TODO
	"int32": shim.ColumnDefinition_INT32,
	"int64": shim.ColumnDefinition_INT64,
	"string": shim.ColumnDefinition_STRING,
	"uint32": shim.ColumnDefinition_UINT32,
	"uint64": shim.ColumnDefinition_UINT64,
	"Saveable": shim.ColumnDefinition_INT64, // Id field (TODO: recursively find subfields of anonymous fields)
}

var logger = shim.NewLogger("orm")

// Create a table of the passed item. Types are automatically inferred.
func CreateTable(stub shim.ChaincodeStubInterface, item BlockchainItemizer) error {
	name := reflect.TypeOf(item).Elem().Name()
	logger.Infof("Create Table %s", name)

	cds := createColumnDefinitions(item)
	logger.Debugf("Columns: %v", cds)
	return stub.CreateTable(name, cds)
}

// Get an item by Id
func Get(stub shim.ChaincodeStubInterface, item BlockchainItemizer, id int64) error {
	if (id == 0) {
		return errors.New("Id should be larger than 0")
	}

	// Query
	var columns []shim.Column
	col1 := shim.Column{Value: &shim.Column_Int64{Int64: id}} // How to set key?
	columns = append(columns, col1)

	// Table / Item name
	name := reflect.TypeOf(item).Elem().Name()

	// Get table
	if tbl, err := stub.GetTable(name); err != nil {
		return errors.Wrap(err, "Could not get table "+name)

	// Get row based on query
	} else if row, err := stub.GetRow(name, columns); err != nil {
		return errors.Wrap(err, "Could not get "+name+" with id "+string(id))

	// Set values of item based on row values
	} else if err = setValues(tbl, row, item); err != nil {
		return errors.Wrap(err, "Error setting values")
	}

	if (item.GetId() == 0) {
		return errors.New("Item not found.")
	}

	logger.Debugf("Got item %v", item)
	return nil
}

// Get all items by passing a slice of the correct type
func GetAll(stub shim.ChaincodeStubInterface, items interface{}) error {
	v := reflect.ValueOf(items).Elem()
	if v.Kind() != reflect.Slice {
		return errors.New("Object passed to GetAll should be a slice.")
	}

	t := reflect.TypeOf(items).Elem().Elem()
	name := t.Name();

	//logger.Debugf("Getting all %vs", name)

	// Query (TODO)
	columns := []shim.Column{
	//	shim.Column{Value: &shim.Column_Int64{Int64: 1}},
	//	shim.Column{Value: &shim.Column_Int64{Int64: 2}},
	}

	tbl, err := stub.GetTable(name)
	if err != nil {
		return errors.Wrap(err, "Could not get table "+name)
	}

	rowChannel, err := stub.GetRows(name, columns)
	if err != nil {
		return fmt.Errorf("getRows operation failed. %s", err)
	}
	for {
		select {
		case row, ok := <-rowChannel:
			if !ok {
				rowChannel = nil
			} else {
				logger.Debugf("Columns: %v", row.Columns)
				item := reflect.New(t).Interface()

				if err:= setValues(tbl, row, item); err != nil {
					return errors.Wrap(err, "Error setting values.")
				}

				logger.Debugf("Adding item: %v", item)
				v.Set(reflect.Append(v, reflect.ValueOf(item).Elem()))
			}
		}
		if rowChannel == nil {
			break
		}
	}
	return nil
}

// Insert a row for the item in the database
func Create(stub shim.ChaincodeStubInterface, item BlockchainItemizer) error {
	t := reflect.TypeOf(item).Elem()
	v := reflect.ValueOf(item).Elem()
	logger.Infof("Creating %v: %v", t.Name(), v)

	if id, err := generateId(stub, t.Name()); err != nil {
		return errors.Wrap(err, "Generate id failed.")
	} else {
		item.SetId(id)
	}

	if row, err := createRow(t, v); err != nil {
		return err
	} else {
		_, err := stub.InsertRow(t.Name(), row)
		return err
	}
}

// Update an item
func Update(stub shim.ChaincodeStubInterface, item BlockchainItemizer) error {
	t := reflect.TypeOf(item).Elem()
	v := reflect.ValueOf(item).Elem()
	logger.Infof("Updating %v: %v", t.Name(), v)

	if item.GetId() == 0 {
		return errors.New("Item cannot have id 0")
	}

	if row, err := createRow(t, v); err != nil {
		return err
	} else {
		_, err := stub.ReplaceRow(t.Name(), row)
		return err
	}

}

// Delete an item
func Delete(stub shim.ChaincodeStubInterface, item BlockchainItemizer) error {
	t := reflect.TypeOf(item).Elem()
	v := reflect.ValueOf(item).Elem()
	logger.Infof("Deleting %v: %v", t.Name(), v)

	if item.GetId() == 0 {
		return errors.New("Item cannot have id 0")
	}

	columns := []shim.Column {
		shim.Column{Value: &shim.Column_Int64{Int64: item.GetId()}},
	}

	return stub.DeleteRow(t.Name(), columns)
}


// Set the values of a retrieved row to an item
func setValues(tbl *shim.Table, row shim.Row, item interface{}) error {
	v := reflect.Indirect(reflect.ValueOf(item))
	if !v.IsValid() {
		return errors.New("Zero value passed to setValues")
	} else if !v.CanSet() {
		return errors.New("Cannot set item")
	}

	// Get the column names and set the value based on the row values
	for i, c := range row.GetColumns() {
		name := tbl.ColumnDefinitions[i].Name
		fieldType := tbl.ColumnDefinitions[i].Type //ColumnDefinition_Type
		logger.Debugf("[%v] %v = %v", fieldType, name, c.GetValue())
		f := v.FieldByName(name)

		switch fieldType {
		case shim.ColumnDefinition_BOOL:
			f.SetBool(c.GetBool())
			break
		case shim.ColumnDefinition_BYTES:
			f.SetBytes(c.GetBytes())
			break
		case shim.ColumnDefinition_INT32:
			f.SetInt(int64(c.GetInt32()))	 // ???
			break
		case shim.ColumnDefinition_INT64:
			f.SetInt(c.GetInt64())
			break
		case shim.ColumnDefinition_STRING:
			f.SetString(c.GetString_())
			break
		case shim.ColumnDefinition_UINT32:
			f.SetUint(uint64(c.GetUint32())) // ???
			break
		case shim.ColumnDefinition_UINT64:
			f.SetUint(c.GetUint64())
			break
		default:
			return errors.New("Type " + fieldType.String() + " not recognized.")
		}
	}
	return nil
}

// Create a row
func createRow(t reflect.Type, v reflect.Value) (shim.Row, error) {
	row := shim.Row{}
	for i := 0; i < t.NumField(); i++ {
		f := v.Field(i);
		if !f.CanSet() {
			continue // Field not exported?
		}
		if column, err := createColumnValue(t.Field(i), f.Interface()); err != nil {
			return row, errors.Wrap(err, "Create item failed - Can't create column value")
		} else {
			row.Columns = append(row.Columns, &column)
		}
	}
	return row, nil
}


// Create definitions for the table that will be created.
func createColumnDefinitions(iface interface{}) []*shim.ColumnDefinition {
	defs := make([]*shim.ColumnDefinition, 0)
	t := reflect.TypeOf(iface).Elem()
	v := reflect.ValueOf(iface).Elem()

	for i := 0; i < t.NumField(); i++ {
		if !v.Field(i).CanSet() {
			continue // Field not exported?
		}
		f := t.Field(i)
		logger.Debugf("field: %v", f)

		isKey := f.Tag.Get("key") == "true"
		name := f.Name

		if typ, ok := columnDefinitions[t.Field(i).Type.Name()]; ok {
			// FIXME: this is a hack. Should be solved recursively
			if t.Field(i).Type.Name() == "Saveable" {
				name = "Id"
				isKey = true
			}
			defs = append(defs, &shim.ColumnDefinition{Name: name, Type: typ, Key: isKey})
		} else {
			logger.Errorf("Field type not recognized: %v %v", f.Name, t.Field(i).Type.Name())
		}
	}

	return defs
}

// Set the value of a field
func createColumnValue(field reflect.StructField, val interface{}) (shim.Column, error) {
	switch field.Type.Name() {
	case "bool":
		return shim.Column{Value: &shim.Column_Bool{Bool: val.(bool)}}, nil
	// case "[]uint8":
	// 	return shim.Column{Value: &shim.Column_Bytes{Bytes: val.([]uint8)}}, nil // TODO
	case "int32":
		return shim.Column{Value: &shim.Column_Int32{Int32: val.(int32)}}, nil
	case "int64":
		return shim.Column{Value: &shim.Column_Int64{Int64: val.(int64)}}, nil
	case "string":
		return shim.Column{Value: &shim.Column_String_{String_: val.(string)}}, nil
	case "uint32":
		return shim.Column{Value: &shim.Column_Uint32{Uint32: val.(uint32)}}, nil
	case "uint64":
		return shim.Column{Value: &shim.Column_Uint64{Uint64: val.(uint64)}}, nil
	case "Saveable":
		return shim.Column{Value: &shim.Column_Int64{Int64: val.(Saveable).Id}}, nil //FIXME

	}
	return shim.Column{}, errors.New("Type of " + field.Type.Name() + " not recognized.")
}

// Generates an id that's one higher than the latest update.
// FIXME: race condition when creating multiple items in one call
func generateId(stub shim.ChaincodeStubInterface, tableName string) (int64, error) {
	rowChannel, err := stub.GetRows(tableName, []shim.Column{})
	if err != nil {
		return 0, fmt.Errorf("getRows operation failed. %s", err)
	}
	id := int64(0)
	for {
		select {
		case row, ok := <-rowChannel:
			if !ok {
				rowChannel = nil
			} else {
				logger.Debugf("Columns: %v", row.Columns)
				if val := row.Columns[0].GetInt64(); val > id {
					id = val
				}
			}
		}
		if rowChannel == nil {
			break
		}
	}
	id++
	logger.Debugf("Generated id %d for %s", id, tableName)
	return id, nil
}