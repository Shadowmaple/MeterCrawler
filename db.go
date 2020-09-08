package main

import (
	"context"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var DB *mongo.Client

func GetMongoDB(uri string) *mongo.Client {
	clientOptions := options.Client().ApplyURI(uri)

	// Connect to MongoDB
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Println("Database connection failed." + err.Error())
		return nil
	}

	// Check the connection
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Println("Database connection failed." + err.Error())
		return nil
	}

	log.Println("Connected to MongoDB!")
	return client
}

// ----------------------------------------------------

const (
	MongoDB  = "electricity"
	MeterCol = "meter"
)

var (
	MeterCollection *mongo.Collection
)

// 电表号储存结构
type MeterModel struct {
	Building string       `bson:"building"` // 楼栋，如东16
	Room     string       `bson:"room"`     // 寝室，如101
	Meters   []*MeterInfo `bson:"meters"`
}

type MeterInfo struct {
	Kind    string `bson:"kind"`     // 类型，light/air
	MeterID string `bson:"meter_id"` // 电表号
}

// 添加电表信息
// func (m *MeterModel) Create() error {
// 	collection := DB.Database(MongoDB).Collection(MeterCol)

// 	if _, err := collection.InsertOne(context.TODO(), &m); err != nil {
// 		return err
// 	}

// 	return nil
// }

func AddMeterData(building, room, meterID, kind string) error {
	collection := DB.Database(MongoDB).Collection(MeterCol)

	newMeterInfo := &MeterInfo{
		Kind:    kind,
		MeterID: meterID,
	}

	// 有记录则为替换，无记录就插入
	haveDoc, err := HasDormitoryData(building, room)
	if err != nil {
		return err
	}
	fmt.Println(haveDoc)

	if !haveDoc {
		_, err = collection.InsertOne(context.TODO(), MeterModel{
			Building: building,
			Room:     room,
			Meters: []*MeterInfo{
				newMeterInfo,
			},
		})
		return err
	}

	// 获取已存在的数据
	document, err := GetMeterModelByBuildingAndRoom(building, room)
	if err != nil {
		return err
	}

	fmt.Println(document)

	// 查看该电表号类型是否已存在
	for _, meter := range document.Meters {
		if meter.Kind == kind {
			return nil
		}
	}

	// 添加新
	document.Meters = append(document.Meters, newMeterInfo)

	fmt.Println(document)

	_, err = collection.ReplaceOne(context.TODO(), bson.M{"building": building, "room": room}, document)
	if err != nil {
		return err
	}

	return nil
}

// 获取mongodb中的寝室电表信息
func GetMeterModelByBuildingAndRoom(building, room string) (*MeterModel, error) {
	collection := DB.Database(MongoDB).Collection(MeterCol)

	single := collection.FindOne(context.TODO(), bson.M{"building": building, "room": room})

	var result MeterModel
	if err := single.Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

// 查看 mongodb 中是否有该宿舍信息
func HasDormitoryData(building, room string) (bool, error) {
	collection := DB.Database(MongoDB).Collection(MeterCol)

	count, err := collection.CountDocuments(context.TODO(), bson.M{
		"building": building,
		"room":     room,
	})
	if err != nil {
		return false, err
	}

	return count != 0, nil
}
