package main

import "container/heap"
import "flag"
import "fmt"
import "log"
import "os"
import "path/filepath"

type node struct {
	path string
	size int64
	treeSize int64
	children []*node
	isDir bool
	expanded bool
}

type queue []*node

var verbose bool

func main() {
	flag.BoolVar(&verbose, "verbose", true, "")

    var maxNodes int = 0
	flag.IntVar(&maxNodes, "max", 100, "Maximum number of files displayed")

	flag.Parse()
	if flag.NArg() < 1 || maxNodes == 0 {
		fmt.Println("Usage: baldu -max NUM [-verbose] PATH")
		os.Exit(1)
	}

	root := flag.Arg(0)
	tree := &node{path: root, isDir: true}
	tree.expand(maxNodes)
	selected := []*node{tree}
	fringe := []*node{tree}

    for len(selected) < maxNodes && len(fringe) > 0 {
		biggest := heap.Pop(queue(fringe)).(*node)
		if biggest.isDir && !biggest.expanded {
			// TODO: Expand all unexpanded nodes on fringe, with
			// count allocated proportionately to size.
			// XXX Is this the best approach?
		}
		if len(selected) + len(biggest.children) <= maxNodes {
			selected = append(selected, biggest.children...)
			for _, c := range biggest.children {
				heap.Push(queue(fringe), c)
			}
		}
	}
}

func (n *node) expand(targetCount int) {
	expanded := 0
	q := []*node{n}
	unsized := []*node{}
	for expanded < targetCount && len(q) > 0 {
		n = q[0]
		q = q[1:]
		expanded++
		n.expanded = true
		f, err := os.Open(n.path)
		if err != nil {
			log.Println(err)
			continue
		}
		childInfo, err := f.Readdir(0)
		if err != nil {
			log.Println(err)
		}
		_ = f.Close()
		n.children = make([]*node, 0, len(childInfo))
		for _, c := range childInfo {
			child := &node{path: filepath.Join(n.path, c.Name()),
			               size: c.Size(), isDir: c.IsDir()}
			if !child.isDir {
				child.treeSize = child.size
			}
			n.children = append(n.children, child)
			unsized = append(unsized, child)
		}
	}
	for _, n = range q {
		if n.treeSize == 0 {
			unsized = append(unsized, n)
		}
	}

	for _, n = range unsized {
		n.treeSize = treeSize(n.path, n.size, n.isDir)
	}
}

// treeSize returns the total size of the file at the path and any files in it,
// if it's a directory.
func treeSize(path string, size int64, isDir bool) int64 {
	// The size of the root node has already been checked, so we just add
	// the children.
	if !isDir {
		return size
	}

	f, err := os.Open(path)
	if err != nil {
		log.Println(err)
		return size
	}

	childInfo, err := f.Readdir(0)
	if err != nil {
		log.Println(err)
	}
	f.Close()
	for _, c := range childInfo {
		p := filepath.Join(path, c.Name())
		size += treeSize(p, c.Size(), c.IsDir())
	}
	return size
}

func (q queue) Len() int {
	return len(q)
}

func (q queue) Less(i, j int) bool {
	// It's actually a max-heap.
	return q[i].treeSize > q[j].treeSize
}

func (q queue) Swap(i, j int) {
	q[i], q[j] = q[j], q[i]
}

func (q *queue) Push(x interface{}) {
	*q = append(*q, x.(*node))
}

func (q *queue) Pop() interface{} {
	n := len(*q) - 1
	x := (*q)[n]
	*q = (*q)[:n]
	return x
}

