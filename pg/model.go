package pg

import "github.com/wcharczuk/go-incr"

type Graph struct {
	ID                 incr.Identifier
	StabilizationNum   uint64
	NumNodes           uint64
	NumNodesRecomputed uint64
	NumNodesChanged    uint64
}

type Node struct {
	ID      incr.Identifier
	GraphID incr.Identifier
}

type NodeRelationship struct {
	GraphID  incr.Identifier
	ParentID incr.Identifier
}
