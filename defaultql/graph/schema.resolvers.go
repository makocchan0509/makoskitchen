package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"log"
	"makoskitchen/defaultql/addmodel"
	"makoskitchen/defaultql/database"
	"makoskitchen/defaultql/graph/generated"
	"makoskitchen/defaultql/graph/model"
	"makoskitchen/defaultql/util"
)

func (r *mutationResolver) CreateTodo(ctx context.Context, input model.NewTodo) (string, error) {
	log.Printf("[mutationResolver.CreateTodo] input: %#v", input)
	id := util.CreateUniqueID()
	err := database.NewTodoDao(r.DB).InsertOne(&database.Todo{
		ID:     id,
		Text:   input.Text,
		Done:   false,
		UserID: input.UserID,
	})
	if err != nil {
		return "", err
	}
	return id, nil
}

func (r *mutationResolver) CreateUser(ctx context.Context, input model.NewUser) (string, error) {
	log.Printf("[mutationResolver.CreateUser] input: %#v", input)
	id := util.CreateUniqueID()
	err := database.NewUserDao(r.DB).InsertOne(&database.User{
		ID:   id,
		Name: input.Name,
	})
	if err != nil {
		return "", err
	}
	return id, nil
}

func (r *queryResolver) Todos(ctx context.Context) ([]*addmodel.Todo, error) {
	log.Println("[queryResolver.Todos]")
	todos, err := database.NewTodoDao(r.DB).FindAll()
	if err != nil {
		return nil, err
	}
	var results []*addmodel.Todo
	for _, todo := range todos {
		results = append(results, &addmodel.Todo{
			ID:   todo.ID,
			Text: todo.Text,
			Done: todo.Done,
		})
	}
	return results, nil
}

func (r *queryResolver) Todo(ctx context.Context, id string) (*addmodel.Todo, error) {
	log.Printf("[queryResolver.Todo] id: %#v", id)
	todo, err := database.NewTodoDao(r.DB).FindOne(id)
	if err != nil {
		return nil, err
	}
	return &addmodel.Todo{
		ID:   todo.ID,
		Text: todo.Text,
		Done: todo.Done,
	}, nil
}

func (r *queryResolver) Users(ctx context.Context) ([]*addmodel.User, error) {
	log.Println("[queryResolver.Users]")
	users, err := database.NewUserDao(r.DB).FindAll()
	if err != nil {
		return nil, err
	}
	var results []*addmodel.User
	for _, user := range users {
		results = append(results, &addmodel.User{
			ID:   user.ID,
			Name: user.Name,
		})
	}
	return results, nil
}

func (r *queryResolver) User(ctx context.Context, id string) (*addmodel.User, error) {
	log.Printf("[queryResolver.User] id: %#v", id)
	user, err := database.NewUserDao(r.DB).FindOne(id)
	if err != nil {
		return nil, err
	}
	return &addmodel.User{
		ID:   user.ID,
		Name: user.Name,
	}, nil
}

func (r *todoResolver) User(ctx context.Context, obj *addmodel.Todo) (*addmodel.User, error) {
	log.Printf("[queryResolver.User] id: %#v", obj)
	user, err := database.NewUserDao(r.DB).FindByTodoID(obj.ID)
	if err != nil {
		return nil, err
	}
	return &addmodel.User{
		ID:   user.ID,
		Name: user.Name,
	}, nil

}

func (r *userResolver) Todos(ctx context.Context, obj *addmodel.User) ([]*addmodel.Todo, error) {
	log.Printf("[queryResolver.Todos] id: %#v", obj)
	todos, err := database.NewTodoDao(r.DB).FindByUserID(obj.ID)
	if err != nil {
		return nil, err
	}
	var results []*addmodel.Todo
	for _, todo := range todos {
		results = append(results, &addmodel.Todo{
			ID:   todo.ID,
			Text: todo.Text,
			Done: todo.Done,
		})
	}
	return results, nil
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

// Todo returns generated.TodoResolver implementation.
func (r *Resolver) Todo() generated.TodoResolver { return &todoResolver{r} }

// User returns generated.UserResolver implementation.
func (r *Resolver) User() generated.UserResolver { return &userResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
type todoResolver struct{ *Resolver }
type userResolver struct{ *Resolver }
