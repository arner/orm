# orm
Simple wrapper for Hyperledger Fabric 0.6 tables.  
  
Use at your own risk, it's an early version and the interface may change.  
 
## Usage
```golang    
    user := new(User)  
 	if err := orm.CreateTable(stub, user); err != nil {  
 		return nil, errors.Wrap(err, "Failed creating table.")  
 	}  
  
 	arne := User{FirstName: "Arne"}  
 	if err := orm.Create(stub, &arne); err != nil {  
 		return nil, errors.Wrap(err, "Failed creating user.")  
 	}  
 	
    var user User  
    if err := orm.Get(stub, &user, 1); err != nil {  
        return nil, err  
    }  
    
    user.FirstName = "Dave"  
    if err := orm.Update(stub, &user); err != nil {  
        return nil, err  
    }  
    
    if err := orm.Delete(stub, &user); err != nil {    
        return nil, err  
    }  
 ```
