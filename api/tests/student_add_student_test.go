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
	userId = "e5e79b0d-1e2c-49f0-97db-4930c9ea8c43"

	addStudentLogin    = "addedStudentValid"
	addStudentPassword = "addedPasswordValid"

	addDoubleStudentLogin    = "doubleStudentLogin"
	addDoubleStudentPassword = "doubleStudentPassword"

	existedStudentId       = "70d0ed1a-25e1-40f7-877d-9fb9ce28969f"
	existedStudentLogin    = "existedStudent"
	existedStudentPassword = "existedPassword"

	invalidLogin    = "invalidLogin"
	invalidPassword = "invalidPassword"
)

func TestAddNewStudent(t *testing.T) {
	ctx, st := suite.New(t)

	studentResp, err := st.StudentClient.AddStudent(ctx, &apiv1.AddStudentRequest{UserToken: userId, Login: addStudentLogin, Password: addStudentPassword})

	require.NoError(t, err)
	assert.NotEmpty(t, studentResp.GetStudentToken())
	parsed, err := uuid.Parse(studentResp.GetStudentToken())
	require.NoError(t, err)
	assert.Equal(t, uuid.Version(4), parsed.Version())
}

func TestAddExistedStudent(t *testing.T) {
	ctx, st := suite.New(t)

	studentResp, err := st.StudentClient.AddStudent(ctx, &apiv1.AddStudentRequest{UserToken: userId, Login: existedStudentLogin, Password: existedStudentPassword})

	require.NoError(t, err)
	assert.NotEmpty(t, studentResp.GetStudentToken())
	parsed, err := uuid.Parse(studentResp.GetStudentToken())
	require.NoError(t, err)
	assert.Equal(t, uuid.Version(4), parsed.Version())
	assert.Equal(t, existedStudentId, studentResp.GetStudentToken())
}

func TestAddStudentDouble(t *testing.T) {
	ctx, st := suite.New(t)

	studentResp1, err := st.StudentClient.AddStudent(ctx, &apiv1.AddStudentRequest{UserToken: userId, Login: addDoubleStudentLogin, Password: addDoubleStudentPassword})

	require.NoError(t, err)
	assert.NotEmpty(t, studentResp1.GetStudentToken())
	parsed1, err := uuid.Parse(studentResp1.GetStudentToken())
	require.NoError(t, err)
	assert.Equal(t, uuid.Version(4), parsed1.Version())

	studentResp2, err := st.StudentClient.AddStudent(ctx, &apiv1.AddStudentRequest{UserToken: userId, Login: addDoubleStudentLogin, Password: addDoubleStudentPassword})

	require.NoError(t, err)
	assert.NotEmpty(t, studentResp2.GetStudentToken())
	parsed2, err := uuid.Parse(studentResp2.GetStudentToken())
	require.NoError(t, err)
	assert.Equal(t, uuid.Version(4), parsed2.Version())

	assert.Equal(t, parsed1, parsed2)
}

func TestAddInvalidStudent(t *testing.T) {
	ctx, st := suite.New(t)

	studentResp, err := st.StudentClient.AddStudent(ctx, &apiv1.AddStudentRequest{UserToken: userId, Login: invalidLogin, Password: invalidPassword})

	require.Error(t, err)
	assert.Empty(t, studentResp.GetStudentToken())
	assert.ErrorContains(t, err, "add student error")
}
