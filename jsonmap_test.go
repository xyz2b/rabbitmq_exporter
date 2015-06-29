package main

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestWithInvalidJSON(t *testing.T) {
	invalidJSONDecoder := json.NewDecoder(strings.NewReader("I'm no json"))

	if mm := MakeMap(invalidJSONDecoder); mm == nil {
		t.Errorf("Json is invalid. Empty map should be returned. Value: %v", mm)
	}
	if qm := MakeQueueMap(invalidJSONDecoder); qm == nil {
		t.Errorf("Json is invalid. Empty map should be returned. Value: %v", qm)
	}

	if mm := MakeMap(nil); mm == nil {
		t.Errorf("Empty map should be returned. Value: %v", mm)
	}
	if qm := MakeQueueMap(nil); qm == nil {
		t.Errorf("Empty map should be returned.. Value: %v", qm)
	}
}

func checkMap(flMap map[string]float64, t *testing.T, addValue float64) {
	if flMap == nil {
		t.Error("Map should not be nil")
	}

	if v := flMap["FloatKey"]; v != addValue+4 {
		t.Errorf("Map should contain FloatKey but found '%v'", v)
	}

	if v := flMap["nes.ted"]; v != addValue+5 {
		t.Errorf("Map should contain nes.ted key but found '%v'", v)
	}

	if v, ok := flMap["st"]; ok {
		t.Errorf("key 'st' should not be included in map as it contains a string. Value: %v", v)
	}
}

func TestMakeMap(t *testing.T) {
	jsonObject := strings.NewReader(`{"FloatKey":4, "st":"string","nes":{"ted":5}}`)
	decoder := json.NewDecoder(jsonObject)

	flMap := MakeMap(decoder)
	checkMap(flMap, t, 0)
}

func TestMakeQueueMap(t *testing.T) {
	jsonArray := strings.NewReader(`[{"name":"q1", "FloatKey":14,"nes":{"ted":15}},{"name":"q2", "FloatKey":24,"nes":{"ted":25}}]`)
	decoder := json.NewDecoder(jsonArray)

	qmap := MakeQueueMap(decoder)
	if _, ok := qmap["q1"]; !ok {
		t.Error("map should contain queue map. queue=q1")
	}
	if _, ok := qmap["q2"]; !ok {
		t.Error("map should contain queue map. queue=q2")
	}
	checkMap(qmap["q1"], t, 10)
	checkMap(qmap["q2"], t, 20)
}
