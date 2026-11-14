package service

import (
	"errors"
	"fmt"
	"math/rand"
	"server/core/db"
	"server/model"
	"time"

	"gorm.io/gorm"
)

// ------------------------ CreateOrder 创建订单 ------------------------

type CreateOrderItemInput struct {
	ID  int64 `json:"id,string"` // 商品 ID
	Qty int   `json:"qty"`       // 购买数量
}
type CreateOrderInput struct {
	Items []CreateOrderItemInput `json:"items"` // 购物车明细
}
type CreateOrderOutput struct {
	ID      int64  `json:"id,string"`
	OrderNo string `json:"order_no"` // 订单号（打印小票用）
	Total   int    `json:"total"`    // 合计，单位：分
}

// CreateOrder 下单：价格一律按库中当前价结算（不信任前端传价），
// 校验上架/库存，并在事务中扣减库存、写订单与明细快照
func CreateOrder(i CreateOrderInput) (o CreateOrderOutput, err error) {
	if len(i.Items) == 0 {
		return o, errors.New("购物车是空的")
	}

	// 1. 合并同商品数量并校验基本值
	qtyByID := map[int64]int{}
	ids := []int64{}
	for _, it := range i.Items {
		if it.ID <= 0 || it.Qty <= 0 {
			return o, errors.New("购物车数据异常")
		}
		if _, seen := qtyByID[it.ID]; !seen {
			ids = append(ids, it.ID)
		}
		qtyByID[it.ID] += it.Qty
		if qtyByID[it.ID] > matchMaxQty {
			return o, fmt.Errorf("单商品数量不能超过 %d", matchMaxQty)
		}
	}

	// 2. 取上架商品（价格以库中为准）
	var goods []model.Goods
	if err = db.DB.Where("id IN ? AND status = ?", ids, 1).Find(&goods).Error; err != nil {
		return o, err
	}
	byID := make(map[int64]model.Goods, len(goods))
	for _, g := range goods {
		byID[g.ID] = g
	}

	total, count := 0, 0
	for _, id := range ids {
		g, ok := byID[id]
		if !ok {
			return o, fmt.Errorf("商品已下架或不存在，请刷新购物车")
		}
		qty := qtyByID[id]
		if g.Stock > 0 && qty > g.Stock {
			return o, fmt.Errorf("「%s」库存不足（剩 %d）", g.Name, g.Stock)
		}
		total += g.Price * qty
		count += qty
	}

	// 3. 事务：扣库存 → 建订单 → 写明细快照
	o.Total = total
	o.OrderNo = time.Now().Format("20060102150405") + fmt.Sprintf("%06d", rand.Intn(1e6))
	err = db.DB.Transaction(func(tx *gorm.DB) error {
		for _, id := range ids {
			g, qty := byID[id], qtyByID[id]
			if g.Stock <= 0 {
				continue // -1 不限库存
			}
			res := tx.Model(&model.Goods{}).
				Where("id = ? AND stock >= ?", id, qty).
				Update("stock", gorm.Expr("stock - ?", qty))
			if res.Error != nil {
				return res.Error
			}
			if res.RowsAffected == 0 {
				return fmt.Errorf("「%s」库存不足", g.Name)
			}
		}

		ord := model.Order{OrderNo: o.OrderNo, Total: total, Count: count}
		if e := tx.Create(&ord).Error; e != nil {
			return e
		}
		items := make([]model.OrderItem, 0, len(ids))
		for _, id := range ids {
			g, qty := byID[id], qtyByID[id]
			items = append(items, model.OrderItem{
				OrderID: ord.ID, GoodsID: g.ID, Name: g.Name,
				Price: g.Price, Spec: g.Spec, Qty: qty, Image: g.Image,
			})
		}
		if e := tx.Create(&items).Error; e != nil {
			return e
		}
		o.ID = ord.ID
		return nil
	})
	if err != nil {
		return o, err
	}
	return o, nil
}

// ------------------------ GetOrders 订单分页 ------------------------

type GetOrdersInput struct {
	Page     int `json:"page"`      // 页码，从 1 开始
	PageSize int `json:"page_size"` // 每页条数
}
type OrderItemView struct {
	GoodsID int64  `json:"goods_id,string"`
	Name    string `json:"name"`
	Price   int    `json:"price"` // 单价，分
	Spec    string `json:"spec"`
	Qty     int    `json:"qty"`
	Image   string `json:"image"`
}
type OrderView struct {
	ID        int64           `json:"id,string"`
	OrderNo   string          `json:"order_no"`
	Total     int             `json:"total"` // 合计，分
	Count     int             `json:"count"` // 商品总件数
	CreatedAt string          `json:"created_at"`
	Items     []OrderItemView `json:"items"`
}
type GetOrdersOutput struct {
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
	List     []OrderView `json:"list"`
}

// GetOrders 订单分页：最新在前，附带每单明细（供列表展示与小票打印）
func GetOrders(i GetOrdersInput) (o GetOrdersOutput, err error) {
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

	if err = db.DB.Model(&model.Order{}).Count(&o.Total).Error; err != nil {
		return o, err
	}
	var orders []model.Order
	if err = db.DB.Order("id DESC").Limit(i.PageSize).Offset((i.Page - 1) * i.PageSize).
		Find(&orders).Error; err != nil {
		return o, err
	}
	o.Page = i.Page
	o.PageSize = i.PageSize
	o.List = make([]OrderView, 0, len(orders))
	if len(orders) == 0 {
		return o, nil
	}

	// 批量取明细，按订单分组
	orderIDs := make([]int64, 0, len(orders))
	for _, ord := range orders {
		orderIDs = append(orderIDs, ord.ID)
	}
	var items []model.OrderItem
	if err = db.DB.Where("order_id IN ?", orderIDs).Order("id ASC").Find(&items).Error; err != nil {
		return o, err
	}
	itemsByOrder := map[int64][]OrderItemView{}
	for _, it := range items {
		itemsByOrder[it.OrderID] = append(itemsByOrder[it.OrderID], OrderItemView{
			GoodsID: it.GoodsID, Name: it.Name, Price: it.Price,
			Spec: it.Spec, Qty: it.Qty, Image: it.Image,
		})
	}

	for _, ord := range orders {
		list := itemsByOrder[ord.ID]
		if list == nil {
			list = []OrderItemView{} // 保证输出 [] 而非 null
		}
		o.List = append(o.List, OrderView{
			ID:        ord.ID,
			OrderNo:   ord.OrderNo,
			Total:     ord.Total,
			Count:     ord.Count,
			CreatedAt: ord.CreatedAt.Format("2006-01-02 15:04:05"),
			Items:     list,
		})
	}
	return o, nil
}
