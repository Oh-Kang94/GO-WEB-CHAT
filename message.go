package main

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/julienschmidt/httprouter"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const messageFetchSize = 10

type Message struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	RoomID    primitive.ObjectID `bson:"room_id" json:"room_id"`
	Content   string             `bson:"content" json:"content"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	User      *User              `bson:"user" json:"user"`
}

func (m *Message) CreateMsg() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	// 몽고 DB 세션 종료 defer
	defer cancel()

	m.ID = primitive.NewObjectID()

	m.CreatedAt = time.Now()

	coll := client.Database("go-web-chat").Collection("messages")

	result, err := coll.InsertOne(ctx, m)

	if err != nil {
		return err
	}

	log.Printf("Inserted message with ID: %v", result.InsertedID)

	return nil
}

func RetriveMessages(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	// 몽고 DB 세션 종료 defer
	defer cancel()

	// 쿼리 매개변수로 limit 값 확인

	limit, errs := strconv.Atoi(req.URL.Query().Get("limit"))
	if errs != nil {
		limit = messageFetchSize
	}

	var messages []Message

	coll := client.Database("go-web-chat").Collection("messages")

	// _id역순으로 정렬하여, limit만큼 msg 조회
	// Find 옵션 (id 역순 정렬 + limit)
	findOpts := options.Find().
		SetSort(bson.D{{Key: "_id", Value: -1}}). // -1이 Desc
		SetLimit(int64(limit))

	cursor, err := coll.Find(ctx, bson.D{}, findOpts)
	if err != nil {
		renderer.JSON(w, http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
		return
	}

	defer cursor.Close(ctx)

	if err := cursor.All(ctx, &messages); err != nil {
		renderer.JSON(w, http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
		return
	}

	renderer.JSON(w, http.StatusOK, messages)

}
