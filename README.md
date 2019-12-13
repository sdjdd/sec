# SEC - sdjdd's expression calculator

**不止是玩具！**

- 支持函数、变量
- 丰富的错误类型和友好的错误提示
- 充分的测试(WIP)
- 可作为词法分析的入门参考

## 开始使用

```go
package main

import (
	"fmt"
	"github.com/sdjdd/sec"
)

func main() {
	fmt.Println(sec.Eval("1+1"))
}

```

使用独立的解析器

```go
var psr sec.Parser
expr, _ := psr.Parse("1+1")
fmt.Println(expr.Val(sec.DefaultEnv))
```

### 一元运算符

- 取正: `+`
- 取负: `-`

### 二元运算符

- 加: `+`
- 减: `-`
- 乘以: `*`
- 除以: `/`
- 取余: `%`
- 求幂: `**`
- 除以并取整: `//`（sec 中的值均为 `float64` 类型）

### 使用变量

> 虽然无法在表达式中更改变量的值（sec 不支持自增和自减运算符），但定义变量的宿主程序可以随意修改变量的值，所以 sec 依然将其称为“变量”。

```go
sec.DefaultEnv.Vars["yjspi"] = 114514
val, _ := sec.Eval("yjspi")
fmt.Println(val) // output: 114514
```

### 使用函数

sec 中的函数：

- 必须返回且仅返回一个 `float64` 类型的值
- 不含参数或参数类型全部为 `float64`

```go
sec.DefaultEnv.Funcs["timestamp"] = func() float64 {
    return float64(time.Now().Unix())
}
val, _ := sec.Eval("timestamp()")
fmt.Println(val) // output current timestamp

// 支持可变参数
sec.DefaultEnv.Funcs["sum"] = func(nums ...float64) (sum float64) {
    for _, n := range nums {
        sum += n
    }
    return
}
val, _ = sec.Eval("sum(1, 2, 3, 4, 5)")
fmt.Println(val) // output: 15
```
