package tests

import (
	"Elschool-API/tests/suite"
	apiv1 "github.com/Ilya-Repin/elschooler/protos/gen/api/api.v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

const (
	marksUserId      = "599ce889-11be-4abf-89fb-2940a5b3bfec"
	marksStudentId   = "a7b7de0e-d637-41ef-a26a-3a02294ade44"
	date             = "26.04.2023"
	dayWorstMark     = 4
	it               = "Дезинформатика"
	chemistry        = "Алхимия"
	language         = "Эльфийский язык как государственный язык Лесного Королевства"
	period           = 2
	averageWorstMark = "3"
	finalWorstMark   = 2
)

func TestGetDayMarks(t *testing.T) {
	ctx, st := suite.New(t)

	marksResp, err := st.MarksClient.GetDayMarks(ctx, &apiv1.DayMarksRequest{UserToken: marksUserId, StudentToken: marksStudentId, Date: date})

	require.NoError(t, err)
	assert.NotEmpty(t, marksResp.GetMarks())
	assert.NotEmpty(t, marksResp.GetWorstMark())
	assert.Equal(t, int32(dayWorstMark), marksResp.GetWorstMark())
	marks := marksResp.GetMarks()
	assert.Equal(t, 1, len(marks))
	subjectMarks := marks[language]
	assert.NotEmpty(t, subjectMarks)
	assert.Equal(t, []int32{4, 4, 4, 4, 5, 5, 5}, subjectMarks.GetMarks())
}

func TestGetAverageMarks(t *testing.T) {
	ctx, st := suite.New(t)

	marksResp, err := st.MarksClient.GetAverageMarks(ctx, &apiv1.AverageMarksRequest{UserToken: marksUserId, StudentToken: marksStudentId, Period: period})

	require.NoError(t, err)
	assert.NotEmpty(t, marksResp.GetMarks())
	assert.NotEmpty(t, marksResp.GetWorstMark())
	assert.Equal(t, averageWorstMark, marksResp.GetWorstMark())

	marks := marksResp.GetMarks()
	assert.Equal(t, 3, len(marks))

	assert.Equal(t, "4.2", marks[chemistry])
	assert.Equal(t, "4.82", marks[it])
	assert.Equal(t, "3", marks[language])
}

func TestGetFinalMarks(t *testing.T) {
	ctx, st := suite.New(t)

	marksResp, err := st.MarksClient.GetFinalMarks(ctx, &apiv1.FinalMarksRequest{UserToken: marksUserId, StudentToken: marksStudentId})

	require.NoError(t, err)
	assert.NotEmpty(t, marksResp.GetMarks())
	assert.NotEmpty(t, marksResp.GetWorstMark())
	assert.Equal(t, int32(finalWorstMark), marksResp.GetWorstMark())

	finalMarks := marksResp.GetMarks()
	assert.Equal(t, 3, len(finalMarks))

	marks := finalMarks[chemistry]
	assert.NotEmpty(t, marks)
	assert.Equal(t, []int32{3, 4, 5, 3, 4, 4, 4}, marks.GetMarks())

	marks = finalMarks[it]
	assert.NotEmpty(t, marks)
	assert.Equal(t, []int32{5, 5, 5, 5, 5, 5, 5}, marks.GetMarks())

	marks = finalMarks[language]
	assert.NotEmpty(t, marks)
	assert.Equal(t, []int32{2, 3, 4, 5, 0, 5, 0}, marks.GetMarks())
}
