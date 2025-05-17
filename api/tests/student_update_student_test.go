package tests

import (
	"Elschool-API/tests/suite"
	apiv1 "github.com/Ilya-Repin/elschooler/protos/gen/api/api.v1"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

const (
	updateUserExistId = "7e889d15-8dda-4d72-82ae-70d380b580ff"

	originStudentId       = "62f486db-a0b8-4ce9-8101-d718bedea98f"
	targetStudentId       = "103dfbac-e79e-42ff-aef7-5627d1badc63"
	targetStudentLogin    = "testStudentUpdateExisted"
	targetStudentPassword = "testPasswordUpdateExisted"

	updateUserNewId          = "531bce8d-43b9-4ac5-bbc3-e47977189bf1"
	updateStudentId          = "964810cf-984a-41aa-8050-9d3c6dab81bf"
	updateStudentLoginNew    = "testStudentUpdate2New"
	updateStudentPasswordNew = "testPasswordUpdate2New"
)

func TestUpdateStudentToExisted(t *testing.T) {
	ctx, st := suite.New(t)

	studentResp, err := st.StudentClient.UpdateStudent(ctx, &apiv1.UpdateStudentRequest{UserToken: updateUserExistId, StudentToken: originStudentId, Login: targetStudentLogin, Password: targetStudentPassword})

	require.NoError(t, err)
	assert.NotEmpty(t, studentResp.GetStudentToken())
	parsed, err := uuid.Parse(studentResp.GetStudentToken())
	require.NoError(t, err)
	assert.Equal(t, uuid.Version(4), parsed.Version())
	assert.Equal(t, targetStudentId, parsed.String())
}

func TestUpdateStudentToNew(t *testing.T) {
	ctx, st := suite.New(t)

	studentResp, err := st.StudentClient.UpdateStudent(ctx, &apiv1.UpdateStudentRequest{UserToken: updateUserNewId, StudentToken: updateStudentId, Login: updateStudentLoginNew, Password: updateStudentPasswordNew})

	require.NoError(t, err)
	assert.NotEmpty(t, studentResp.GetStudentToken())
	parsed, err := uuid.Parse(studentResp.GetStudentToken())
	require.NoError(t, err)
	assert.Equal(t, uuid.Version(4), parsed.Version())
}
