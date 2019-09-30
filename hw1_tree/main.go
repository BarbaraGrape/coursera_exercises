package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"sort"
	"strconv"
)

type Node struct {
	FileInfo os.FileInfo
	Children []*Node
}

func dirTree(out io.Writer, path string, printFiles bool) error {
	rootName := path
	isFileSuitable := func(fileInfo os.FileInfo) bool { return fileInfo.IsDir() }
	if printFiles {
		isFileSuitable = func(fileInfo os.FileInfo) bool { return true }
	}
	root, err := readDir(rootName, "", isFileSuitable)
	//printChildren(root, "", printFiles, out)
	printNode(root, "", true, false, out)
	return err
}

func printName(node Node, printFile bool) string {
	if node.FileInfo.IsDir() {
		return node.FileInfo.Name()
	}

	if !printFile {
		return ""
	}

	if node.FileInfo.Size() == 0 {
		return node.FileInfo.Name() + "(empty)"
	}
	return node.FileInfo.Name() + "(" + strconv.FormatInt(node.FileInfo.Size(), 10) + "b)"

}

func printNode(root *Node, previousPrefix string, isLast bool, printRoot bool, output io.Writer) {
	// if not nil
	currentPrefix := "├───"
	childPrefix := "│\t"
	if isLast {
		currentPrefix = "└───"
		childPrefix = "\t"
	}
	if printRoot {
		var postfix string
		if root.FileInfo.IsDir() {
			postfix = root.FileInfo.Name()
		} else {
			size := root.FileInfo.Size()
			sizeStr := fmt.Sprintf("%db", size)
			if size == 0 {
				sizeStr = "empty"
			}
			postfix = fmt.Sprintf("%s (%s)", root.FileInfo.Name(), sizeStr)
		}

		fmt.Fprintf(output, "%s%s%s\n", previousPrefix, currentPrefix, postfix)
	} else {
		childPrefix = ""
	}

	for i, child := range root.Children {
		isChildLast := false
		if i == len(root.Children)-1 {
			isChildLast = true
		}

		printNode(child, previousPrefix+childPrefix, isChildLast, true, output)
	}
}

type IsSuitableFunc func(fileInfo os.FileInfo) bool

func readDir(rootName string, ancestors string, isSuitable IsSuitableFunc) (*Node, error) {
	rootName = path.Join(ancestors, rootName)
	fileInfo, err := os.Stat(rootName)
	if !isSuitable(fileInfo) {
		return nil, nil
	}
	node := &Node{fileInfo, nil}
	if err != nil {
		return nil, fmt.Errorf("Unable to get fileInfo: %v", err)
	}
	if !fileInfo.IsDir() {
		return node, nil
	}
	files, err := ioutil.ReadDir(rootName)
	if err != nil {
		return nil, fmt.Errorf("Unable to get subdirectories list: %v", err)
	}

	var children []*Node
	for _, f := range files {
		//fmt.Println(f.Name())
		child, err := readDir(f.Name(), rootName, isSuitable)
		if err != nil {
			return nil, fmt.Errorf("Error in child node: %v", err)
		}
		if child == nil {
			continue
		}
		children = append(children, child)
	}
	sort.Slice(children, func(i, j int) bool {
		return children[i].FileInfo.Name() <= children[j].FileInfo.Name()
	})
	node.Children = children
	return node, err
}

func main() {
	out := os.Stdout
	if !(len(os.Args) == 2 || len(os.Args) == 3) {
		panic("usage go run main.go . [-f]")
	}
	path := os.Args[1]
	printFiles := len(os.Args) == 3 && os.Args[2] == "-f"
	err := dirTree(out, path, printFiles)
	if err != nil {
		panic(err.Error())
	}
}
