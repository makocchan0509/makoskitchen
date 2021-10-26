package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"log"
	"makoskitchen/kitchen/menu/databases"
	"makoskitchen/kitchen/menu/graph/generated"
	"makoskitchen/kitchen/menu/graph/model"
	"makoskitchen/kitchen/menu/util"
)

func (r *mutationResolver) CreateMenu(ctx context.Context, input model.NewMenu) (string, error) {
	log.Printf("[mutaionResolver.CreateMenu] input: %#v", input)
	id := util.GenUUID()
	err := databases.NewMenuDao(r.DB).InsertOne(&databases.Menu{
		ID:    id,
		Name:  input.Name,
		Price: input.Price,
		Type:  input.Type,
	})
	if err != nil {
		return "", err
	}
	return id, nil
}

func (r *queryResolver) Menus(ctx context.Context) ([]*model.Menu, error) {
	log.Println("[queryResolver.Menus]")
	menus, err := databases.NewMenuDao(r.DB).FindAll()
	if err != err {
		return nil, err
	}
	var results []*model.Menu
	for _, menu := range menus {
		results = append(results, &model.Menu{
			ID:    menu.ID,
			Name:  menu.Name,
			Price: menu.Price,
			Type:  menu.Type,
		})
	}
	return results, nil
}

func (r *queryResolver) Menu(ctx context.Context, id string) (*model.Menu, error) {
	log.Printf("[queryResolver.Menu] id: %#v", id)
	menu, err := databases.NewMenuDao(r.DB).FindOne(id)
	if err != nil {
		return nil, err
	}
	if menu == nil {
		return nil, nil
	}
	return &model.Menu{
		ID:    menu.ID,
		Name:  menu.Name,
		Price: menu.Price,
		Type:  menu.Type,
	}, nil
}

func (r *queryResolver) MenuByType(ctx context.Context, typeArg string) ([]*model.Menu, error) {
	log.Printf("[queryResolver.Menu] type: %#v", typeArg)
	menus, err := databases.NewMenuDao(r.DB).FindByType(typeArg)
	if err != nil {
		return nil, err
	}
	var results []*model.Menu
	for _, menu := range menus {
		results = append(results, &model.Menu{
			ID:    menu.ID,
			Name:  menu.Name,
			Price: menu.Price,
			Type:  menu.Type,
		})
	}
	return results, nil
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
