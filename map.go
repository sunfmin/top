package top

import(
	"fmt"
	"reflect"
	"strconv"
)

type Map map[string]interface{}

func (m Map) ValueAsString(key string) string {
	switch v := m[key].(type) {
	case nil:
		return ""
	case string:
		return v
	case float64:
		return strconv.FormatFloat(v, 'f', 0, 64)
	default:
		panic(fmt.Sprintf("type can not conver to string %+v", reflect.TypeOf(v)))
	}
	return ""
}

func (m Map) ValueAsMap(key string) Map {
	switch v := m[key].(type) {
	case nil:
		return nil
	case map[string]interface{}:
		return Map(v)
	default:
		panic(fmt.Sprintf("type can not conver to map[string]interface{} %+v", reflect.TypeOf(v)))
	}
	return nil
}

type taobaoMap struct {
	val interface{}
	count int64
}

func (m *taobaoMap) result() (r []Map) {
	switch v := m.val.(type) {
	case map[string]interface{}:
		return []Map{Map(v)}
	case []map[string]interface{}:
		var mapArr []Map
		for _, av := range(v) {
			mapArr = append(mapArr, Map(av))
		}
		return mapArr
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
