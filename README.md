# spikeMall
 Online shopping mall with high concurrency spike system



## 需求分析

### 主要功能

* 用户登录
* 商品展示
* 商品抢购
* 商品后台管理



### 系统需求分析

* 前端页面需要承载大流量
* 在大并发状态下要解决超卖问题
* 后端接口需要满足横向扩展



## 商品后台管理功能

* 目录结构

```
+-- common			# 公共方法
|	|-- common.go	# 类型转换方法
|	|-- form.go		# html请求解析方法
|	|-- mysql.go	# 初始化mysql请求池
+-- datamodels		# 所有模型的存放目录
|	|-- product.go	# product结构
+-- repositories	# 所有数据库操作结构体存目录
|	|-- product_repository.go	# 与商品后台管理相关的数据库操作
+-- service			# 所有逻辑业务代码存放目录
```

 
