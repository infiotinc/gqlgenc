package example

import (
	"context"
	"example/graph"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestSubscription(t *testing.T) {
	ctx := context.Background()

	cli, td, _ := splitcli(ctx)
	defer td()

	time.Sleep(time.Second)

	gql := &graph.Client{
		Client: cli,
	}

	ch, err := gql.SubscribeMessageAdded(ctx)
	if err != nil {
		t.Fatal(err)
	}

	ids := make([]string, 0)

	for msg := range ch {
		if msg.Error != nil {
			t.Fatal(msg.Error)
		}

		ids = append(ids, msg.Data.MessageAdded.ID)
	}

	assert.Len(t, ids, 3)
}

func TestQuery(t *testing.T) {
	ctx := context.Background()

	cli, td, _ := splitcli(ctx)
	defer td()

	time.Sleep(time.Second)

	gql := &graph.Client{
		Client: cli,
	}

	room, err := gql.GetRoom(ctx)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "test", room.Room.Name)
}
