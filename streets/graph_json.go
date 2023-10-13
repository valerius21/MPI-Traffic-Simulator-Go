package streets

import (
	"encoding/json"

	"pchpc/utils"
)

func UnmarshalGraphJSON(data []byte) (GraphJSON, error) {
	var r GraphJSON
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *GraphJSON) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

type GraphJSON struct {
	Filename string `json:"filename"`
	Size     int64  `json:"size"`
	Graph    JGraph `json:"graph"`
}

type JGraph struct {
	Vertices []JVertex `json:"vertices"`
	Edges    []JEdge   `json:"edges"`
}

type JEdge struct {
	From     int     `json:"from"`
	To       int     `json:"to"`
	Length   float64 `json:"length"`
	MaxSpeed string  `json:"max_speed"`
	Name     string  `json:"name"`
	ID       string  `json:"osm_id"`
	Data     Data
}

type JVertex struct {
	X  float64 `json:"x"`
	Y  float64 `json:"y"`
	ID int     `json:"osm_id"`
}

type Data struct {
	ID       string
	Name     string
	MaxSpeed float64
	Length   float64
	Map      *utils.HashMap[string, *Vehicle]
}
