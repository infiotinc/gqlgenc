package server

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"errors"
	"example/server/generated"
	"example/server/model"
	"fmt"
	"os"
	"strconv"
)

func (r *mutationResolver) Post(ctx context.Context, input model.PostCreateInput) (*model.Message, error) {
	return &model.Message{
		Text: input.Text,
	}, nil
}

func (r *queryResolver) Room(ctx context.Context, name string) (*model.Chatroom, error) {
	if name == "error" {
		return nil, errors.New("that's an invalid room")
	}

	return &model.Chatroom{
		Name:     name,
		Messages: nil,
	}, nil
}

func (r *queryResolver) RoomNonNull(ctx context.Context, name string) (*model.Chatroom, error) {
	return r.Room(ctx, name)
}

func (r *queryResolver) Medias(ctx context.Context) ([]model.Media, error) {
	return []model.Media{
		&model.Image{
			Size: 100,
		},
		&model.Video{
			Duration: 200,
		},
	}, nil
}

func (r *queryResolver) Books(ctx context.Context) ([]model.Book, error) {
	return []model.Book{
		&model.Textbook{
			Title:   "Some textbook",
			Courses: []string{"course 1", "course 2"},
		},
		&model.ColoringBook{
			Title:  "Some Coloring Book",
			Colors: []string{"red", "blue"},
		},
	}, nil
}

func (r *subscriptionResolver) MessageAdded(ctx context.Context, roomName string) (<-chan *model.Message, error) {
	ch := make(chan *model.Message)
	debug, _ := strconv.ParseBool(os.Getenv("GQLGENC_WS_LOG"))

	debugPrint := func(a ...interface{}) {
		if debug {
			fmt.Println(a...)
		}
	}

	debugPrint("MESSAGE ADDED")

	go func() {
		i := 0
		for {
			if i == 3 {
				close(ch)
				debugPrint("DONE MESSAGE ADDED")
				return
			}

			msg := &model.Message{
				ID: fmt.Sprintf("msg%v", i),
			}

			select {
			case <-ctx.Done():
				close(ch)
				debugPrint("DONE ctx")
				return
			case ch <- msg:
				debugPrint("SEND")
				i++
			}
		}
	}()

	return ch, nil
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

// Subscription returns generated.SubscriptionResolver implementation.
func (r *Resolver) Subscription() generated.SubscriptionResolver { return &subscriptionResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
type subscriptionResolver struct{ *Resolver }
