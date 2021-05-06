package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"example/graph/generated"
	"example/graph/model"
	"fmt"
)

func (r *mutationResolver) Post(ctx context.Context, text string, username string, roomName string) (*model.Message, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) Room(ctx context.Context, name string) (*model.Chatroom, error) {
	return &model.Chatroom{
		Name:     name,
		Messages: nil,
	}, nil
}

func (r *subscriptionResolver) MessageAdded(ctx context.Context, roomName string) (<-chan *model.Message, error) {
	ch := make(chan *model.Message)

	fmt.Println("MESSAGE ADDED")

	go func() {
		i := 0
		for {
			if i == 3 {
				close(ch)
				fmt.Println("DONE MESSAGE ADDED")
				return
			}

			msg := &model.Message{
				ID: fmt.Sprintf("msg%v", i),
			}

			select {
			case <-ctx.Done():
				close(ch)
				fmt.Println("DONE ctx")
				return
			case ch <- msg:
				fmt.Println("SEND")
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
