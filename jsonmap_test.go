package main

import (
	"testing"
)

func TestWithInvalidJSON(t *testing.T) {
	invalidJSONReply := MakeJSONReply([]byte("I'm no json"))

	if mm := invalidJSONReply.MakeMap(); mm == nil {
		t.Errorf("Json is invalid. Empty map should be returned. Value: %v", mm)
	}
	if qi := invalidJSONReply.MakeStatsInfo(); qi == nil {
		t.Errorf("Json is invalid. Empty map should be returned. Value: %v", qi)
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
	reply := MakeJSONReply([]byte(`{"FloatKey":4, "st":"string","nes":{"ted":5}}`))

	flMap := reply.MakeMap()
	checkMap(flMap, t, 0)
}

func TestMakeStatsInfo(t *testing.T) {
	reply := MakeJSONReply([]byte(`[{"name":"q1", "FloatKey":14,"nes":{"ted":15}},{"name":"q2", "vhost":"foo", "FloatKey":24,"nes":{"ted":25}}]`))

	qinfo := reply.MakeStatsInfo()
	if qinfo[0].name != "q1" {
		t.Errorf("unexpected qinfo name: %v", qinfo[0].name)
	}
	if qinfo[1].name != "q2" {
		t.Errorf("unexpected qinfo name: %v", qinfo[0].name)
	}
	if qinfo[1].vhost != "foo" {
		t.Errorf("unexpected qinfo name: %v", qinfo[0].name)
	}
	checkMap(qinfo[0].metrics, t, 10)
	checkMap(qinfo[1].metrics, t, 20)
}
