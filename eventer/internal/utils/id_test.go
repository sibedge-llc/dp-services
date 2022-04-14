package utils

import (
	"testing"
)

func TestToJsonId(t *testing.T) {
	var data1 = `{"name": "A", "struct":{"number": 7, "flag": true}, "null": null, "number": 1.0, "slice": [{"name": "1"}, {"name": "2"}], "dummy": []}`
	var data2 = `{"dummy": [], "null": null, "name": "A", "slice": [{  "name": "2"}, {"name": "1"}], "struct":{"flag": true, "number": 7}, "number": 1}`
	id1, err := ToJsonId([]byte(data1), false)
	if err != nil {
		t.Fatal("data1 get id failed", err)
	}
	id2, err := ToJsonId([]byte(data2), false)
	if err != nil {
		t.Fatal("data2 get id failed", err)
	}
	if id1 != id2 {
		t.Errorf("id1[%d] != id2[%d]", id1, id2)
	}
}

func BenchmarkIgnoreSliceOrderToJsonId(b *testing.B) {
	var data = `{"name": "A", "struct":{"number": 7, "flag": true}, "null": null, "number": 1.0, "slice": [{"name": "1"}, {"name": "2"}], "dummy": []}`
	for i := 0; i < b.N; i++ {
		_, err := ToJsonId([]byte(data), false)
		if err != nil {
			b.Fatal("data1 get id failed", err)
		}
	}
}

func BenchmarkKeepSliceOrderToJsonId(b *testing.B) {
	var data = `{"name": "A", "struct":{"number": 7, "flag": true}, "null": null, "number": 1.0, "slice": [{"name": "1"}, {"name": "2"}], "dummy": []}`
	for i := 0; i < b.N; i++ {
		_, err := ToJsonId([]byte(data), true)
		if err != nil {
			b.Fatal("data1 get id failed", err)
		}
	}
}
