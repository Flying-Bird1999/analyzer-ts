package parser

import (
	"main/analyzer/utils"
	"strings"

	"github.com/Zzzen/typescript-go/use-at-your-own-risk/ast"
)

// CallExpression represents a function or method call, tailored to the new requirements.
type CallExpression struct {
	Identifier     string         `json:"identifier"`
	Property       string         `json:"property,omitempty"`
	ArgLen         int            `json:"argLen"`
	Type           string         `json:"type"`
	Raw            string         `json:"raw,omitempty"`
	SourceLocation SourceLocation `json:"sourceLocation"`
}

// NewCallExpression creates a new CallExpression instance.
func NewCallExpression(node *ast.CallExpression, sourceCode string) *CallExpression {
	pos, end := node.Pos(), node.End()
	return &CallExpression{
		SourceLocation: SourceLocation{
			Start: NodePosition{Line: pos, Column: 0},
			End:   NodePosition{Line: end, Column: 0},
		},
		Raw:  utils.GetNodeText(node.AsNode(), sourceCode),
		Type: "call", // Hardcode type to 'call' as requested
	}
}

// reconstructExpression recursively builds a clean identifier string from an expression node.
func reconstructExpression(node *ast.Node, sourceCode string) string {
	if node == nil {
		return ""
	}
	switch node.Kind {
	case ast.KindIdentifier:
		return node.AsIdentifier().Text
	case ast.KindPropertyAccessExpression:
		propAccess := node.AsPropertyAccessExpression()
		left := reconstructExpression(propAccess.Expression, sourceCode)
		right := propAccess.Name().Text()
		if left != "" {
			return left + "." + right
		}
		return right
	default:
		// Fallback for other expression types, clean up whitespace.
		return strings.TrimSpace(utils.GetNodeText(node, sourceCode))
	}
}

// analyzeCallExpression extracts information from an ast.CallExpression node based on the new structure.
func (ce *CallExpression) analyzeCallExpression(node *ast.CallExpression, sourceCode string) {
	if node == nil {
		return
	}

	// Get the number of arguments
	ce.ArgLen = len(node.Arguments.Nodes)

	// Analyze the expression being called to determine Identifier and Property
	expressionNode := node.Expression

	switch expressionNode.Kind {
	case ast.KindIdentifier:
		// This is a simple function call, e.g., myFunc()
		ce.Identifier = expressionNode.AsIdentifier().Text
		ce.Property = ""

	case ast.KindPropertyAccessExpression:
		// This is a method call on an object, e.g., myObj.myMethod()
		propAccess := expressionNode.AsPropertyAccessExpression()
		ce.Identifier = reconstructExpression(propAccess.Expression, sourceCode)
		ce.Property = propAccess.Name().Text()
		ce.Type = "member" // Hardcode type to 'call' as requested

	default:
		// Fallback for other types of expressions (e.g., IIFE, new expressions)
		// We'll record the whole expression as the identifier and leave property empty.
		ce.Identifier = strings.TrimSpace(utils.GetNodeText(expressionNode, sourceCode))
		ce.Property = ""
	}
}
