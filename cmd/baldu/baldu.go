package main

import "container/heap"
import "flag"
import "fmt"
import "log"
import "math"
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
	flag.BoolVar(&verbose, "verbose", false, "")

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
	fringe := queue([]*node{tree})

    for len(selected) < maxNodes && len(fringe) > 0 {
		biggest := heap.Pop(&fringe).(*node)
		if verbose {
			fmt.Println("Next selected: ", biggest.path)
		}

		if biggest.isDir && !biggest.expanded {
			thisSize := float64(biggest.treeSize)
			totalSize := thisSize
			for _, m := range fringe {
				totalSize += float64(m.treeSize)
			}

			// Assume that the number of nodes below all of those currently
			// in the fringe will be proportionate to their tree sizes.
			nodesNeeded := float64(maxNodes - len(selected))
			p := thisSize / totalSize * nodesNeeded
			thisTargetCount := int(math.Min(2, math.Ceil(p)))
			biggest.expand(thisTargetCount)
		}
		if len(selected) + len(biggest.children) <= maxNodes {
			selected = append(selected, biggest.children...)
			for _, c := range biggest.children {
				heap.Push(&fringe, c)
			}
		}
	}

	for _, n := range selected {
		fmt.Printf("%v\t%s\n", n.treeSize, n.path)
	}
}

func (n *node) expand(targetCount int) {
	if verbose {
		fmt.Println("Expanding ", n.path, "; target:", targetCount)
	}

	expanded := 0
	q := []*node{n}
	dirs := []*node{n}
	for expanded < targetCount && len(q) > 0 {
		n = q[0]
		q = q[1:]
		expanded++
		n.expanded = true
		if verbose {
			fmt.Println("Expanded ", n.path)
		}

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
			if child.isDir {
				q = append(q, child)
				dirs = append(dirs, child)
			} else {
				child.treeSize = child.size
			} 
			n.children = append(n.children, child)
		}
	}

	for _, n = range dirs {
		n.setTreeSizes()
	}
}

// setTreeSizes updates the treeSizes of the nodes in the subtree.
func (n *node) setTreeSizes() {
	if (len(n.children) > 0) {
		n.treeSize = n.size
		for _, c := range n.children {
			c.setTreeSizes()
			n.treeSize += c.treeSize
		}
	} else {
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

