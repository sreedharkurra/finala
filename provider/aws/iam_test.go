package aws_test

import (
	"errors"
	"finala/provider/aws"
	"finala/testutils"
	"strings"
	"testing"
	"time"

	awsClient "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
)

var defaultUsersMock = iam.ListUsersOutput{
	Users: []*iam.User{
		{UserName: awsClient.String("foo")},
		{UserName: awsClient.String("foo2")},
		{UserName: awsClient.String("test")},
	},
}

type MockIAMClient struct {
	errListUser             error
	errListAccessKeys       error
	errGetAccessKeyLastUsed error
}

func (im *MockIAMClient) ListUsers(input *iam.ListUsersInput) (*iam.ListUsersOutput, error) {

	return &defaultUsersMock, im.errListUser

}

func (im *MockIAMClient) ListAccessKeys(input *iam.ListAccessKeysInput) (*iam.ListAccessKeysOutput, error) {

	response := iam.ListAccessKeysOutput{
		AccessKeyMetadata: []*iam.AccessKeyMetadata{
			{
				AccessKeyId: input.UserName,
			},
		},
	}
	return &response, im.errListAccessKeys

}

func (im *MockIAMClient) GetAccessKeyLastUsed(input *iam.GetAccessKeyLastUsedInput) (*iam.GetAccessKeyLastUsedOutput, error) {
	now := time.Now()

	lastUsedDate := now.AddDate(0, 0, -1)
	if strings.HasPrefix(*input.AccessKeyId, "foo") {
		lastUsedDate = now.AddDate(0, -1, 0)
	}
	response := iam.GetAccessKeyLastUsedOutput{
		AccessKeyLastUsed: &iam.AccessKeyLastUsed{
			LastUsedDate: &lastUsedDate,
		},
	}
	return &response, im.errGetAccessKeyLastUsed

}

func TestDescribeUsers(t *testing.T) {
	mockStorage := testutils.NewMockStorage()

	t.Run("valid", func(t *testing.T) {
		mockClient := MockIAMClient{}
		iamManager := aws.NewIAMUseranager(&mockClient, mockStorage)
		response, _ := iamManager.GetUsers(nil, nil)

		if len(response) != len(defaultUsersMock.Users) {
			t.Fatalf("unexpected user count, got %d expected %d", len(response), len(defaultUsersMock.Users))
		}
	})

	t.Run("error", func(t *testing.T) {

		mockClient := MockIAMClient{
			errListUser: errors.New("error"),
		}

		iamManager := aws.NewIAMUseranager(&mockClient, mockStorage)
		_, err := iamManager.GetUsers(nil, nil)

		if err == nil {
			t.Fatalf("unexpected describe Instances error, return empty")
		}
	})

}

func TestLastActivity(t *testing.T) {

	mockStorage := testutils.NewMockStorage()

	t.Run("valid", func(t *testing.T) {
		mockClient := MockIAMClient{}
		iamManager := aws.NewIAMUseranager(&mockClient, mockStorage)
		response, _ := iamManager.LastActivity(10, ">=")

		if len(response) != 2 {
			t.Fatalf("unexpected user detection, got %d expected %d", len(response), 2)
		}

	})

}
