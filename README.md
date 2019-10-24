# SEC - sdjdd's expression calculator

## 使用

```go
package main

import (
	"fmt"

	"github.com/sdjdd/sec"
)

func main() {
	calc := sec.New()
	fmt.Println(calc.Eval("(110 + 1919 % 5) * 1000 + 600 - 86"))
}

```

## 功能

### 一元操作

操作符|示例|结果
:-:|:-:|:-:
+|`+1`|1
-|`-1`|-1

### 二元操作

操作符|示例|结果
:-:|:-:|:-:
+|`1+1`|2
-|`2-1`|1
*|`3*4`|12
/|`6/3`|2

### 变量

```go
calc := sec.New()
calc.Env.Vars["yjspi"] = 191981
calc.Eval("yjspi * 10")
```

sec 不允许使用未定义的变量，但您仍有机会在语义分析之前定义表达式使用的变量，这对依靠变量名称执行操作的应用很有帮助。

```go
fmt.Println(calc.Eval("zero")) // 错误："zero" 未定义

calc.BeforeEval = func(env sec.Env, varNames []string) {
    for _, name := range varNames {
        env.Vars[name] = 0
    }
}
fmt.Println(calc.Eval("zero"))
```

### 函数

```go
calc.Env.Funcs["timestamp"] = func() float64 {
    return float64(time.Now().Unix())
}

// 支持可变参数
calc.Env.Funcs["sum"] = func(nums ...float64) (sum float64) {
    for _, n := range nums {
        sum += n
    }
    return
}

fmt.Println(calc.Eval("timestamp()"))
fmt.Println(calc.Eval("sum(1, 2, 3, 4, 5)"))
```

sec 使用的函数有一些限制：

- 必须返回**一个** `float64` 类型的值
- 参数类型全部为 `float64`

sec 预定义了一些 `math` 包中的符合规定的函数

```go
calc.Env.Funcs = sec.MathFuncs
fmt.Println(calc.Eval("pow(2, 10)"))
```
