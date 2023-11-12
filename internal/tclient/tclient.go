package tclient

import (
	"context"

	"tonexplorer/pkg/wrapper"

	"github.com/xssnick/tonutils-go/liteclient"
	"github.com/xssnick/tonutils-go/ton"
)

func NewClient() (ton.APIClientWrapped, error) {
	client := liteclient.NewConnectionPool()

	cfg, err := liteclient.GetConfigFromUrl(context.Background(), "https://ton.org/global.config.json")
	if err != nil {
		return nil, wrapper.Wrap(err)
	}

	// connect to mainnet lite servers
	err = client.AddConnectionsFromConfig(context.Background(), cfg)
	if err != nil {
		return nil, wrapper.Wrap(err)
	}

	// initialize ton api lite connection wrapper with full proof checks
	api := ton.NewAPIClient(client, ton.ProofCheckPolicySecure).WithRetry()
	api.SetTrustedBlockFromConfig(cfg)

	return api, nil
}
