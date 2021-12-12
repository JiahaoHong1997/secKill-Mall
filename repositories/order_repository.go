package repositories

import (
	"database/sql"
	"github.com/pkg/errors"
	"log"
	"seckill/common"
	"seckill/datamodels"
	"strconv"
)

type IOrderRepository interface {
	Conn()
	Insert(*datamodels.Order) (int64, error)
	Delete(int64) (bool, error)
	Update(*datamodels.Order) error
	SelectByKey(int64) (*datamodels.Order, error)
	SelectAll() ([]*datamodels.Order, error)
	SelectAllWithInfo() (map[int]map[string]string, error)
}

type OrderManagerRepository struct {
	table     string
	mysqlConn *sql.DB
}

func NewOrderManagerRepository(table string, sql *sql.DB) IOrderRepository {
	return &OrderManagerRepository{
		table:     table,
		mysqlConn: sql,
	}
}

func (o *OrderManagerRepository) Conn() {
	if o.mysqlConn == nil {
		o.mysqlConn = common.DBConn()
	}
	if o.table == "" {
		o.table = "order"
	}
}

func (o *OrderManagerRepository) Insert(order *datamodels.Order) (int64, error) {
	// 1.判断连接是否存在
	o.Conn()

	// 2.准备sql
	sql := "INSERT `" + o.table + "` SET userID=?,productID=?,orderStatus=?"
	stmt, err := o.mysqlConn.Prepare(sql)
	if err != nil {
		return 0, errors.Wrap(err, "Order#Insert: prepare sql failed")
	}
	defer stmt.Close()

	// 3.传入sql
	result, err := stmt.Exec(order.UserId, order.ProductId, order.OrderStatus)
	if err != nil {
		return 0, errors.Wrap(err, "Order#Insert: insert failed")
	}
	return result.LastInsertId()
}

func (o *OrderManagerRepository) Delete(productID int64) (bool, error) {
	// 1.判断连接是否存在
	o.Conn()

	// 2.准备sql
	sql := "DELETE FROM `" + o.table + "` where ID=?"
	stmt, err := o.mysqlConn.Prepare(sql)
	if err != nil {
		return false, errors.Wrap(err, "Order#Delete: prepare sql failed")
	}
	defer stmt.Close()

	// 3.传入sql
	_, err = stmt.Exec(productID)
	if err != nil {
		return false, errors.Wrap(err, "Order#Delete: delete failed")
	}

	return true, nil
}

func (o *OrderManagerRepository) Update(order *datamodels.Order) error {
	// 1.判断连接是否存在
	o.Conn()

	// 2.准备sql
	sql := "UPDATE `" + o.table + "` SET userID=?,productID=?,orderStatus=? where ID=?"
	stmt, err := o.mysqlConn.Prepare(sql)
	if err != nil {
		return errors.Wrap(err, "Order#Update: prepare sql failed")
	}
	defer stmt.Close()

	// 3.传入sql
	_, err = stmt.Exec(order.UserId, order.ProductId, order.OrderStatus, order.ID)
	if err != nil {
		return errors.Wrap(err, "Order#Update: update failed")
	}
	return nil
}

func (o *OrderManagerRepository) SelectByKey(productID int64) (*datamodels.Order, error) {
	// 1.判断连接是否存在
	o.Conn()

	// 2.查询sql
	sql := "SELECT * FROM `" + o.table + "` WHERE ID=" + strconv.FormatInt(productID, 10)
	row, err := o.mysqlConn.Query(sql)
	if err != nil {
		return &datamodels.Order{}, errors.Wrap(err, "Order#SelectById: query failed")
	}
	defer row.Close()

	// 3.获取首行的查询结果
	result := common.GetResultRow(row)
	if len(result) == 0 {
		log.Println("Order:SelectByKey, no info got\n")
		return &datamodels.Order{}, errors.New("Order#SelectById: not found")
	}

	orderResult := &datamodels.Order{}
	common.DataToStructByTagSql(result, orderResult)
	return orderResult, nil
}

func (o *OrderManagerRepository) SelectAll() ([]*datamodels.Order, error) {
	// 1.判断连接是否正常
	o.Conn()

	// 2.查询sql
	sql := "SELECT * FROM `" + o.table + "`"
	rows, err := o.mysqlConn.Query(sql)
	if err != nil {
		return nil, errors.Wrap(err, "Order#SelectAll: query failed")
	}
	defer rows.Close()

	// 3. 获取所有查询结果
	result := common.GetResultRows(rows)
	if len(result) == 0 {
		return nil, errors.New("Order#SelectAll: not found")
	}

	orderArray := []*datamodels.Order{}
	for _, v := range result {
		order := &datamodels.Order{}
		common.DataToStructByTagSql(v, order)
		orderArray = append(orderArray, order)
	}
	return orderArray, nil
}

func (o *OrderManagerRepository) SelectAllWithInfo() (map[int]map[string]string, error) {
	// 1.判断连接是否正常
	o.Conn()

	// 2.准备sql
	sql := "SELECT o.ID,o.userID,p.productName,o.orderStatus FROM secKill.order as o left join product as p on o.productID=p.ID"
	rows, err := o.mysqlConn.Query(sql)
	if err != nil {
		return nil, errors.Wrap(err, "Order#SelectAllWithInfo: query failed")
	}
	defer rows.Close()

	// 3. 获取所有查询结果
	return common.GetResultRows(rows), nil
}
