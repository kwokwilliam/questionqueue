package main

import (
	"encoding/json"
	"log"
	"questionqueue/src/db"
)

func main() {

	ms, _ := db.NewMongoStore("mongodb://localhost:27017")

	allClass, _ := ms.GetAllClass()

	j, _ := json.Marshal(allClass)
	log.Println(string(j))
}

