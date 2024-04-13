package getallaccounts

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/organizations"
)

type Account struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Arn  string `json:"arn"`
}

// AccountFetcher defines an interface for fetching account information.
type AccountFetcher interface {
	FetchAccounts(ctx context.Context) ([]Account, error)
}

type DynamoDBSource struct {
	Client    *dynamodb.Client
	TableName string
}

type OrgsSource struct {
	Client *organizations.Client
}

// Implement FetchAccounts for DynamoDB
func (d *DynamoDBSource) FetchAccounts(ctx context.Context) ([]Account, error) {
	out, err := d.Client.Scan(ctx, &dynamodb.ScanInput{
		TableName: aws.String(d.TableName),
	})
	if err != nil {
		return nil, err
	}

	var accounts []Account
	for _, item := range out.Items {
		account := Account{
			ID:   aws.ToString(item["ID"].(*types.AttributeValueMemberS)),
			Name: aws.ToString(item["Name"].(*types.AttributeValueMemberS)),
			Arn:  aws.ToString(item["Arn"].(*types.AttributeValueMemberS)),
		}
		accounts = append(accounts, account)
	}
	return accounts, nil
}

// Implement FetchAccounts for AWS Organizations
func (o *OrgsSource) FetchAccounts(ctx context.Context) ([]Account, error) {
	paginator := organizations.NewListAccountsPaginator(o.Client, &organizations.ListAccountsInput{})
	var accounts []Account

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, acc := range page.Accounts {
			accounts = append(accounts, Account{
				ID:   *acc.Id,
				Name: *acc.Name,
				Arn:  *acc.Arn,
			})
		}
	}

	return accounts, nil
}
