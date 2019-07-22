package run

import (
	"log"

	"github.com/sotah-inc/steamwheedle-cartel/pkg/messenger"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/metric"

	"github.com/sotah-inc/steamwheedle-cartel/pkg/hell"

	"cloud.google.com/go/storage"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/sotah/gameversions"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/state"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/store"
	"github.com/twinj/uuid"
)

type GatewayStateConfig struct {
	ProjectId string

	MessengerHost string
	MessengerPort int
}

func NewGatewayState(config GatewayStateConfig) (GatewayState, error) {
	// establishing an initial state
	sta := GatewayState{
		State: state.NewState(uuid.NewV4(), true),
	}

	var err error

	// connecting to hell
	sta.IO.HellClient, err = hell.NewClient(config.ProjectId)
	if err != nil {
		log.Fatalf("Failed to connect to firebase: %s", err.Error())

		return GatewayState{}, err
	}

	sta.actEndpoints, err = sta.IO.HellClient.GetActEndpoints()
	if err != nil {
		log.Fatalf("Failed to fetch act endpoints: %s", err.Error())

		return GatewayState{}, err
	}

	// connecting to the messenger host
	mess, err := messenger.NewMessenger(config.MessengerHost, config.MessengerPort)
	if err != nil {
		return GatewayState{}, err
	}
	sta.IO.Messenger = mess

	// initializing a reporter
	sta.IO.Reporter = metric.NewReporter(mess)

	// initializing a store client
	sta.IO.StoreClient, err = store.NewClient(config.ProjectId)
	if err != nil {
		log.Fatalf("Failed to create new store client: %s", err.Error())

		return GatewayState{}, err
	}

	sta.bootBase = store.NewBootBase(sta.IO.StoreClient, "us-central1")
	sta.bootBucket, err = sta.bootBase.GetFirmBucket()
	if err != nil {
		log.Fatalf("Failed to get firm bucket: %s", err.Error())

		return GatewayState{}, err
	}

	sta.realmsBase = store.NewRealmsBase(sta.IO.StoreClient, "us-central1", gameversions.Retail)
	sta.realmsBucket, err = sta.realmsBase.GetFirmBucket()
	if err != nil {
		log.Fatalf("Failed to get firm bucket: %s", err.Error())

		return GatewayState{}, err
	}

	return sta, nil
}

type GatewayState struct {
	state.State

	bootBase     store.BootBase
	bootBucket   *storage.BucketHandle
	realmsBase   store.RealmsBase
	realmsBucket *storage.BucketHandle

	actEndpoints hell.ActEndpoints
}
