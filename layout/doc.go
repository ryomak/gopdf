// Package layout provides types and functions for PDF layout analysis and manipulation.
//
// This package defines the core types for representing page layouts, including
// text blocks, image blocks, and their spatial relationships. It provides
// utilities for:
//
//   - Analyzing page structure and content positioning
//   - Adjusting layouts using various strategies
//   - Managing content blocks and their bounds
//   - Handling coordinate transformations
//
// # Content Blocks
//
// ContentBlock is the unified interface for all page content elements:
//
//	type ContentBlock interface {
//	    Bounds() Rectangle
//	    Type() ContentBlockType
//	    Position() (x, y float64)
//	}
//
// Two types of content blocks are supported:
//
//	ContentBlockTypeText  - Text content
//	ContentBlockTypeImage - Image content
//
// # Layout Strategies
//
// The package supports several layout adjustment strategies:
//
//	StrategyPreservePosition - Preserve original positions
//	StrategyCompact          - Pack content tightly from top
//	StrategyEvenSpacing      - Distribute content evenly
//	StrategyFlowDown         - Flow content downward
//	StrategyFitContent       - Fit content within blocks
package layout
