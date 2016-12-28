[![Build Status](https://travis-ci.org/arner/orm.svg?branch=master)](https://travis-ci.org/arner/orm)

# orm
Simple wrapper for Hyperledger Fabric 0.6 tables.  
  
Use at your own risk, it's an early version and the interface may change.  
 
## Usage
```golang
    type User struct {
        FirstName string
        orm.Saveable // This adds an Id field and makes sure we can save.
    }
    
    // Create table
    user := new(User)  
 	if err := orm.CreateTable(stub, user); err != nil {  
 		return nil, errors.Wrap(err, "Failed creating table.")  
 	}  
  
    // Create User
 	arne := User{FirstName: "Arne"}  
 	if err := orm.Create(stub, &arne); err != nil {  
 		return nil, errors.Wrap(err, "Failed creating user.")  
 	}  
 	
 	// Get User with Id 1
    var user User  
    if err := orm.Get(stub, &user, 1); err != nil {  
        return nil, err  
    }  
    
    // Update User
    user.FirstName = "Dave"  
    if err := orm.Update(stub, &user); err != nil {  
        return nil, err  
    }  
    
    // Delete User
    if err := orm.Delete(stub, &user); err != nil {    
        return nil, err  
    }  
 ```
