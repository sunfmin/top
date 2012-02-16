package top

import(
	"fmt"
)

type taobaoMap map[string]interface{}

func (m taobaoMap) unwrap() (r interface{}, count string) {
	convertedMap := map[string]interface{}(m)

	fmt.Println("%+v", convertedMap)
	countObj, countExist := convertedMap["total_results"]
	if countExist {
		countStr, countIsString := countObj.(string)
		if countIsString {
			fmt.Printf("%+v", countStr)
			count = countStr
		}
		delete(convertedMap, "total_results")
	}

	if len(convertedMap) == 1 {
		for k, _ := range(convertedMap) {
			v, ok := convertedMap[k].(map[string]interface{})
			fmt.Println("%+v", v)
			if ok {
				return taobaoMap(v).unwrap(), count
			}
		}
	}
	return convertedMap, count
}
