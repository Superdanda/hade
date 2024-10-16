package framework

import (
	"errors"
	"strings"
)

type Tree struct {
	root *node // 根节点
}

func NewTree() *Tree {
	root := newNode()
	return &Tree{root}
}

type node struct {
	isLast   bool              // 代表这个节点是否可以成为最终的路由规则。该节点是否能成为一个独立的uri, 是否自身就是一个终极节点
	segment  string            // uri中的字符串，代表这个节点表示的路由中某个段的字符串
	handler  ControllerHandler // 代表这个节点中包含的控制器，用于最终加载调用
	children []*node           // 代表这个节点下的子节点
}

func newNode() *node {
	return &node{
		isLast:   false,
		segment:  "",
		children: []*node{},
	}
}

func isWildSegment(segment string) bool {
	return strings.HasPrefix(segment, ":")
}

func (n *node) filterChildNodes(segment string) []*node {
	if len(n.children) == 0 {
		return nil
	}

	if isWildSegment(segment) {
		return n.children
	}

	nodes := make([]*node, 0, len(n.children))
	for _, child := range n.children {
		if isWildSegment(child.segment) {
			nodes = append(nodes, child)
		} else if child.segment == segment {
			nodes = append(nodes, child)
		}
	}
	return nodes
}

func (n *node) matchNode(uri string) *node {
	segments := strings.SplitN(uri, "/", 2) // 使用分隔符将uri切割为两个部分
	segment := segments[0]                  // 第一个部分用于匹配下一层子节点

	if segment == "" {
		return n.matchNode(segments[1])
	}

	if !isWildSegment(segment) {
		segment = strings.ToUpper(segment)
	}

	children := n.filterChildNodes(segment) // 匹配符合的下一层子节点
	if children == nil || len(children) == 0 {
		return nil
	}

	if len(segments) == 1 {
		// 如果segment已经是最后一个节点，判断这些children是否有isLast标志
		for _, tn := range children {
			if tn.isLast {
				return tn
			}
		}
		return nil
	}

	// 如果有2个segment, 递归每个子节点继续进行查找
	for _, child := range children {
		tnMatch := child.matchNode(segments[1])
		if tnMatch != nil {
			return tnMatch
		}
	}
	return nil
}

func (tree *Tree) AddRouter(uri string, handler ControllerHandler) error {
	n := tree.root
	// 确认路由是否冲突
	if n.matchNode(uri) != nil {
		return errors.New("route exist: " + uri)
	}

	segments := strings.Split(uri, "/")

	for index, segment := range segments {

		if segment == "" {
			continue
		}

		if !isWildSegment(segment) {
			segment = strings.ToUpper(segment)
		}

		isLast := index == len(segments)-1
		var objNode *node // 标记是否有合适的子节点
		children := n.filterChildNodes(segment)
		if len(children) > 0 {
			// 如果有segment相同的子节点，则选择这个子节点
			for _, child := range children {
				if child.segment == segment {
					objNode = child
					break
				}
			}
		}

		//如果没有找到合适的子节点就要创建一个新节点
		if objNode == nil {
			child := &node{}
			child.segment = segment
			if isLast {
				child.isLast = true
				child.handler = handler
			}
			n.children = append(n.children, child)
			objNode = child
		}
		n = objNode
	}
	return nil
}

func (tree *Tree) FindHandler(uri string) ControllerHandler {
	match := tree.root.matchNode(uri)
	if match == nil {
		return nil
	}
	return match.handler
}
