package top

import(
	"fmt"
	"reflect"
)

type taobaoMap struct {
	val interface{}
	count int64
}

func (m *taobaoMap) result() (r []map[string]interface{}) {
	switch v := m.val.(type) {
	case map[string]interface{}:
		return []map[string]interface{}{v}
	case []map[string]interface{}:
		return v
	default:
		panic(fmt.Sprintf("type not allowed %+v", reflect.TypeOf(m.val)))
	}
	return
}

func (m *taobaoMap) unwrap() *taobaoMap {

	convertedMap, isMap := m.val.(map[string]interface{})

	if !isMap {
		return m
	}

	var count int64
	count = 1

	if m.count != 0 {
		count = m.count
	}

	countObj, countExist := convertedMap["total_results"]

	if countExist {

		switch i := countObj.(type) {
		case float64:
			count = int64(i)
		default:
		}

		delete(convertedMap, "total_results")
	}

	if len(convertedMap) == 1 {
		for _, valObj := range(convertedMap) {

			v, ok := valObj.(map[string]interface{})
			if ok {

				newtaobaoMap := &taobaoMap{v, count}
				return newtaobaoMap.unwrap()
			}

			arrV, arrOk := valObj.([]interface{})

			if arrOk {
				var newMapArray []map[string]interface{}
				for _, v := range(arrV) {
					newMapValue, mapValueOk := v.(map[string]interface{})
					if mapValueOk {
						newMapArray = append(newMapArray, newMapValue)
					}
				}
				newtaobaoMap := &taobaoMap{newMapArray, count}
				return newtaobaoMap
			}
		}
	}
	return m
}
