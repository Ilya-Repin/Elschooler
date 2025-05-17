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
	complexStudentLogin       = "complexStudentLogin"
	complexStudentPassword    = "complexStudentPassword"
	newComplexStudentLogin    = "newComplexStudentLogin"
	newComplexStudentPassword = "newComplexStudentPassword"
)

func TestComplexStudent(t *testing.T) {
	ctx, st := suite.New(t)

	service := "testsuite"
	userResp, err := st.UserClient.RegUser(ctx, &apiv1.RegUserRequest{Service: service})
	require.NoError(t, err)
	userToken := userResp.GetUserToken()
	assert.NotEmpty(t, userToken)
	parsed, err := uuid.Parse(userToken)
	require.NoError(t, err)
	assert.Equal(t, uuid.Version(4), parsed.Version())

	addResp, err := st.StudentClient.AddStudent(ctx, &apiv1.AddStudentRequest{UserToken: userToken, Login: complexStudentLogin, Password: complexStudentPassword})
	require.NoError(t, err)
	studentToken := addResp.GetStudentToken()
	assert.NotEmpty(t, studentToken)
	parsed, err = uuid.Parse(studentToken)
	require.NoError(t, err)
	assert.Equal(t, uuid.Version(4), parsed.Version())

	updateResp, err := st.StudentClient.UpdateStudent(ctx, &apiv1.UpdateStudentRequest{UserToken: userToken, StudentToken: studentToken, Login: newComplexStudentLogin, Password: newComplexStudentPassword})
	require.NoError(t, err)
	updatedStudentToken := updateResp.GetStudentToken()
	assert.NotEmpty(t, updatedStudentToken)
	parsed, err = uuid.Parse(updatedStudentToken)
	require.NoError(t, err)
	assert.Equal(t, uuid.Version(4), parsed.Version())
	assert.NotEqual(t, studentToken, updatedStudentToken)

	deleteResp, err := st.StudentClient.DeleteStudent(ctx, &apiv1.DeleteStudentRequest{UserToken: userToken, StudentToken: updatedStudentToken})
	require.NoError(t, err)
	assert.NotEmpty(t, deleteResp.GetSuccess())
	assert.True(t, deleteResp.GetSuccess())
}
