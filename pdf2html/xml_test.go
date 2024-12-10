package pdf2html_test

import (
	"testing"

	"github.com/nextunit-io/go-pdf2X/pdf2html"
	"github.com/stretchr/testify/assert"
)

var (
	xmlPage = pdf2html.PdfXmlPage{
		Texts: []pdf2html.PdfXmlText{
			{
				Top:      pointerHelperFn(20),
				Left:     pointerHelperFn(100),
				Text:     pointerHelperFn("test-should-be-above-and-removed"),
				BoldText: pointerHelperFn("test-text-bold"),
			},
			{
				Top:      pointerHelperFn(400),
				Left:     pointerHelperFn(100),
				Text:     pointerHelperFn("test-should-be-below-and-removed"),
				BoldText: pointerHelperFn("test-text-bold"),
			},
			{
				Top:      pointerHelperFn(100),
				Left:     pointerHelperFn(100),
				Text:     pointerHelperFn("1 row - 1 column"),
				BoldText: pointerHelperFn("test-text-bold"),
			},
			{
				Top:      pointerHelperFn(102),
				Left:     pointerHelperFn(200),
				Text:     pointerHelperFn("1 row - 2 column"),
				BoldText: pointerHelperFn("test-text-bold"),
			},
			{
				Top:      pointerHelperFn(98),
				Left:     pointerHelperFn(400),
				Text:     pointerHelperFn("1 row - 3 column"),
				BoldText: pointerHelperFn("test-text-bold"),
			},
			{
				Top:      pointerHelperFn(115),
				Left:     pointerHelperFn(99),
				Text:     pointerHelperFn("2 row - 1 column"),
				BoldText: pointerHelperFn("test-text-bold"),
			},
			{
				Top:      pointerHelperFn(115),
				Left:     pointerHelperFn(202),
				Text:     pointerHelperFn("2 row - 2 column"),
				BoldText: pointerHelperFn("test-text-bold"),
			},
			{
				Top:      pointerHelperFn(115),
				Left:     pointerHelperFn(403),
				Text:     pointerHelperFn("2 row - 3 column"),
				BoldText: pointerHelperFn("test-text-bold"),
			},
			{
				Top:      pointerHelperFn(115),
				Left:     pointerHelperFn(50),
				Text:     pointerHelperFn("same line but outside the table (before) - removed"),
				BoldText: pointerHelperFn("test-text-bold"),
			},
			{
				Top:      pointerHelperFn(115),
				Left:     pointerHelperFn(410),
				Text:     pointerHelperFn("same line but outside the table (after) - removed"),
				BoldText: pointerHelperFn("test-text-bold"),
			},

			{
				Top:      pointerHelperFn(140),
				Left:     pointerHelperFn(100),
				Text:     pointerHelperFn("3 row - 1 column - filter func remove (if filter func is set)"),
				BoldText: pointerHelperFn("test-text-bold"),
			},
			{
				Top:      pointerHelperFn(140),
				Left:     pointerHelperFn(200),
				Text:     pointerHelperFn("3 row - 2 column - filter func remove (if filter func is set)"),
				BoldText: pointerHelperFn("test-text-bold"),
			},
			{
				Top:      pointerHelperFn(140),
				Left:     pointerHelperFn(300),
				Text:     pointerHelperFn("3 row - 3 column - filter func remove (if filter func is set) but not matching a column"),
				BoldText: pointerHelperFn("test-text-bold"),
			},
			{
				Top:      pointerHelperFn(500),
				Left:     pointerHelperFn(100),
				Text:     pointerHelperFn("test-should-be-below-2-and-removed"),
				BoldText: pointerHelperFn("test-text-bold"),
			},
		},
	}

	expectedXmlTable = []*pdf2html.PdfXmlTableEntry{
		{
			MinLeft: 100,
			MaxLeft: 400,
			MinTop:  98,
			MaxTop:  102,
			Content: []*pdf2html.PdfXmlTableEntryContent{
				{
					Text:     pointerHelperFn("1 row - 1 column"),
					BoldText: pointerHelperFn("test-text-bold"),
				},
				{
					Text:     pointerHelperFn("1 row - 2 column"),
					BoldText: pointerHelperFn("test-text-bold"),
				},
				{
					Text:     pointerHelperFn("1 row - 3 column"),
					BoldText: pointerHelperFn("test-text-bold"),
				},
			},
		},
		{
			MinLeft: 50,
			MaxLeft: 403,
			MinTop:  115,
			MaxTop:  115,
			Content: []*pdf2html.PdfXmlTableEntryContent{
				{
					Text:     pointerHelperFn("2 row - 1 column"),
					BoldText: pointerHelperFn("test-text-bold"),
				},
				{
					Text:     pointerHelperFn("2 row - 2 column"),
					BoldText: pointerHelperFn("test-text-bold"),
				},
				{
					Text:     pointerHelperFn("2 row - 3 column"),
					BoldText: pointerHelperFn("test-text-bold"),
				},
			},
		},
	}
)

func TestExtractTableContent(t *testing.T) {
	t.Helper()

	// Should remove the whole 3rd row, since it has the top 140 everywhere
	filterFunc := func(entry pdf2html.PdfXmlTableEntry) bool {
		return entry.MinTop != 140
	}

	t.Run("Check with filter function", func(t *testing.T) {
		assert.Equal(t, expectedXmlTable, xmlPage.ExtractTableContent(pdf2html.PdfXmlTableOption{
			From:                       50,
			To:                         200,
			Columns:                    3,
			ColumnAveragePosition:      []int{100, 200, 400},
			AllowedHeightVariance:      5,
			ColumnAllowedWidthVariance: 10,
			FilterFunc:                 &filterFunc,
		}))
	})

	t.Run("Check without filter function", func(t *testing.T) {
		entryContent := make([]*pdf2html.PdfXmlTableEntryContent, 3)
		entryContent[0] = &pdf2html.PdfXmlTableEntryContent{
			Text:     pointerHelperFn("3 row - 1 column - filter func remove (if filter func is set)"),
			BoldText: pointerHelperFn("test-text-bold"),
		}
		entryContent[1] = &pdf2html.PdfXmlTableEntryContent{
			Text:     pointerHelperFn("3 row - 2 column - filter func remove (if filter func is set)"),
			BoldText: pointerHelperFn("test-text-bold"),
		}
		enhancedExpectedXmlTable := expectedXmlTable
		enhancedExpectedXmlTable = append(enhancedExpectedXmlTable, &pdf2html.PdfXmlTableEntry{
			MinLeft: 100,
			MaxLeft: 200,
			MinTop:  140,
			MaxTop:  140,
			Content: entryContent,
		})

		assert.Equal(t, enhancedExpectedXmlTable, xmlPage.ExtractTableContent(pdf2html.PdfXmlTableOption{
			From:                       50,
			To:                         200,
			Columns:                    3,
			ColumnAveragePosition:      []int{100, 200, 400},
			AllowedHeightVariance:      5,
			ColumnAllowedWidthVariance: 10,
		}))
	})
}
