package dao

import (
	"github.com/jinzhu/gorm"
	"{{.Mod}}/model"
)

type {{.Table.GoName}}Dao struct{ db *gorm.DB }

func New{{.Table.GoName}}Dao(db *gorm.DB) *{{.Table.GoName}}Dao { return &{{.Table.GoName}}Dao{db: db} }

func (d *{{.Table.GoName}}Dao) Create(m *model.{{.Table.GoName}}) error { return d.db.Create(m).Error }
func (d *{{.Table.GoName}}Dao) Get(id uint) (*model.{{.Table.GoName}}, error) {
	var m model.{{.Table.GoName}}
	err := d.db.First(&m, id).Error
	return &m, err
}
func (d *{{.Table.GoName}}Dao) Update(m *model.{{.Table.GoName}}) error { return d.db.Save(m).Error }
func (d *{{.Table.GoName}}Dao) Delete(id uint) error { return d.db.Delete(&model.{{.Table.GoName}}{}, id).Error }
func (d *{{.Table.GoName}}Dao) List() ([]model.{{.Table.GoName}}, error) {
	var list []model.{{.Table.GoName}}
	err := d.db.Find(&list).Error
	return list, err
}