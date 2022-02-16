package dao

import (
	"context"
	"database/sql"
	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
	"log"
	"seckill/common"
	"seckill/dao/db"
	"seckill/models"
	"strconv"
	"time"
)

type IProduct interface {
	Conn()
	Insert(*models.Product) (int64, error)
	Delete(int64) (bool, error)
	Update(*models.Product) error
	SelectByKey(int64) (*models.Product, error)
	SelectAll() ([]*models.Product, error)
	SubProductNum(productID int64) error
	AddSecProduct(productID int64, productNum int64, duration float64) error
}

type ProductManager struct {
	table     string
	mysqlConn *sql.DB
	redisPool *redis.Client
}

func NewProductManager(table string, db *sql.DB, rdb *redis.Client) IProduct {
	return &ProductManager{
		table:     table,
		mysqlConn: db,
		redisPool: rdb,
	}
}

// 数据库连接
func (p *ProductManager) Conn() {
	if p.mysqlConn == nil {
		p.mysqlConn = db.DBConn()
	}
	if p.table == "" {
		p.table = "product"
	}
}

// redis连接
func (p *ProductManager) RedisConn() {
	if p.redisPool == nil {
		p.redisPool = db.NewRedisConn()
	}
}

// 插入
func (p *ProductManager) Insert(product *models.Product) (productId int64, err error) {
	// 1.判断连接是否存在
	p.Conn()

	// 2.准备sql
	sql := "INSERT product SET productName=?, productNum=?, productImage=?, productUrl=?"
	stmt, err := p.mysqlConn.Prepare(sql)
	if err != nil {
		return 0, errors.Wrap(err, "product_repository#Insert: sql prepare failed")
	}
	defer stmt.Close()
	// 3.传入sql
	result, err := stmt.Exec(product.ProductName, product.ProductNum, product.ProductImage, product.ProductUrl)
	if err != nil {
		return 0, errors.Wrap(err, "product_repository#Insert: insert failed")
	}
	return result.LastInsertId() // 获得最后插入的id
}

// 删除
func (p *ProductManager) Delete(productId int64) (bool, error) {
	// 1.判断连接是否存在
	p.Conn()

	// 2.准备sql
	sql := "DELETE FROM product WHERE ID=?"
	stmt, err := p.mysqlConn.Prepare(sql)
	if err != nil {
		return false, errors.Wrap(err, "product_repository#Delete: sql prepare failed")
	}
	defer stmt.Close()
	// 3.传入sql
	_, err = stmt.Exec(productId)
	if err != nil {
		return false, errors.Wrap(err, "product_repository#Delete: delete failed")
	}
	return true, nil
}

// 修改
func (p *ProductManager) Update(product *models.Product) error {
	// 1.判断连接是否存在
	p.Conn()

	// 2.准备sql
	sql := "UPDATE product SET productName=?, productNum=?, productImage=?, productUrl=? where ID=" + strconv.FormatInt(product.ID, 10)
	stmt, err := p.mysqlConn.Prepare(sql)
	if err != nil {
		return errors.Wrap(err, "product_repository#Update: sql prepare failed")
	}
	defer stmt.Close()
	// 3.传入sql
	_, err = stmt.Exec(product.ProductName, product.ProductNum, product.ProductImage, product.ProductUrl)
	if err != nil {
		return errors.Wrap(err, "product_repository#Update: update failed")
	}
	return nil
}

// 查询
func (p *ProductManager) SelectByKey(productID int64) (*models.Product, error) {
	// 1.判断连接是否存在
	p.Conn()

	// 2.查询sql
	sql := "SELECT * FROM " + p.table + " WHERE ID=" + strconv.FormatInt(productID, 10)
	row, err := p.mysqlConn.Query(sql)
	if err != nil {
		return &models.Product{}, errors.Wrap(err, "product_repository#SelectByKey: query failed")
	}
	defer row.Close()
	// 获取查询结果的首行
	result := db.GetResultRow(row)
	if len(result) == 0 {
		return &models.Product{}, errors.New("product_repository#SelectByKey: not found")
	}

	productResult := &models.Product{}
	common.DataToStructByTagSql(result, productResult)
	return productResult, nil
}

//获取所有商品
func (p *ProductManager) SelectAll() (productArray []*models.Product, errProduct error) {
	//1.判断连接是否存在
	p.Conn()

	sql := "Select * from " + p.table
	rows, err := p.mysqlConn.Query(sql)
	if err != nil {
		return nil, errors.Wrap(err, "product_repository#SelectAll: query failed")
	}
	defer rows.Close()

	result := db.GetResultRows(rows)
	if len(result) == 0 {
		return nil, errors.New("product_repository#SelectAll: not found")
	}

	for _, v := range result {
		product := &models.Product{}
		common.DataToStructByTagSql(v, product)
		productArray = append(productArray, product)
	}
	return
}

// 库存减一
func (p *ProductManager) SubProductNum(productID int64) error {
	p.Conn()
	sql := "update " + p.table + " set " + " productNum=productNum-1 where ID =" + strconv.FormatInt(productID, 10)
	stmt, err := p.mysqlConn.Prepare(sql)

	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec()
	return err
}

func (p *ProductManager) AddSecProduct(productID int64, productNum int64, duration float64) error {
	p.RedisConn()

	ctx := context.Background()
	idString := strconv.FormatInt(productID, 10)
	numString := strconv.FormatInt(productNum, 10)
	countDown := int(duration*3600)
	_, err := p.redisPool.Set(ctx, idString, numString, 0).Result()
	if err != nil {
		return errors.Wrap(err,"product_repository#AddSecProduct: set secKill product failed.")
	}

	var t time.Duration
	t = time.Duration(countDown)
	ok, err := p.redisPool.ExpireAt(ctx, idString, time.Now().Add(t*time.Second)).Result()
	if !ok {
		log.Println("failed set redis")
	}
	return nil
}