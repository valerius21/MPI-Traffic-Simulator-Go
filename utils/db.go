package utils

import (
	"gorm.io/gorm"
)

var db *gorm.DB

// TrafficNode represents a node in the traffic network
type TrafficNode struct {
	ID         int
	Properties string
	NodeId     int
}

// TrafficEdge represents an edge in the traffic network
type TrafficEdge struct {
	ID       int
	Src      string
	Dst      string
	MaxSpeed int
	Length   float32
}

func init() {
	//innerDB, err := gorm.Open(sqlite.Open("/home/valerius/code/hpc/download/database.sqlite"), &gorm.Config{})
	////if err != nil {
	////	panic("failed to connect database")
	////}
	//db = innerDB
}

// GetDb returns the database connection
func GetDb() *gorm.DB {
	return db
}

// GetVertices returns all vertices from the database
func GetVertices() []TrafficNode {
	var result []TrafficNode
	db := GetDb()
	db.Raw("SELECT id, node_id, properties FROM TrafficNode").Scan(&result)

	return result
}

// GetEdges returns all edges from the database
func GetEdges() []TrafficEdge {
	var result []TrafficEdge

	db := GetDb()
	db.Raw("SELECT id, src, dst, maxSpeed, length FROM TrafficEdge").Scan(&result)
	return result
}
