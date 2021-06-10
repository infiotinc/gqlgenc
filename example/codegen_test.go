package example

import (
	"context"
	"example/client"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSubscription(t *testing.T) {
	t.Parallel()

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
	t.Parallel()

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
	t.Parallel()

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

func TestQueryUnion(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	cli, td, _ := splitcli(ctx)
	defer td()

	gql := &client.Client{
		Client: cli,
	}

	res, _, err := gql.GetMedias(ctx)
	if err != nil {
		t.Fatal(err)
	}

	assert.Len(t, res.Medias, 2)

	assert.Equal(t, int64(100), res.Medias[0].Image.Size)
	assert.Equal(t, int64(200), res.Medias[1].Video.Duration)
}

func TestQueryInterface(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	cli, td, _ := splitcli(ctx)
	defer td()

	gql := &client.Client{
		Client: cli,
	}

	res, _, err := gql.GetBooks(ctx)
	if err != nil {
		t.Fatal(err)
	}

	assert.Len(t, res.Books, 2)

	assert.Equal(t, "Some textbook", res.Books[0].Title)
	assert.Equal(t, []string{"course 1", "course 2"}, res.Books[0].Textbook.Courses)

	assert.Equal(t, "Some Coloring Book", res.Books[1].Title)
	assert.Equal(t, []string{"red", "blue"}, res.Books[1].ColoringBook.Colors)
}

func TestMutationInput(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	cli, td, _ := splitcli(ctx)
	defer td()

	gql := &client.Client{
		Client: cli,
	}

	res, _, err := gql.CreatePost(ctx, client.PostCreateInput{
		Text: "some text",
	})
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "some text", res.Post.Text)
}
