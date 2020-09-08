package main

import (
	"encoding/xml"
	"io/ioutil"
	"net/http"
)

// 宿舍信息
type DormList struct {
	Dorms []DormInfo `xml:"roomInfoList>RoomInfo"`
}

// 宿舍信息
type DormInfo struct {
	ID   string `xml:"RoomNo"`
	Name string `xml:"RoomName"`
}

// 楼栋信息
type ArchitectureList struct {
	List []*ArchitectureInfo `xml:"architectureInfoList>architectureInfo"`
}

// 楼栋信息
type ArchitectureInfo struct {
	ID          string `xml:"ArchitectureID"`
	Name        string `xml:"ArchitectureName"`
	TopFloor    int    `xml:"ArchitectureStorys"` // 最高层数
	BottomFloor int    `xml:"ArchitectureBegin"`  // 最低的层数
}

// 请求获取楼栋信息
func MakeArchitecureRequest(areaId string) (*ArchitectureList, error) {
	var result ArchitectureList
	url := "http://jnb.ccnu.edu.cn/icbs/PurchaseWebService.asmx/getArchitectureInfo?Area_ID=" + areaId

	if err := MakeRequest(url, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// 请求获取寝室信息
func MakeDormitoryRequest(architectureId string, floorId string) (*DormList, error) {
	var result DormList
	url := "http://jnb.ccnu.edu.cn/icbs/PurchaseWebService.asmx/" +
		"getRoomInfo?Architecture_ID=" + architectureId + "&Floor=" + floorId

	if err := MakeRequest(url, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

type MeterInfoPayload struct {
	MeterId string `xml:"meterList>MeterInfo>meterId"` // 电表号 id
}

// 请求获取电表信息
func MakeMeterInfoRequest(roomId string) (string, error) {
	var data MeterInfoPayload
	url := "http://jnb.ccnu.edu.cn/icbs/PurchaseWebService.asmx/getMeterInfo?Room_ID=" + roomId

	if err := MakeRequest(url, &data); err != nil {
		return "", err
	}

	return data.MeterId, nil
}

// 发起请求，解析数据
func MakeRequest(url string, data interface{}) error {
	// 发送 HTTP GET 请求
	body, err := SendHTTPGetRequest(url)
	if err != nil {
		return err
	}

	// 解析 XML body data
	if err := UnmarshalXMLBody(body, data); err != nil {
		return err
	}

	return nil
}

// SendHTTPGetRequest send HTTP GET request.
func SendHTTPGetRequest(requestURL string) ([]byte, error) {
	resp, err := http.Get(requestURL)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

// UnmarshalXMLBody unmarshal body data with XML.
func UnmarshalXMLBody(body []byte, data interface{}) error {
	return xml.Unmarshal(body, data)
}
