package main

import (
	"context"
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/mholt/binding"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Room struct {
	ID   primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name string             `bson:"name" json:"name"`
}

func (r *Room) FieldMap(req *http.Request) binding.FieldMap {
	return binding.FieldMap{&r.Name: "name"}
}

func CreateRoom(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
	r := new(Room)

	if err := binding.Bind(req, r); err.Handle(w) {
		return
	}

	// 몽고 DB 세션 생성
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	// 몽고 DB 세션 종료 defer
	defer cancel()

	// 몽고 ID 생성
	r.ID = primitive.NewObjectID()
	// room 정보 저장을 위한 컬렉션 객체 생성
	coll := client.Database("go-web-chat").Collection("rooms")

	// rooms 컬렉션에 room 정보 저장
	_, err := coll.InsertOne(ctx, r)

	if err != nil {
		// Insert 에러시  Insert 추가 필요!
		renderer.JSON(w, http.StatusInternalServerError, err)
		return
	}

	renderer.JSON(w, http.StatusCreated, r)
}

func RetrieveRooms(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
	// MongoDB 세션 생성 (Context with Timeout)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var rooms []Room

	// rooms 컬렉션 접근
	coll := client.Database("go-web-chat").Collection("rooms")

	// 모든 Room 정보 조회
	cursor, err := coll.Find(ctx, bson.D{})
	if err != nil {
		renderer.JSON(w, http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
		return
	}
	defer cursor.Close(ctx)

	// cursor → []Room 변환
	if err := cursor.All(ctx, &rooms); err != nil {
		renderer.JSON(w, http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
		return
	}

	// 조회 결과 반환
	renderer.JSON(w, http.StatusOK, rooms)
}
