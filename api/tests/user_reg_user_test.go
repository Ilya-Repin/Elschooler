package tests

import (
	"Elschool-API/tests/suite"
	apiv1 "github.com/Ilya-Repin/elschooler/protos/gen/api/api.v1"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

// e5e79b0d1e2c49f097db4930c9ea8c43

func TestRegisterUser(t *testing.T) {
	ctx, st := suite.New(t)

	service := "testsuite"

	userResp, err := st.UserClient.RegUser(ctx, &apiv1.RegUserRequest{Service: service})

	require.NoError(t, err)
	assert.NotEmpty(t, userResp.GetUserToken())
	parsed, err := uuid.Parse(userResp.GetUserToken())
	require.NoError(t, err)
	assert.Equal(t, uuid.Version(4), parsed.Version())
}

func TestRegisterUserNoService(t *testing.T) {
	ctx, st := suite.New(t)

	service := ""

	userResp, err := st.UserClient.RegUser(ctx, &apiv1.RegUserRequest{Service: service})

	require.Error(t, err)
	assert.Empty(t, userResp.GetUserToken())
	assert.ErrorContains(t, err, "service name required")
}
