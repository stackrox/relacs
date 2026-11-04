package konflux

import (
	"context"
	"testing"

	"github.com/h2non/gock"
	"github.com/stackrox/relacs/tests/helpers"
	"github.com/stretchr/testify/suite"
)

func TestUsernameTestSuite(t *testing.T) {
	suite.Run(t, new(UsernameTestSuite))
}

type UsernameTestSuite struct {
	suite.Suite
	ctx context.Context
}

func (s *UsernameTestSuite) SetupTest() {
	s.ctx = context.Background()
}

func (s *UsernameTestSuite) TearDownTest() {
	helpers.TearDownGock()
}

func (s *UsernameTestSuite) TestGetCurrentUsername_Success() {
	helpers.MockSelfSubjectReview(helpers.TestK8sAPIServerURL, "user@redhat.com")

	testClient := helpers.SetupTestClient()

	username, err := testClient.GetKonflux().GetCurrentUsername(s.ctx)

	s.NoError(err)
	s.Equal("user@redhat.com", username)
	s.True(gock.IsDone())
}

func (s *UsernameTestSuite) TestGetCurrentUsername_AuthFailure() {
	helpers.MockSelfSubjectReviewAuthFailure(helpers.TestK8sAPIServerURL)

	testClient := helpers.SetupTestClient()

	_, err := testClient.GetKonflux().GetCurrentUsername(s.ctx)

	s.Error(err)
	s.Contains(err.Error(), "401")
	s.Contains(err.Error(), "oc login")
	s.True(gock.IsDone())
}

func (s *UsernameTestSuite) TestGetCurrentUsername_EmptyUsername() {
	helpers.MockSelfSubjectReview(helpers.TestK8sAPIServerURL, "")

	testClient := helpers.SetupTestClient()

	_, err := testClient.GetKonflux().GetCurrentUsername(s.ctx)

	s.Error(err)
	s.Contains(err.Error(), "empty username")
	s.True(gock.IsDone())
}
