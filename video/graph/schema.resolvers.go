package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"

	"nkonev.name/video/graph/generated"
	"nkonev.name/video/graph/model"
)

// CreateLinkVideo is the resolver for the createLinkVideo field.
func (r *mutationResolver) CreateLinkVideo(ctx context.Context, input model.NewLinkVideo) (*model.LinkVideo, error) {
	var link model.LinkVideo
	link.Address = input.Address
	link.Title = input.Title + " VIDEO!!!"
	return &link, nil
}

// LinksVideo is the resolver for the linksVideo field.
func (r *queryResolver) LinksVideo(ctx context.Context) ([]*model.LinkVideo, error) {
	var links []*model.LinkVideo
	dummyLink := model.LinkVideo{
		Title:   "our VIDEO dummy link",
		Address: "https://video-address.org",
	}
	links = append(links, &dummyLink)
	return links, nil
}

// SubscribeVideo is the resolver for the subscribeVideo field.
func (r *subscriptionResolver) SubscribeVideo(ctx context.Context, subscriber string) (<-chan string, error) {
	panic(fmt.Errorf("not implemented: SubscribeVideo - subscribeVideo"))
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

// !!! WARNING !!!
// The code below was going to be deleted when updating resolvers. It has been copied here so you have
// one last chance to move it out of harms way if you want. There are two reasons this happens:
//   - When renaming or deleting a resolver the old code will be put in here. You can safely delete
//     it when you're done.
//   - You have helper methods in this file. Move them out to keep these resolver files clean.
func (r *mutationResolver) CreateLink(ctx context.Context, input model.NewLinkVideo) (*model.LinkVideo, error) {
	panic(fmt.Errorf("not implemented: CreateLink - createLink"))
}
func (r *queryResolver) Links(ctx context.Context) ([]*model.LinkVideo, error) {
	panic(fmt.Errorf("not implemented: Links - links"))
}
