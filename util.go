package main

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

/*
学生宿舍楼栋：
	西区1-8
	东区1-16，东13东/西、东15东/西、东附一
	元宝山1-5
	南湖1-13
	国交（重点）：4\8\9
*/

// 楼栋名，西3栋，东03栋

// 南4、南5、南13：南学13-105A，南13-224

var (
	// 匹配：
	// 西x-xxx空调，元x-xxx空调，
	// 西5-121照明辅导员，西6-123照明活动室，西7-426 照明，
	// 东4-新401空调辅导员，东16-113辅导员空调，
	// 国4-0114照明公用，
	// 东7-101A空调，东7-101B空调
	commonPattern = `(.+)-.*([1-9][\d]{2,3}[A|B]?).*`

	// 匹配：
	// 国8、国9：国9-3楼-1照明
	doubleHyphenPattern = `(.+)-([1-9]).-([\d]{1,2}).*`

	// 匹配：
	// 东13-东501A空调，东13-西102照明，东15-西501A空调
	hyphenIncludePattern = `(.{2,3}-.)([\d]{2,3}).*`

	// 匹配：103空调
	directRoomMatchPattern = `([\d]{3}).+`
)

// 处理寝室名，格式：东16-101
// 返回：
//     东16
//     101
//     error
func ProcessDormName(s string) (string, string, error) {
	var flag = 0
	var pattern = commonPattern

	if strings.Contains(s, "国8") || strings.Contains(s, "国9") {
		flag = 1
		pattern = doubleHyphenPattern
	} else if strings.Contains(s, "东13") || strings.Contains(s, "东15") {
		flag = 2
		pattern = hyphenIncludePattern
	} else if !strings.Contains(s, "-") {
		flag = 3
		pattern = directRoomMatchPattern
	}

	rgx := regexp.MustCompile(pattern)
	matchGroups := rgx.FindStringSubmatch(s)

	if len(matchGroups) == 0 {
		return "", "", errors.New("mathch failed")
	}

	if flag == 1 && len(matchGroups) < 4 ||
		flag == 3 && len(matchGroups) < 2 ||
		(flag == 0 || flag == 2) && len(matchGroups) < 3 {

		return "", "", errors.New("length is wrong")
	}

	if flag == 3 {
		// 东07栋的 “103空调” 只有宿舍号
		return "", matchGroups[1], nil
	} else if flag == 1 {
		roomNum, err := strconv.Atoi(matchGroups[3])
		if err != nil {
			return "", "", err
		}
		room := fmt.Sprintf("%s%02d", matchGroups[2], roomNum)
		return matchGroups[1], room, nil
	}

	return matchGroups[1], matchGroups[2], nil
}

// 判断空调/照明
func JudgeMeterKind(s string) string {
	var kind string

	if strings.Contains(s, "照明") {
		kind = "light"
	} else if strings.Contains(s, "空调") {
		kind = "air"
	} else if strings.Contains(s, "南") {
		kind = "air"
	}

	return kind
}
