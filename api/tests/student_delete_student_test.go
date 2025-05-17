package tests

import (
	"Elschool-API/tests/suite"
	apiv1 "github.com/Ilya-Repin/elschooler/protos/gen/api/api.v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

const (
	deleteStudentUserId = "531bce8d-43b9-4ac5-bbc3-e47977189bf1"

	deleteStudentId       = "4aaa94b8-7c71-4371-a2b2-5a91a44e18fa"
	deleteStudentLogin    = "testStudentDelete"
	deleteStudentPassword = "testPasswordDelete"

	popularStudentUserId      = "7e889d15-8dda-4d72-82ae-70d380b580ff"
	popularStudentId          = "7000f792-8145-4157-af2c-149d26b647a5"
	popularStudentLoginNew    = "popularStudentDelete"
	popularStudentPasswordNew = "popularPasswordDelete"
)

func TestDeleteSingleStudent(t *testing.T) {
	ctx, st := suite.New(t)

	studentResp, err := st.StudentClient.DeleteStudent(ctx, &apiv1.DeleteStudentRequest{UserToken: deleteStudentUserId, StudentToken: deleteStudentId})

	require.NoError(t, err)
	assert.NotEmpty(t, studentResp.GetSuccess())
	assert.True(t, studentResp.GetSuccess())
}

func TestDeletePopularStudent(t *testing.T) {
	ctx, st := suite.New(t)

	studentResp, err := st.StudentClient.DeleteStudent(ctx, &apiv1.DeleteStudentRequest{UserToken: deleteStudentUserId, StudentToken: popularStudentId})

	require.NoError(t, err)
	assert.NotEmpty(t, studentResp.GetSuccess())
	assert.True(t, studentResp.GetSuccess())

	studentResp, err = st.StudentClient.DeleteStudent(ctx, &apiv1.DeleteStudentRequest{UserToken: popularStudentUserId, StudentToken: popularStudentId})

	require.NoError(t, err)
	assert.NotEmpty(t, studentResp.GetSuccess())
	assert.True(t, studentResp.GetSuccess())
}
