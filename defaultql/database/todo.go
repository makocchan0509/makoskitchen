package database

import "github.com/jinzhu/gorm"

type Todo struct {
	ID     string `gorm:"column:id;primary_key"`
	Text   string `gorm:"column:text"`
	Done   bool   `gorm:"column:done"`
	UserID string `gorm:"column:user_id"`
}

func (u *Todo) TableName() string {
	return "todo"
}

type TodoDao interface {
	InsertOne(u *Todo) error
	FindAll() ([]*Todo, error)
	FindOne(id string) (*Todo, error)
	FindByUserID(user_id string) ([]*Todo, error)
}

type todoDao struct {
	db *gorm.DB
}

func NewTodoDao(db *gorm.DB) TodoDao {
	return &todoDao{db: db}
}

func (d *todoDao) InsertOne(u *Todo) error {
	res := d.db.Create(u)
	if err := res.Error; err != nil {
		return err
	}
	return nil
}
func (d *todoDao) FindAll() ([]*Todo, error) {
	var todos []*Todo
	res := d.db.Find(&todos)
	if err := res.Error; err != nil {
		return nil, err
	}
	return todos, nil
}

func (d *todoDao) FindOne(id string) (*Todo, error) {
	var todos []*Todo
	res := d.db.Where("id = ?", id).Find(todos)
	if err := res.Error; err != nil {
		return nil, err
	}
	if len(todos) < 1 {
		return nil, nil
	}
	return todos[0], nil
}
func (d *todoDao) FindByUserID(user_id string) ([]*Todo, error) {
	var todos []*Todo
	res := d.db.Where("user_id = ?", user_id).Find(&todos)
	if err := res.Error; err != nil {
		return nil, err
	}
	return todos, nil

}
