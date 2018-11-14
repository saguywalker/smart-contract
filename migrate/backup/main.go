package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	did "github.com/ndidplatform/smart-contract/abci/did/v1"
	"github.com/tendermint/iavl"
	dbm "github.com/tendermint/tendermint/libs/db"
)

var (
	kvPairPrefixKey = []byte("kvPairKey:")
)

func main() {
	// Variable
	dbFile := "DB1"
	dbName := "didDB"
	backupDBFile := "Backup_DB"
	backupDataFileName := "data"
	backupValidatorFileName := "validators"

	// Delete backup file
	deleteFile("migrate/data/" + backupDataFileName + ".txt")
	deleteFile("migrate/data/" + backupValidatorFileName + ".txt")
	os.Remove("Backup_DB")

	// Copy stateDB dir
	copyDir(dbFile, backupDBFile)

	// Save kv from backup DB
	db := dbm.NewDB(dbName, "leveldb", backupDBFile)
	oldTree := iavl.NewMutableTree(db, 0)
	oldTree.Load()
	tree, _ := oldTree.GetImmutable(oldTree.Version())
	_, ndidNodeID := tree.Get(prefixKey([]byte("MasterNDID")))
	tree.Iterate(func(key []byte, value []byte) (stop bool) {
		// Validator
		if strings.Contains(string(key), "val:") {
			var kv did.KeyValue
			kv.Key = key
			kv.Value = value
			jsonStr, err := json.Marshal(kv)
			if err != nil {
				panic(err)
			}
			fWriteLn(backupValidatorFileName, jsonStr)
			return false
		}

		if strings.Contains(string(key), string(ndidNodeID)) {
			return false
		}
		if strings.Contains(string(key), "MasterNDID") {
			return false
		}
		if strings.Contains(string(key), "InitState") {
			return false
		}
		var kv did.KeyValue
		kv.Key = key
		kv.Value = value
		jsonStr, err := json.Marshal(kv)
		if err != nil {
			panic(err)
		}
		fWriteLn(backupDataFileName, jsonStr)
		return false
	})
}

func copyDir(source string, dest string) (err error) {
	sourceinfo, err := os.Stat(source)
	if err != nil {
		return err
	}
	err = os.MkdirAll(dest, sourceinfo.Mode())
	if err != nil {
		return err
	}
	directory, _ := os.Open(source)
	objects, err := directory.Readdir(-1)
	for _, obj := range objects {
		sourcefilepointer := source + "/" + obj.Name()
		destinationfilepointer := dest + "/" + obj.Name()
		if obj.IsDir() {
			err = copyDir(sourcefilepointer, destinationfilepointer)
			if err != nil {
				fmt.Println(err)
			}
		} else {
			err = copyFile(sourcefilepointer, destinationfilepointer)
			if err != nil {
				fmt.Println(err)
			}
		}
	}
	return
}

func copyFile(source string, dest string) (err error) {
	sourcefile, err := os.Open(source)
	if err != nil {
		return err
	}
	defer sourcefile.Close()
	destfile, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer destfile.Close()
	_, err = io.Copy(destfile, sourcefile)
	if err == nil {
		sourceinfo, err := os.Stat(source)
		if err != nil {
			err = os.Chmod(dest, sourceinfo.Mode())
		}

	}
	return
}

func prefixKey(key []byte) []byte {
	return append(kvPairPrefixKey, key...)
}

func fWriteLn(filename string, data []byte) {
	createDirIfNotExist("migrate/data")
	f, err := os.OpenFile("migrate/data/"+filename+".txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	_, err = f.Write(data)
	if err != nil {
		panic(err)
	}
	_, err = f.WriteString("\r\n")
	if err != nil {
		panic(err)
	}
}

func createDirIfNotExist(dir string) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			panic(err)
		}
	}
}

func deleteFile(dir string) {
	_, err := os.Stat(dir)
	if err != nil {
		return
	}
	err = os.Remove(dir)
	if err != nil {
		panic(err)
	}
}
