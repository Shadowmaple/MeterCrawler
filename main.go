package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

var file *os.File

func main() {
	mongoURI := "mongodb://127.0.0.1:27017/?compressors=disabled&gssapiServiceName=mongodb"
	DB = GetMongoDB(mongoURI)

	// if err := NewDBData("东14", "233", "1111.1", "air"); err != nil {
	// 	fmt.Println(err)
	// 	return
	// }

	// println("OK")

	// fmt.Println(ProcessDormName("东6-活动室5照明"))

	var err error
	file, err = os.OpenFile("data.log", os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	Crawler()
}

func Crawler() {
	// 西区/0001，东区/0002，元宝山/0003，南湖/0004，国际园区/0006
	var areas = []string{"0001", "0002", "0003", "0004", "0006"}

	for _, area := range areas {
		fmt.Printf("\n========= %s ==========\n", area)
		fmt.Fprintf(file, "\n========= %s ==========\n", area)

		CrawlerInOneArea(area)
	}
}

func CrawlerInOneArea(area string) {
	buildings, err := MakeArchitecureRequest(area)
	if err != nil {
		log.Println(err)
		return
	}

	if len(buildings.List) == 0 {
		log.Printf("no building data in area %s", area)
		return
	}

	// 遍历楼栋
	for _, building := range buildings.List {
		// 忽略楼栋
		if SkipBuilding(area, building.Name) {
			continue
		}

		fmt.Println("--- 楼栋：", building)
		fmt.Fprintln(file, "--- 楼栋：", building)
		CrawlerInOneBuilding(building)

		time.Sleep(10)
	}
}

func CrawlerInOneBuilding(building *ArchitectureInfo) {
	// 遍历楼层
	for floor := building.BottomFloor; floor <= building.TopFloor; floor++ {
		// 获取宿舍
		rooms, err := MakeDormitoryRequest(building.ID, strconv.Itoa(floor))
		if err != nil {
			msg := fmt.Sprintf("Get building %s floor %d error: %s", building.Name, floor, err.Error())
			log.Println(msg)
			fmt.Fprintf(file, msg)
			return
		}

		if len(rooms.Dorms) == 0 {
			msg := fmt.Sprintf("no room data in the building %s", building.ID)
			log.Println(msg)
			fmt.Fprintf(file, msg)
			return
		}

		// 遍历某层寝室
		for _, room := range rooms.Dorms {
			if SkipRoom(room.Name) {
				log.Println(room.Name)
				fmt.Fprintln(file, "------------- skip -------", room.Name)
				continue
			}

			buildingName, roomNum, err := ProcessDormName(room.Name)
			if err != nil {
				log.Println(err, room.Name)
				fmt.Fprintln(file, "------------- error -------", err, room.Name)
				continue
			}

			// 东7有客厅空调：101空调，无楼栋名
			if buildingName == "" && building.Name == "东07栋" {
				buildingName = "东7"
			}

			// 南3 -> 南湖3
			buildingName = strings.Replace(buildingName, "南", "南湖", 1)

			// 类型，照明还是空调
			kind := JudgeMeterKind(room.Name)

			// 无法判断，断定非学生在用，或是跳过的宿舍楼
			if kind == "" || buildingName == "" || roomNum == "" {
				fmt.Fprintln(file, "------------- skip -------", room.Name)
				continue
			}

			fmt.Println(buildingName, roomNum, kind)
			// if _, err := fmt.Fprintln(file, buildingName, roomNum, "-->", kind); err != nil {
			// 	log.Fatal(err)
			// 	return
			// }

			meterID, err := MakeMeterInfoRequest(room.ID)
			if err != nil {
				msg := fmt.Sprintf("Get room %s error: %s", room.Name, err.Error())
				log.Println(msg)
				fmt.Fprintln(file, msg)
				return
			}

			fmt.Printf("%s --> %s", room.Name, meterID)

			if err := NewDBData(buildingName, roomNum, meterID, kind); err != nil {
				msg := fmt.Sprintf("Room=%s %s meterID=%s add to mongo error: %s", buildingName, roomNum, meterID, err.Error())
				log.Println(msg)
				fmt.Fprintln(file, msg)
				continue
			}
			// fmt.Println("Add OK")

			time.Sleep(10)
		}
		time.Sleep(10)
	}
}

// 忽略楼栋
func SkipBuilding(area, building string) bool {
	// 国交暂时只考虑国4、国8、国9
	// var needed = []string{"国4栋", "国8栋", "国9栋"}
	if area == "0006" && building != "国4栋" && building != "国8栋" && building != "国9栋" {
		return true
	}

	var skips = []string{"东23栋"}
	for _, s := range skips {
		if s == building {
			return true
		}
	}
	return false
}

// 忽略宿舍
func SkipRoom(room string) bool {
	if strings.Contains(room, "南学") {
		return true
	}
	return false
}

func NewDBData(building, room, meterID, kind string) error {
	if err := AddMeterData(building, room, meterID, kind); err != nil {
		return err
	}
	return nil
}
