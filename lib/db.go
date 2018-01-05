package lib

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/boltdb/bolt"
)

var db *bolt.DB
var dbOpen bool

// Database ...
var Database = "httplive.db"

// DatabasePath ...
var DatabasePath string

// OpenDb ...
func OpenDb() error {
	var err error
	config := &bolt.Options{Timeout: 1 * time.Second}
	db, err = bolt.Open(Environments.DbFile, 0600, config)
	if err != nil {
		log.Fatal(err)
	}
	dbOpen = true
	return nil
}

// CloseDb ...
func CloseDb() {
	dbOpen = false
	db.Close()
}

// CreateDbBucket ...
func CreateDbBucket() error {
	if !dbOpen {
		return fmt.Errorf("open db connection first")
	}
	err := db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(Environments.Port))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})
	return err
}

// InitDbValues ...
func InitDbValues() {
	apis := []APIDataModel{
		{Endpoint: "/api/token/mobiletoken", Method: "GET", Body: `{
	"array": [
		1,
		2,
		3
	],
	"boolean": true,
	"null": null,
	"number": 123,
	"object": {
		"a": "b",
		"c": "d",
		"e": "f"
	},
	"string": "Hello World"
}`}}

	for _, api := range apis {
		key := CreateEndpointKey(api.Method, api.Endpoint)
		model, _ := GetEndpoint(key)
		if model == nil {
			SaveEndpoint(&api)
		}
	}
}

// Encode ...
func (model *APIDataModel) Encode() ([]byte, error) {
	enc, err := json.Marshal(model)
	if err != nil {
		return nil, err
	}
	return enc, nil
}

// Decode ...
func Decode(data []byte) (*APIDataModel, error) {
	var model *APIDataModel
	err := json.Unmarshal(data, &model)
	if err != nil {
		return nil, err
	}
	return model, nil
}

// GobEncode ...
func (model *APIDataModel) GobEncode() ([]byte, error) {
	buf := new(bytes.Buffer)
	enc := gob.NewEncoder(buf)
	err := enc.Encode(model)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// GobDecode ...
func GobDecode(data []byte) (*APIDataModel, error) {
	var model *APIDataModel
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	err := dec.Decode(&model)
	if err != nil {
		return nil, err
	}
	return model, nil
}

// SaveEndpoint ...
func SaveEndpoint(model *APIDataModel) error {
	if !dbOpen {
		return fmt.Errorf("open db connection first")
	}

	if model.Endpoint == "" || model.Method == "" {
		return fmt.Errorf("model endpoint and method could not be empty")
	}

	key := CreateEndpointKey(model.Method, model.Endpoint)
	err := db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(Environments.Port))
		if model.ID <= 0 {
			id, _ := bucket.NextSequence()
			model.ID = int(id)
		}
		enc, err := model.Encode()
		if err != nil {
			return fmt.Errorf("could not encode APIDataModel %s: %s", key, err)
		}
		err = bucket.Put([]byte(key), enc)
		return err
	})
	return err
}

// DeleteEndpoint ...
func DeleteEndpoint(endpointKey string) error {
	if !dbOpen {
		return fmt.Errorf("db must be opened before saving")
	}

	if endpointKey == "" {
		return fmt.Errorf("endpointKey")
	}

	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(Environments.Port))
		k := []byte(endpointKey)
		return b.Delete(k)
	})

	return err
}

// GetEndpoint ...
func GetEndpoint(endpointKey string) (*APIDataModel, error) {
	if !dbOpen {
		return nil, fmt.Errorf("db must be opened before saving")
	}

	if endpointKey == "" {
		return nil, fmt.Errorf("endpointKey")
	}

	var model *APIDataModel
	err := db.View(func(tx *bolt.Tx) error {
		var err error
		b := tx.Bucket([]byte(Environments.Port))
		k := []byte(endpointKey)
		model, err = Decode(b.Get(k))
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		fmt.Printf("Could not get APIDataModel with key: %s", endpointKey)
		return nil, err
	}
	return model, nil
}

// OrderByID ...
func OrderByID(items map[string]APIDataModel) PairList {
	pl := make(PairList, len(items))
	i := 0
	for k, v := range items {
		pl[i] = Pair{k, v}
		i++
	}
	sort.Sort(sort.Reverse(pl))
	return pl
}

// EndpointList ...
func EndpointList() []APIDataModel {
	data := make(map[string]APIDataModel)
	db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket([]byte(Environments.Port)).Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			key := string(k)
			model, err := Decode(v)
			if err == nil {
				data[key] = *model
			}
		}
		return nil
	})

	pairList := OrderByID(data)
	items := []APIDataModel{}
	for _, v := range pairList {
		items = append(items, v.Value)
	}

	return items
}
