package repositories

import (
	"database/sql"
	"log"
	"seckill/common"
	"seckill/datamodels"
	"strconv"
)

type IOrderRepository interface {
	Conn() error
	Insert(*datamodels.Order) (int64, error)
	Delete(int64) bool
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

func (o *OrderManagerRepository) Conn() error {
	if o.mysqlConn == nil {
		o.mysqlConn = common.DBConn()
	}
	if o.table == "" {
		o.table = "order"
	}
	return nil
}

func (o *OrderManagerRepository) Insert(order *datamodels.Order) (int64, error) {
	// 1.判断连接是否存在
	if err := o.Conn(); err != nil {
		log.Printf("Order:Insert, failed to connect to mysql: %v\n", err)
		return 0, nil
	}

	// 2.准备sql
	sql := "INSERT `" + o.table + "` SET userID=?,productID=?,orderStatus=?"
	stmt, err := o.mysqlConn.Prepare(sql)
	defer stmt.Close()
	if err != nil {
		log.Printf("Order:Insert, failed to prepare for mysql: %v\n", err)
		return 0, err
	}

	// 3.传入sql
	result, err := stmt.Exec(order.UserId, order.ProductId, order.OrderStatus)
	if err != nil {
		log.Printf("Order:Insert, failed to exec the insert opration: %v\n", err)
		return 0, err
	}
	return result.LastInsertId()
}

func (o *OrderManagerRepository) Delete(productID int64) bool {
	// 1.判断连接是否存在
	if err := o.Conn(); err != nil {
		log.Printf("Order:Delete, failed to connect to mysql: %v\n", err)
		return false
	}

	// 2.准备sql
	sql := "DELETE FROM `" + o.table + "` where ID=?"
	stmt, err := o.mysqlConn.Prepare(sql)
	defer stmt.Close()
	if err != nil {
		log.Printf("Order:Delete, failed to prepare for mysql: %v\n", err)
		return false
	}

	// 3.传入sql
	_, err = stmt.Exec(productID)
	if err != nil {
		log.Printf("Order:Delete, failed to exec the delete opration: %v\n", err)
		return false
	}

	return true
}

func (o *OrderManagerRepository) Update(order *datamodels.Order) error {
	// 1.判断连接是否存在
	if err := o.Conn(); err != nil {
		log.Printf("Order:Update, failed to connect to mysql: %v\n", err)
		return err
	}

	// 2.准备sql
	sql := "UPDATE `" + o.table + "` SET userID=?,productID=?,orderStatus=? where ID=?"
	log.Println(sql)
	stmt, err := o.mysqlConn.Prepare(sql)
	defer stmt.Close()
	if err != nil {
		log.Printf("Order:Update, failed to prepare for mysql: %v\n", err)
		return err
	}

	// 3.传入sql
	_, err = stmt.Exec(order.UserId, order.ProductId, order.OrderStatus, order.ID)
	if err != nil {
		log.Printf("Order:Update, failed to exec the update opration: %v\n", err)
		return err
	}
	return nil
}

func (o *OrderManagerRepository) SelectByKey(productID int64) (*datamodels.Order, error) {
	// 1.判断连接是否存在
	if err := o.Conn(); err != nil {
		log.Printf("Order:SelectByKey, failed to connect to mysql: %v\n", err)
		return &datamodels.Order{}, err
	}

	// 2.查询sql
	sql := "SELECT * FROM `" + o.table + "` WHERE ID=" + strconv.FormatInt(productID, 10)
	log.Println(sql)
	row, err := o.mysqlConn.Query(sql)
	defer row.Close()
	if err != nil {
		log.Printf("Order:SelectByKey, failed to query information: %v\n", err)
		return &datamodels.Order{}, err
	}

	// 3.获取首行的查询结果
	result := common.GetResultRow(row)
	if len(result) == 0 {
		log.Println("Order:SelectByKey, no info got\n")
		return &datamodels.Order{}, nil
	}

	orderResult := &datamodels.Order{}
	common.DataToStructByTagSql(result, orderResult)
	return orderResult, nil
}

func (o *OrderManagerRepository) SelectAll() ([]*datamodels.Order, error) {
	// 1.判断连接是否正常
	if err := o.Conn(); err != nil {
		log.Printf("Order:SelectAll, failed to connect to mysql: %v\n", err)
		return nil, err
	}

	// 2.查询sql
	sql := "SELECT * FROM `" + o.table + "`"
	rows, err := o.mysqlConn.Query(sql)
	defer rows.Close()
	if err != nil {
		log.Printf("Order:SelectAll: failed to query information: %v\n", err)
		return nil, err
	}

	// 3. 获取所有查询结果
	result := common.GetResultRows(rows)
	if len(result) == 0 {
		log.Println("Order:SelectAll: no information got\n")
		return nil, nil
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
	if err := o.Conn(); err != nil {
		log.Printf("Order:SelectAllWithInfo, failed to connect to mysql: %v\n", err)
		return nil, err
	}

	// 2.准备sql
	sql := "SELECT o.ID,o.userID,p.productName,o.orderStatus FROM secKill.order as o left join product as p on o.productID=p.ID"
	rows, err := o.mysqlConn.Query(sql)
	defer rows.Close()
	if err != nil {
		log.Printf("Order:SelectAllWithInfo: failed to query information: %v\n", err)
		return nil, err
	}

	// 3. 获取所有查询结果
	return common.GetResultRows(rows), nil
}
