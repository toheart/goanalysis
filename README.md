# goanalysis

根据不同的项目, 可选择不同的分析算法。 具体多尝试。

## 使用
```shell
Usage:
  goanalysis analysis [options] <programDir> [flags]

Flags:
  -a, --algo string         The algorithm used to construct the call graph. Possible values inlcude: "vta", "static", "cha", "rta", default: "vta" (default "vta")
      --cacheFlag           Whether to enable caching, default true (default true)
  -c, --cachePath string    FuncNode cache output path,default: ./cache.json (default "./cachePath.json")
  -h, --help                help for analysis
  -i, --ignore string       Ignore methods paths containing given suffix
  -p, --onlyMethod string   Only output relevant package names and method names
  -o, --outputPath string   Image output path,default: ./default.png (default "./default.png")

``` 
> 注：onlyMethod使用的是callGraph中内部唯一Key, 可以通过查看缓存文件索引来获取。 <br>
> ignore的使用, 可以使用, 为分隔, 使用的是前缀匹配。

## 测试方法
```shell
goanalysis analysis ./example
```


