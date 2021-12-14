package common

import "net/http"

// 声明一个新的数据类型（函数类型）
type FilterHandler func(w http.ResponseWriter, r *http.Request) error

// 拦截器结构体
type Filter struct {
	// 用来存储需要拦截的URI
	filterMap map[string]FilterHandler
}

func NewFilter() *Filter {
	return &Filter{filterMap: make(map[string]FilterHandler)}
}

// 注册拦截器
func (f *Filter) RegisterFilterUri(uri string, handler FilterHandler) {
	f.filterMap[uri] = handler
}

// 根据uri获取对应的handler
func (f *Filter) GetFileHandler(uri string) FilterHandler {
	return f.filterMap[uri]
}

type WebHandle func(w http.ResponseWriter, r *http.Request)

// 执行拦截器
func (f *Filter) Handle(webHandle WebHandle) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		for path, handle := range f.filterMap {
			if path == r.RequestURI {
				// 执行拦截业务逻辑
				err := handle(w, r)
				if err != nil {
					w.Write([]byte(err.Error()))
					return
				}
				break
			}
		}
		// 执行正常注册的函数
		webHandle(w, r)
	}
}
