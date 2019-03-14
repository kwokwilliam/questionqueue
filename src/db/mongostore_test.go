package db

import (
	"encoding/json"
	"log"
)

// connector
const mongoAddr = "mongodb://localhost:27017"
var ms, _ = NewMongoStore(mongoAddr)

func ExampleMongoStore_GetActiveQuestions() {
	res, err := ms.GetActiveQuestions()
	if err != nil {
		log.Printf("got error: %v", err)
	} else {
		js, _ := json.Marshal(res)
		log.Println(string(js))
	}
}

//func ExampleMongoStore_SolveQuestion() {
//	const name = "someone"
//	res, err := ms.SolveQuestion(name)
//	if err != nil {
//		log.Printf("cannot update question belongs to %v: %v", name, err)
//	} else {
//		log.Println(res.ModifiedCount)
//	}
//}