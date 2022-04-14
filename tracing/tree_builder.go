package trace

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type TreeBuilder struct {
	Trace *Trace

	rootNode Trace_Node
	nodes    map[string]NodeMap

	stopped   bool
	startTime *time.Time
	mu        sync.Mutex
}

type NodeMap struct {
	self   *Trace_Node
	parent *Trace_Node
}

func NewTreeBuilder() *TreeBuilder {
	tb := TreeBuilder{
		rootNode: Trace_Node{},
	}

	t := Trace{
		Root: &tb.rootNode,
	}
	tb.nodes = make(map[string]NodeMap)
	tb.nodes[""] = NodeMap{self: &tb.rootNode, parent: nil}

	tb.Trace = &t

	return &tb
}

func (tb *TreeBuilder) StartTimer() {
	if tb.startTime != nil {
		fmt.Println(fmt.Errorf("StartTimer called twice"))
	}
	if tb.stopped {
		fmt.Println(fmt.Errorf("StartTimer called after StopTimer"))
	}
	ts := time.Now().UTC()
	tb.Trace.StartTime = timestamppb.New(ts)
	tb.startTime = &ts
}

func (tb *TreeBuilder) StopTimer() {
	if tb.startTime == nil {
		fmt.Println(fmt.Errorf("StopTimer called before StartTimer"))
	}
	if tb.stopped {
		fmt.Println(fmt.Errorf("StopTimer called twice"))

	}
	ts := time.Now().UTC()
	tb.Trace.DurationNs = uint64(ts.Sub(*tb.startTime).Nanoseconds())
	tb.Trace.EndTime = timestamppb.New(ts)
	tb.stopped = true
}

func (tb *TreeBuilder) WillResolveField(ctx context.Context) {
	if tb.startTime == nil {
		fmt.Println(fmt.Errorf("WillResolveField called before StartTimer"))
		return
	}
	if tb.stopped {
		fmt.Println(fmt.Errorf("WillResolveField called after StopTimer"))
		return
	}
	fc := graphql.GetFieldContext(ctx)

	node := tb.newNode(fc)
	node.StartTime = uint64(time.Since(*tb.startTime).Nanoseconds())
	defer func() {
		node.EndTime = uint64(time.Since(*tb.startTime).Nanoseconds())
	}()

	node.Type = fc.Field.Definition.Type.String()
	if fc.Parent != nil {
		node.ParentType = fc.Object
	}

}

func (tb *TreeBuilder) newNode(path *graphql.FieldContext) *Trace_Node {
	if path.Path().String() == "" {
		return &tb.rootNode
	}

	self := &Trace_Node{}
	pn := tb.ensureParentNode(path)

	if path.Index != nil {
		self.Id = &Trace_Node_Index{Index: uint32(*path.Index)}
	} else {
		self.Id = &Trace_Node_ResponseName{ResponseName: path.Field.Name}
	}

	// lock the map from being read concurrently to avoid panics
	tb.mu.Lock()
	nodeRef := tb.nodes[path.Path().String()]
	nodeRef.parent = pn
	nodeRef.self = self

	nodeRef.parent.Child = append(nodeRef.parent.Child, self)
	nodeRef.self = self
	tb.nodes[path.Path().String()] = nodeRef
	//once finisehd writing, unlock
	tb.mu.Unlock()
	return self
}

func (tb *TreeBuilder) ensureParentNode(path *graphql.FieldContext) *Trace_Node {
	// lock to read briefly
	tb.mu.Lock()
	nodeRef := tb.nodes[path.Parent.Path().String()]
	// unlock
	tb.mu.Unlock()
	if nodeRef.self != nil {
		return nodeRef.self
	}

	return tb.newNode(path.Parent)
}
