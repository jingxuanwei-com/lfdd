package service

import (
	"server/core/db"
	"server/model"
)

// ------------------------ GetGoodsSidebar 商品侧边菜单 ------------------------

// SidebarMenu 侧边菜单项
type SidebarMenu struct {
	ID   int64  `json:"id,string"`
	Name string `json:"name"`
}

type GetGoodsSidebarInput struct {
}
type GetGoodsSidebarOutput struct {
	Menus []SidebarMenu `json:"menus"`
}

// GetGoodsSidebar 商品页侧边菜单：全部菜单按 sort 升序（数值越小越靠前）
func GetGoodsSidebar(i GetGoodsSidebarInput) (o GetGoodsSidebarOutput, err error) {
	var menus []model.Menu
	if err = db.DB.Order("sort ASC, id ASC").Find(&menus).Error; err != nil {
		return o, err
	}

	o.Menus = make([]SidebarMenu, 0, len(menus))
	for _, m := range menus {
		o.Menus = append(o.Menus, SidebarMenu{ID: m.ID, Name: m.Name})
	}
	return o, nil
}

// ------------------------ GetGoodsPage 商品页面 ------------------------

// GoodsView 商品视图（价格单位：分，1850 = 18.50 元）
type GoodsView struct {
	ID     int64  `json:"id,string"`
	MenuID int64  `json:"menu_id,string"`
	Name   string `json:"name"`
	Price  int    `json:"price"`
	Spec   string `json:"spec"`
	Stock  int    `json:"stock"`
	Image  string `json:"image"`
	Status int8   `json:"status"`
}

type GetGoodsPageInput struct {
	MenuID   int64 `json:"menu_id,string"` // 菜单（分类）ID，0 = 全部
	Page     int   `json:"page"`           // 页码，从 1 开始
	PageSize int   `json:"page_size"`      // 每页条数
}
type GetGoodsPageOutput struct {
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
	List     []GoodsView `json:"list"`
}

// GetGoodsPage 商品页面：指定菜单下上架商品（status=1）的分页列表
func GetGoodsPage(i GetGoodsPageInput) (o GetGoodsPageOutput, err error) {
	// 分页参数：默认 1 页 20 条，单页上限 100
	if i.Page < 1 {
		i.Page = 1
	}
	if i.PageSize < 1 {
		i.PageSize = 20
	}
	if i.PageSize > 100 {
		i.PageSize = 100
	}

	// 1. 统计总数（只算上架商品）
	countQ := db.DB.Model(&model.Goods{}).Where("status = ?", 1)
	if i.MenuID > 0 {
		countQ = countQ.Where("menu_id = ?", i.MenuID)
	}
	if err = countQ.Count(&o.Total).Error; err != nil {
		return o, err
	}

	// 2. 分页查询（新的查询链，避免复用条件串扰）
	listQ := db.DB.Model(&model.Goods{}).Where("status = ?", 1)
	if i.MenuID > 0 {
		listQ = listQ.Where("menu_id = ?", i.MenuID)
	}
	var goods []model.Goods
	if err = listQ.Order("id DESC").Limit(i.PageSize).Offset((i.Page - 1) * i.PageSize).
		Find(&goods).Error; err != nil {
		return o, err
	}

	o.List = make([]GoodsView, 0, len(goods))
	for _, g := range goods {
		o.List = append(o.List, GoodsView{
			ID: g.ID, MenuID: g.MenuID, Name: g.Name, Price: g.Price,
			Spec: g.Spec, Stock: g.Stock, Image: g.Image, Status: g.Status,
		})
	}
	o.Page = i.Page
	o.PageSize = i.PageSize
	return o, nil
}
