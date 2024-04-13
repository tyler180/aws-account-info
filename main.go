package main

import (
	"context"
	"os"
	"your-module-path/getallaccounts" // Update this path to match your setup

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/organizations"
	ginadapter "github.com/awslabs/aws-lambda-go-api-proxy/gin"
	"github.com/gin-gonic/gin"
)

var ginLambda *ginadapter.GinLambda

func init() {
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.GET("/all-accounts", func(c *gin.Context) {
		source := c.Query("source") // Query parameter to choose the source
		ctx := c.Request.Context()

		var fetcher getallaccounts.AccountFetcher
		cfg, err := config.LoadDefaultConfig(ctx)
		if err != nil {
			c.JSON(500, gin.H{"error": "unable to load SDK config"})
			return
		}

		if source == "orgs" {
			orgsClient := organizations.NewFromConfig(cfg)
			fetcher = &getallaccounts.OrgsSource{Client: orgsClient}
		} else {
			dynamoClient := dynamodb.NewFromConfig(cfg)
			fetcher = &getallaccounts.DynamoDBSource{
				Client:    dynamoClient,
				TableName: os.Getenv("DYNAMODB_TABLE_NAME"),
			}
		}

		accounts, err := fetcher.FetchAccounts(ctx)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, accounts)
	})

	ginLambda = ginadapter.New(r)
}

func Handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return ginLambda.ProxyWithContext(ctx, req)
}

func main() {
	lambda.Start(Handler)
}
