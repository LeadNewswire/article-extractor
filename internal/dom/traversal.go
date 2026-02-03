package dom

import (
	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
)

// GetParent returns the parent selection.
func GetParent(sel *goquery.Selection) *goquery.Selection {
	return sel.Parent()
}

// GetGrandparent returns the grandparent selection.
func GetGrandparent(sel *goquery.Selection) *goquery.Selection {
	return sel.Parent().Parent()
}

// GetAncestors returns all ancestor selections up to a certain depth.
func GetAncestors(sel *goquery.Selection, maxDepth int) []*goquery.Selection {
	var ancestors []*goquery.Selection
	current := sel.Parent()
	depth := 0

	for current.Length() > 0 && depth < maxDepth {
		ancestors = append(ancestors, current)
		current = current.Parent()
		depth++
	}

	return ancestors
}

// GetSiblings returns all sibling selections.
func GetSiblings(sel *goquery.Selection) *goquery.Selection {
	return sel.Siblings()
}

// GetPreviousSiblings returns all previous sibling selections.
func GetPreviousSiblings(sel *goquery.Selection) *goquery.Selection {
	return sel.PrevAll()
}

// GetNextSiblings returns all next sibling selections.
func GetNextSiblings(sel *goquery.Selection) *goquery.Selection {
	return sel.NextAll()
}

// GetChildren returns all child selections.
func GetChildren(sel *goquery.Selection) *goquery.Selection {
	return sel.Children()
}

// GetDescendants returns all descendant selections matching a selector.
func GetDescendants(sel *goquery.Selection, selector string) *goquery.Selection {
	return sel.Find(selector)
}

// ForEachNode iterates over each node in the selection.
func ForEachNode(sel *goquery.Selection, fn func(int, *goquery.Selection)) {
	sel.Each(fn)
}

// FilterNodes filters nodes based on a predicate.
func FilterNodes(sel *goquery.Selection, fn func(*goquery.Selection) bool) *goquery.Selection {
	return sel.FilterFunction(func(_ int, s *goquery.Selection) bool {
		return fn(s)
	})
}

// FindByTag finds all elements with the given tag name.
func FindByTag(sel *goquery.Selection, tag string) *goquery.Selection {
	return sel.Find(tag)
}

// FindByClass finds all elements with the given class.
func FindByClass(sel *goquery.Selection, class string) *goquery.Selection {
	return sel.Find("." + class)
}

// FindByID finds the element with the given ID.
func FindByID(sel *goquery.Selection, id string) *goquery.Selection {
	return sel.Find("#" + id)
}

// HasClass checks if the selection has a specific class.
func HasClass(sel *goquery.Selection, class string) bool {
	return sel.HasClass(class)
}

// GetNodeType returns the node type of the first element.
func GetNodeType(sel *goquery.Selection) html.NodeType {
	if sel.Length() == 0 {
		return html.ErrorNode
	}
	nodes := sel.Nodes
	if len(nodes) == 0 {
		return html.ErrorNode
	}
	return nodes[0].Type
}

// IsElementNode checks if the node is an element node.
func IsElementNode(sel *goquery.Selection) bool {
	return GetNodeType(sel) == html.ElementNode
}

// IsTextNode checks if the node is a text node.
func IsTextNode(sel *goquery.Selection) bool {
	return GetNodeType(sel) == html.TextNode
}

// WalkTree walks the DOM tree depth-first.
func WalkTree(sel *goquery.Selection, fn func(*goquery.Selection) bool) {
	var walk func(*goquery.Selection)
	walk = func(s *goquery.Selection) {
		s.Each(func(_ int, node *goquery.Selection) {
			if fn(node) {
				node.Children().Each(func(_ int, child *goquery.Selection) {
					walk(child)
				})
			}
		})
	}
	walk(sel)
}
