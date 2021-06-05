package example

import (
	"context"
	"example/client"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSubscription(t *testing.T) {
	ctx := context.Background()

	cli, td, _ := splitcli(ctx)
	defer td()

	gql := &client.Client{
		Client: cli,
	}

	ch, stop := gql.SubscribeMessageAdded(ctx)
	defer stop()

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

	gql := &client.Client{
		Client: cli,
	}

	room, _, err := gql.GetRoom(ctx, "test")
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "test", room.Room.Name)
}

func TestQueryCustomType(t *testing.T) {
	ctx := context.Background()

	cli, td, _ := splitcli(ctx)
	defer td()

	gql := &client.Client{
		Client: cli,
	}

	room, _, err := gql.GetRoomCustom(ctx, "test")
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "Room: test", room.String())
}
