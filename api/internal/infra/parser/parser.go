package parser

import (
	"Elschool-API/internal/domain/models"
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"strconv"
	"strings"
)

var (
	ErrCantParse = errors.New("cant parse marks")
)

const (
	biggestMarkInt   = 5
	biggestMarkFloat = 5.0
)

type Parser struct {
}

func New() *Parser {
	return &Parser{}
}

func (p *Parser) ParseDayMarks(date, html string) (marks models.DayMarks, err error) {
	const op = "infra.parser.ParseDayMarks"
	marks.Marks = make(map[string][]int32)
	marks.Date = date
	marks.WorstMark = biggestMarkInt

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return models.DayMarks{}, fmt.Errorf("%s: %w", op, ErrCantParse)
	}

	doc.Find(".GradesTable tbody tr").Each(func(i int, tr *goquery.Selection) {
		lesson := tr.Find(".grades-lesson").Text()
		spans := tr.Find(".grades-marks .mark-span")

		spans.Each(func(i int, span *goquery.Selection) {
			dataContent, _ := span.Attr("data-popover-content")
			dateParts := strings.Split(dataContent, "<p>")
			if len(dateParts) < 2 {
				err = fmt.Errorf("%s: %w", op, ErrCantParse)
				return
			}

			markDate := strings.TrimSpace(strings.Split(dateParts[1], "</p>")[0])[len("Дата проставления: "):]

			if marks.Date == markDate {
				markStr := span.Text()
				mark, parseErr := strconv.ParseInt(strings.TrimSpace(markStr), 10, 32)
				if parseErr != nil {
					err = fmt.Errorf("%s: %w", op, ErrCantParse)
					return
				}

				if marks.WorstMark > int32(mark) {
					marks.WorstMark = int32(mark)
				}

				marks.Marks[lesson] = append(marks.Marks[lesson], int32(mark))
			}
		})
	})

	if err != nil {
		return models.DayMarks{}, err
	}
	if len(marks.Marks) == 0 {
		marks.WorstMark = 0
	}

	return marks, nil
}

func (p *Parser) ParseAverageMarks(period int32, html string) (marks models.AverageMarks, err error) {
	const op = "infra.parser.ParseAverageMarks"

	marks.Marks = make(map[string]string)
	marks.Period = period
	marks.WorstMark = fmt.Sprintf("%.2f", biggestMarkFloat)

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return models.AverageMarks{}, fmt.Errorf("%s: %w", op, ErrCantParse)
	}

	doc.Find(".MobileGrades tbody").Each(func(i int, tbody *goquery.Selection) {
		var periodCount int32
		periodCount = 0
		subject := strings.TrimSpace(tbody.Prev().Text())
		subject = strings.ReplaceAll(subject, "\n", "")

		tbody.Find("tr").Each(func(i int, tr *goquery.Selection) {
			periodCount += 1
			tr.Find(".grades-period-name").Text()

			averageMarkText := tr.Find(".grades-average").Text()
			averageMarkText = strings.Replace(averageMarkText, ",", ".", 1)

			averageMark, parseErr := strconv.ParseFloat(averageMarkText, 32)
			if parseErr != nil {
				if averageMarkText != "" {
					err = fmt.Errorf("%s: %w", op, ErrCantParse)
				}
				return
			}
			worstMark, parseErr := strconv.ParseFloat(marks.WorstMark, 32)
			if parseErr != nil {
				err = fmt.Errorf("W%s: %w", op, ErrCantParse)
				return
			}

			if periodCount == period {
				if worstMark > averageMark {
					marks.WorstMark = averageMarkText
				}
				marks.Marks[subject] = averageMarkText
			}
		})
	})

	if err != nil {

		return models.AverageMarks{}, err
	}

	if len(marks.Marks) == 0 {
		marks.WorstMark = "0"
	}

	return marks, nil
}

func (p *Parser) ParseFinalMarks(html string) (marks models.FinalMarks, err error) {
	const op = "infra.parser.ParseFinalMarks"

	marks.Marks = make(map[string][]int32)
	marks.WorstMark = biggestMarkInt

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return models.FinalMarks{}, fmt.Errorf("%s: %w", op, ErrCantParse)
	}

	doc.Find(".DivForResultsTable .ResultsTable:not(.MobileResults) tbody tr").Each(func(i int, tr *goquery.Selection) {
		lesson := tr.Find("td").First().Text()
		var marksForLesson []int32

		tr.Find("td.results-mark").Each(func(i int, td *goquery.Selection) {
			markText := strings.TrimSpace(td.Text())

			if markText == "" {
				marksForLesson = append(marksForLesson, 0)
			} else {
				mark, parseErr := strconv.Atoi(markText)
				if parseErr != nil {
					err = fmt.Errorf("error parsing mark: %w", parseErr)
					return
				}

				if marks.WorstMark > int32(mark) {
					marks.WorstMark = int32(mark)
				}

				marksForLesson = append(marksForLesson, int32(mark))
			}
		})

		marks.Marks[lesson] = marksForLesson
	})

	if err != nil {
		return models.FinalMarks{}, fmt.Errorf("%s: %w", op, err)
	}

	if len(marks.Marks) == 0 {
		marks.WorstMark = 0
	}

	return marks, nil
}
