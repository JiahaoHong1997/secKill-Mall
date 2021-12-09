package repositories

import (
	"database/sql"
	"seckill/common"
	"seckill/datamodels"
	"strconv"
)

type IProduct interface {
	Conn() error
	Insert(*datamodels.Product) (int64, error)
	Delete(int64) bool
	Update(*datamodels.Product) error
	SelectByKey(int64) (*datamodels.Product, error)
	SelectAll() ([]*datamodels.Product, error)
	SubProductNum(productID int64) error
}

type ProductManager struct {
	table     string
	mysqlConn *sql.DB
}

func NewProductManager(table string, db *sql.DB) IProduct {
	return &ProductManager{
		table:     table,
		mysqlConn: db,
	}
}

// 数据库连接
func (p *ProductManager) Conn() error {
	if p.mysqlConn == nil {
		mysql := common.DBConn()
		p.mysqlConn = mysql
	}
	if p.table == "" {
		p.table = "product"
	}
	return nil
}

// 插入
func (p *ProductManager) Insert(product *datamodels.Product) (productId int64, err error) {
	// 1.判断连接是否存在
	if err := p.Conn(); err != nil {
		return 0, err
	}
	// 2.准备sql
	sql := "INSERT product SET productName=?, productNum=?, productImage=?, productUrl=?"
	stmt, err := p.mysqlConn.Prepare(sql)
	defer stmt.Close()
	if err != nil {
		return 0, err
	}
	// 3.传入sql
	result, err := stmt.Exec(product.ProductName, product.ProductNum, product.ProductImage, product.ProductUrl)
	if err != nil {
		return 0, nil
	}
	return result.LastInsertId() // 获得最后插入的id
}

// 删除
func (p *ProductManager) Delete(productId int64) bool {
	// 1.判断连接是否存在
	if err := p.Conn(); err != nil {
		return false
	}
	// 2.准备sql
	sql := "DELETE FROM product WHERE ID=?"
	stmt, err := p.mysqlConn.Prepare(sql)
	defer stmt.Close()
	if err != nil {
		return false
	}
	// 3.传入sql
	_, err = stmt.Exec(productId)
	if err != nil {
		return false
	}
	return true
}

// 修改
func (p *ProductManager) Update(product *datamodels.Product) error {
	// 1.判断连接是否存在
	if err := p.Conn(); err != nil {
		return err
	}
	// 2.准备sql
	sql := "UPDATE product SET productName=?, productNum=?, productImage=?, productUrl=? where ID=" + strconv.FormatInt(product.ID, 10)
	stmt, err := p.mysqlConn.Prepare(sql)
	defer stmt.Close()
	if err != nil {
		return err
	}
	// 3.传入sql
	_, err = stmt.Exec(product.ProductName, product.ProductNum, product.ProductImage, product.ProductUrl)
	if err != nil {
		return err
	}
	return nil
}

// 查询
func (p *ProductManager) SelectByKey(productID int64) (*datamodels.Product, error) {
	// 1.判断连接是否存在
	if err := p.Conn(); err != nil {
		return &datamodels.Product{}, err
	}
	// 2.查询sql
	sql := "SELECT * FROM " + p.table + " WHERE ID=" + strconv.FormatInt(productID, 10)
	row, err := p.mysqlConn.Query(sql)
	defer row.Close()
	if err != nil {
		return &datamodels.Product{}, err
	}
	// 获取查询结果的首行
	result := common.GetResultRow(row)
	if len(result) == 0 {
		return &datamodels.Product{}, nil
	}

	productResult := &datamodels.Product{}
	common.DataToStructByTagSql(result, productResult)
	return productResult, nil
}

//获取所有商品
func (p *ProductManager) SelectAll() (productArray []*datamodels.Product, errProduct error) {
	//1.判断连接是否存在
	if err := p.Conn(); err != nil {
		return nil, err
	}
	sql := "Select * from " + p.table
	rows, err := p.mysqlConn.Query(sql)
	defer rows.Close()
	if err != nil {
		return nil, err
	}

	result := common.GetResultRows(rows)
	if len(result) == 0 {
		return nil, nil
	}

	for _, v := range result {
		product := &datamodels.Product{}
		common.DataToStructByTagSql(v, product)
		productArray = append(productArray, product)
	}
	return
}

// 库存减一
func (p *ProductManager) SubProductNum(productID int64) error {
	if err := p.Conn(); err != nil {
		return err
	}
	sql := "update " + p.table + " set " + " productNum=productNum-1 where ID =" + strconv.FormatInt(productID, 10)
	stmt, err := p.mysqlConn.Prepare(sql)
	defer stmt.Close()
	if err != nil {
		return err
	}
	_, err = stmt.Exec()
	return err
}
