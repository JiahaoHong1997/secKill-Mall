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
	"strings"
	"time"
)

type IProduct interface {
	Conn()
	Insert(*models.Product) (int64, error)
	Delete(int64) (bool, error)
	Update(*models.Product) (int64, error)
	SelectByKey(int64) (*models.Product, error)
	SelectAll() ([]*models.Product, error)
	SubProductNum(productID int64) error
	AddSecProduct(productID int64, productNum int64, duration float64) error
	InsertCache(product *models.MessageCache) (int64, error)
	DeleteCache(productId int64) (bool, error)
	SelectByIdCache(productId int64) (*models.Product, error)
}

type ProductManager struct {
	table     string
	mysqlConn *sql.DB
	redisPool *redis.Client
	cachePool *redis.Client
}

func NewProductManager(table string, db *sql.DB, rdb *redis.Client, cache *redis.Client) IProduct {
	return &ProductManager{
		table:     table,
		mysqlConn: db,
		redisPool: rdb,
		cachePool: cache,
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

func (p *ProductManager) CacheConn() {
	if p.cachePool == nil {
		p.cachePool = db.NewCachePool()
	}
}

// 数据库
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
func (p *ProductManager) Update(product *models.Product) (int64, error) {
	// 1.判断连接是否存在
	p.Conn()

	// 2.准备sql
	sql := "UPDATE product SET productName=?, productNum=?, productImage=?, productUrl=? where ID=" + strconv.FormatInt(product.ID, 10)
	stmt, err := p.mysqlConn.Prepare(sql)
	if err != nil {
		return -1, errors.Wrap(err, "product_repository#Update: sql prepare failed")
	}
	defer stmt.Close()
	// 3.传入sql
	_, err = stmt.Exec(product.ProductName, product.ProductNum, product.ProductImage, product.ProductUrl)
	if err != nil {
		return -1, errors.Wrap(err, "product_repository#Update: update failed")
	}
	return product.ID, nil
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

// 热点商品
func (p *ProductManager) AddSecProduct(productID int64, productNum int64, duration float64) error {
	p.RedisConn()

	ctx := context.Background()
	idString := strconv.FormatInt(productID, 10)
	s := []string{idString, "inventory"}
	key := strings.Join(s, "#")
	numString := strconv.FormatInt(productNum, 10)

	_, err := p.redisPool.Set(ctx, key, numString, 0).Result()
	if err != nil {
		return errors.Wrap(err, "product_repository#AddSecProduct: set secKill product failed.")
	}

	var t time.Duration
	countDown := int(duration * 3600)
	t = time.Duration(countDown)
	ok, err := p.redisPool.ExpireAt(ctx, key, time.Now().Add(t*time.Second)).Result()
	if !ok {
		log.Println("failed set redis")
	}
	return nil
}

// 缓存
func (p *ProductManager) InsertCache(product *models.MessageCache) (int64, error) {
	p.CacheConn()

	ctx := context.Background()
	id := product.ID
	idString := strconv.FormatInt(id, 10)
	name := product.ProductName
	num := product.ProductNum
	numString := strconv.FormatInt(num, 10)
	image := product.ProductImage
	url := product.ProductUrl
	productInfo := []string{"productName", name, "productInventory", numString, "productImage", image, "productUrl", url}

	successNum, err := p.cachePool.HSet(ctx, idString, productInfo).Result()
	if err != nil || successNum != 4 {
		return -1, errors.Wrap(err, "product_repository#InsertCache: insert cache failed")
	}
	return successNum, nil
}

func (p *ProductManager) DeleteCache(productId int64) (bool, error) {
	p.CacheConn()

	ctx := context.Background()
	idString := strconv.FormatInt(productId, 10)

	successNum, err := p.cachePool.HDel(ctx, idString, "productName", "productInventory", "productImage", "productUrl").Result()
	if err != nil {
		return false, errors.Wrap(err, "product_repository#DeleteCache: delete cache failed")
	}
	if successNum == 0 {
		return true, errors.New("Nothing Deleted")
	}

	return true, nil
}

func (p *ProductManager) SelectByIdCache(productId int64) (*models.Product, error) {
	p.CacheConn()

	ctx := context.Background()
	idString := strconv.FormatInt(productId, 10)
	productInfo, err := p.cachePool.HGetAll(ctx, idString).Result()
	if err != nil {
		return nil, errors.Wrap(err, "product_repository#SelectByIdCache: select failed")
	}
	if len(productInfo) == 0 {
		return nil, errors.New("Cache Missed")
	}
	productResult := &models.Product{}
	for k, v := range productInfo {
		switch k {
		case "productName":
			productResult.ProductName = v
		case "productInventory":
			x, _ := strconv.ParseInt(v, 10, 64)
			productResult.ProductNum = x
		case "productImage":
			productResult.ProductImage = v
		case "productUrl":
			productResult.ProductUrl = v
		}
	}
	productResult.ID = productId
	return productResult, nil
}
