package models

type DayMarks struct {
	Marks     map[string][]int32
	WorstMark int32
	Date      string
}

type AverageMarks struct {
	Marks     map[string]string
	WorstMark string
	Period    int32
}

type FinalMarks struct {
	Marks     map[string][]int32
	WorstMark int32
	//year int
}
