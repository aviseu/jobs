package testutils

import (
	"cloud.google.com/go/pubsub"
	"context"
	"fmt"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"google.golang.org/api/option"
	"google.golang.org/api/option/internaloption"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"time"
)

type PubSubSuite struct {
	suite.Suite

	container testcontainers.Container

	badClient      *pubsub.Client
	BadImportTopic *pubsub.Topic

	client             *pubsub.Client
	ImportTopic        *pubsub.Topic
	ImportSubscription *pubsub.Subscription
}

func (suite *PubSubSuite) SetupSuite() {
	suite.container, suite.client, suite.ImportTopic, suite.badClient, suite.BadImportTopic, suite.ImportSubscription = suite.createDependencies(context.Background())
}

func (suite *PubSubSuite) createDependencies(ctx context.Context) (testcontainers.Container, *pubsub.Client, *pubsub.Topic, *pubsub.Client, *pubsub.Topic, *pubsub.Subscription) {
	req := testcontainers.ContainerRequest{
		Image:        "google/cloud-sdk:emulators",
		ExposedPorts: []string{"8085/tcp"},
		Cmd:          []string{"gcloud", "beta", "emulators", "pubsub", "start", "--host-port=0.0.0.0:8085"},
		WaitingFor:   wait.ForLog("Server started").WithStartupTimeout(10 * time.Second),
	}

	c, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	suite.NoError(err)

	host, err := c.Host(ctx)
	suite.NoError(err)

	port, err := c.MappedPort(ctx, "8085")
	suite.NoError(err)

	client, err := pubsub.NewClient(
		ctx,
		"test-project",
		option.WithEndpoint(fmt.Sprintf("%s:%s", host, port.Port())),
		option.WithGRPCDialOption(grpc.WithTransportCredentials(insecure.NewCredentials())),
		option.WithoutAuthentication(),
		option.WithTelemetryDisabled(),
		internaloption.SkipDialSettingsValidation(),
	)
	suite.NoError(err)

	importTopic, err := client.CreateTopic(ctx, "import-test-topic")
	suite.NoError(err)

	importSubscription, err := client.CreateSubscription(ctx, "import-test-sub", pubsub.SubscriptionConfig{
		Topic: importTopic,
	})
	suite.NoError(err)

	badClient, err := pubsub.NewClient(
		ctx,
		"test-project",
		option.WithEndpoint(fmt.Sprintf("%s:%s", host, port.Port())),
		option.WithGRPCDialOption(grpc.WithTransportCredentials(insecure.NewCredentials())),
		option.WithoutAuthentication(),
		option.WithTelemetryDisabled(),
		internaloption.SkipDialSettingsValidation(),
	)
	suite.NoError(err)

	badImportTopic, err := badClient.CreateTopic(ctx, "import-bad-test-topic")
	suite.NoError(err)
	suite.NoError(badClient.Close())

	return c, client, importTopic, badClient, badImportTopic, importSubscription
}

func (suite *PubSubSuite) TearDownSuite() {
	go func() {
		suite.NoError(suite.client.Close())
		suite.NoError(suite.container.Terminate(context.Background()))
	}()
}
