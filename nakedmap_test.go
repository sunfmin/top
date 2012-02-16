package top

import (
	"testing"
	"encoding/json"
)

var nakedmapTests = []struct {
	jsonIn string
 	existKeys []string
	count string
}{
	{`{"response":{ "data": { "name": "Felix", "gender": "Man"}}}`, []string{"gender", "name"}, "1"},
	{`{"taobaoke_items_convert_response":{"taobaoke_items":{"taobaoke_item":[{"click_url":"http://s.click.taobao.com/t_8?e=7HZ6jHSTbIQ1Foq6TZKhbPgDOyt1AzILzvAKsVn%2B8AHubbgZmH6ogOs4aK0SlekAExSG1t9VJD3Ki0Gl7hB5WACvfhKaUqfLsduq80eGHYLvPdXowQ%3D%3D&p=mm_30129436_0_0&n=19","commission":"5.39","commission_num":"0","commission_rate":"150.00","commission_volume":"0.00","num_iid":14009743597}]},"total_results":1}}`, []string{"commission", "commission_rate"}, "1"},
}


func TestNakedMap(t *testing.T) {
	for _, tt := range(nakedmapTests) {
		var result map[string]interface{}
		err := json.Unmarshal([]byte(tt.jsonIn), &result)
		if err != nil {
			t.Errorf("%+v", err)
		}

		r, count := taobaoMap(result).unwrap()

		if count != tt.count {
			t.Errorf("expected count is %+v, but was %+v", tt.count, count)
		}

		v, ok := r.(map[string]interface{})
		if !ok {
			vArray, aOk := r.([]map[string]interface{})
			if aOk {
				v = vArray[0]
			}
		}

		for _, k := range(tt.existKeys) {
			_, keyOk := v[k]
			if !keyOk {
				t.Errorf("expected %+v has key %+v, but there was none.", v, k)
			}
		}
	}
}

