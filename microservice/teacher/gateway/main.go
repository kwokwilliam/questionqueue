package main

import (
	"log"
	"questionqueue/servers/db"
	"questionqueue/servers/model"
)

func main() {

	ms, _ := db.NewMongoStore("mongodb://localhost:27017")

	//if res, err := collectionTest.InsertOne(nil, bson.M{"title":"a thing"});
	//err != nil {
	//	log.Fatalf("cannot insert one: %v", err)
	//} else {
	//	log.Printf("inserted: %v", res.InsertedID)
	//}

	//if res, err := db.Find(test, map[string]interface{}{"title" : "a thing"});
	//err != nil {
	//	log.Fatalf("cannot find: %v", err)
	//} else {
	//	for res.Next(nil) {
	//		curr := res.Current
	//		log.Println(curr.String())
	//	}
	//}
	cls := model.Class{
		Code: "1234",
		Type: []string{"1","222"},
	}

	//_, _ = ms.InsertClass(&cls)

	newCls := model.Class{
		Code: "1234",
		Type: []string{"000"},
	}

	res, err := ms.UpdateClass(&cls, &newCls)
	if err != nil {
		log.Println(err)
	} else {
		log.Println(res.)
	}
}
