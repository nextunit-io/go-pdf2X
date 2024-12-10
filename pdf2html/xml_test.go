package pdf2html_test

import (
	"fmt"
	"math"
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
				Left:     pointerHelperFn(91),
				Text:     pointerHelperFn("2 row - 1 column"),
				BoldText: pointerHelperFn("test-text-bold"),
			},
			{
				Top:      pointerHelperFn(115),
				Left:     pointerHelperFn(209),
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
			MinLeft: 91,
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
			From:                  50,
			To:                    200,
			Columns:               3,
			GetColumnFunc:         pdf2html.GetColumnCalculationWithVariance([]int{100, 200, 400}, 18),
			AllowedHeightVariance: 5,
			FilterFunc:            &filterFunc,
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
			From:                  50,
			To:                    200,
			Columns:               3,
			GetColumnFunc:         pdf2html.GetColumnCalculationWithVariance([]int{100, 200, 400}, 18),
			AllowedHeightVariance: 5,
		}))
	})
}

func TestGetColumnCalculationWithVariance(t *testing.T) {
	testData := []struct {
		Positions []int
		Variance  int
	}{
		{
			Positions: []int{100, 200, 300, 400},
			Variance:  10,
		},
		{
			Positions: []int{200, 400, 600, 800},
			Variance:  5,
		},
		{
			Positions: []int{5, 10, 15},
			Variance:  2,
		},
	}

	for i, test := range testData {
		t.Run(fmt.Sprintf("Run test %d", i), func(t *testing.T) {
			fn := pdf2html.GetColumnCalculationWithVariance(test.Positions, test.Variance)

			for j, position := range test.Positions {
				t.Run(fmt.Sprintf("Run test for lower starting point than expacted %d, %d", i, j), func(t *testing.T) {
					column, err := fn(pdf2html.PdfXmlText{
						Left: pointerHelperFn(int(math.Round(float64(position) - (0.5 * float64(test.Variance)) - 1))),
					})
					assert.Equal(t, "cannot find correct column", err.Error())
					assert.Equal(t, -1, column)
				})

				for m := position - test.Variance/2; m <= position+(test.Variance/2); m++ {
					t.Run(fmt.Sprintf("Run success test with %d (%d.%d)", m, i, j), func(t *testing.T) {
						column, err := fn(pdf2html.PdfXmlText{
							Left: pointerHelperFn(m),
						})
						assert.Nil(t, err)
						assert.Equal(t, j, column)
					})
				}

				t.Run(fmt.Sprintf("Run test for higher starting point than expacted %d, %d", i, j), func(t *testing.T) {
					column, err := fn(pdf2html.PdfXmlText{
						Left: pointerHelperFn(int(math.Round(float64(position) + (0.5 * float64(test.Variance)) + 1))),
					})
					assert.Equal(t, "cannot find correct column", err.Error())
					assert.Equal(t, -1, column)
				})
			}
		})
	}
}

func TestGetColumnCalculationInRanges(t *testing.T) {
	testData := [][]pdf2html.GetColumnCalculationInRangesOption{
		[]pdf2html.GetColumnCalculationInRangesOption{
			{
				From: 95,
				To:   105,
			},
			{
				From: 195,
				To:   205,
			},
			{
				From: 395,
				To:   405,
			},
		},
		[]pdf2html.GetColumnCalculationInRangesOption{
			{
				From: 198,
				To:   202,
			},
			{
				From: 398,
				To:   402,
			},
			{
				From: 598,
				To:   602,
			},
			{
				From: 798,
				To:   802,
			},
			{
				From: 998,
				To:   1002,
			},
		},
		[]pdf2html.GetColumnCalculationInRangesOption{
			{
				From: 4,
				To:   6,
			},
			{
				From: 10,
				To:   18,
			},
			{
				From: 40,
				To:   100,
			},
			{
				From: 200,
				To:   400,
			},
		},
	}

	for i, test := range testData {
		t.Run(fmt.Sprintf("Run test %d", i), func(t *testing.T) {
			fn := pdf2html.GetColumnCalculationInRanges(test)

			for j, position := range test {
				t.Run(fmt.Sprintf("Run test for lower starting point than expacted %d, %d", i, j), func(t *testing.T) {
					column, err := fn(pdf2html.PdfXmlText{
						Left: pointerHelperFn(position.From - 1),
					})
					assert.Equal(t, "cannot find correct column", err.Error())
					assert.Equal(t, -1, column)
				})

				for m := position.From; m <= position.To; m++ {
					t.Run(fmt.Sprintf("Run success test with %d (%d.%d)", m, i, j), func(t *testing.T) {
						column, err := fn(pdf2html.PdfXmlText{
							Left: pointerHelperFn(m),
						})
						assert.Nil(t, err)
						assert.Equal(t, j, column)
					})
				}

				t.Run(fmt.Sprintf("Run test for higher starting point than expacted %d, %d", i, j), func(t *testing.T) {
					column, err := fn(pdf2html.PdfXmlText{
						Left: pointerHelperFn(position.To + 1),
					})
					assert.Equal(t, "cannot find correct column", err.Error())
					assert.Equal(t, -1, column)
				})
			}
		})
	}
}
