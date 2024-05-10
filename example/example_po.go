package example

import (
	"encoding/json"
)

type Example struct {
	ID   uint   `json:"id" gorm:"primarykey"`
	Key_ string `json:"key" gorm:"column:key"`
	Data string `json:"data" gorm:"column:data"`
}

// 获取对应的表名
func (e *Example) TableName() string {
	return "example"
}

// 获取 key 对应的字段名
func (e *Example) KeyColumn() string {
	return "key"
}

// 获取 key 对应的值
func (e *Example) Key() string {
	return e.Key_
}

func (e *Example) DataColumn() []string {
	return []string{"data"}
}

// 将 object 序列化成字符串
func (e *Example) Write() (string, error) {
	body, err := json.Marshal(e)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

// 读取字符串内容，反序列化到 object 实例中
func (e *Example) Read(body string) error {
	return json.Unmarshal([]byte(body), e)
}
