package common

import "k8s.io/apimachinery/pkg/util/json"

// SwapTo 通过json tag进行结构体赋值
func SwapTo(request, target interface{}) error {
	dataByte, err := json.Marshal(request)
	if err != nil {
		return err
	}
	err = json.Unmarshal(dataByte, target)
	return err
}
