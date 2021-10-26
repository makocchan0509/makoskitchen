package databases

import "github.com/jinzhu/gorm"

type Menu struct {
	ID    string `gorm:"column:menu_id;primary_key"`
	Name  string `gorm:"column:name"`
	Price int    `gorm:"column:price"`
	Type  string `gorm:"column:type"`
}

func (u *Menu) TableName() string {
	return "menu"
}

type MenuDao interface {
	InsertOne(u *Menu) error
	FindAll() ([]*Menu, error)
	FindOne(id string) (*Menu, error)
	FindByType(name string) ([]*Menu, error)
}

type menuDao struct {
	db *gorm.DB
}

func NewMenuDao(db *gorm.DB) MenuDao {
	return &menuDao{db: db}
}

func (d *menuDao) InsertOne(u *Menu) error {
	res := d.db.Create(u)
	if err := res.Error; err != nil {
		return err
	}
	return nil
}

func (d *menuDao) FindAll() ([]*Menu, error) {
	var menus []*Menu
	res := d.db.Find(&menus)
	if err := res.Error; err != nil {
		return nil, err
	}
	return menus, nil
}

func (d *menuDao) FindOne(id string) (*Menu, error) {
	var menus []*Menu
	res := d.db.Where("menu_id = ?", id).Find(&menus)
	if err := res.Error; err != nil {
		return nil, err
	}
	if len(menus) < 1 {
		return nil, nil
	}
	return menus[0], nil
}

func (d *menuDao) FindByType(type_name string) ([]*Menu, error) {
	var menus []*Menu
	res := d.db.Where("type = ?", type_name).Find(&menus)
	if err := res.Error; err != nil {
		return nil, err
	}
	return menus, nil
}
