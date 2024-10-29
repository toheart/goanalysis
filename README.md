
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

## 示例
![image](./example/default.png)


## 代码跟踪
```shell
Usage:                                 
  goanalysis track <programDir> [flags]
```
示例
输出日志：
```text
time=2024-10-28T23:12:07.398+08:00 level=INFO msg=->example/inner/B.init.0 gid=1 params="[[]], "
time=2024-10-28T23:12:07.415+08:00 level=INFO msg=*<-example/inner/B.init.0 gid=1
time=2024-10-28T23:12:07.416+08:00 level=INFO msg=->example/inner/A.init.0 gid=1 params="[[]], "
time=2024-10-28T23:12:07.416+08:00 level=INFO msg=*->example/inner/B.CalledA gid=1 params="[[]], "
time=2024-10-28T23:12:07.416+08:00 level=INFO msg=**<-example/inner/B.CalledA gid=1
time=2024-10-28T23:12:07.416+08:00 level=INFO msg=*<-example/inner/A.init.0 gid=1
time=2024-10-28T23:12:07.416+08:00 level=INFO msg=->main.main gid=1 params="[[]], "
time=2024-10-28T23:12:07.416+08:00 level=INFO msg=*->example/inner/A.NewCallA gid=1 params="[[tly]], "
time=2024-10-28T23:12:07.416+08:00 level=INFO msg=**<-example/inner/A.NewCallA gid=1
time=2024-10-28T23:12:07.416+08:00 level=INFO msg=*->example/inner/A.CallA.PrintB gid=1 params="[[]], "
time=2024-10-28T23:12:07.416+08:00 level=INFO msg=**->example/inner/B.NewCallB gid=1 params="[[levi]], "
time=2024-10-28T23:12:07.416+08:00 level=INFO msg=***<-example/inner/B.NewCallB gid=1
time=2024-10-28T23:12:07.416+08:00 level=INFO msg=**->example/inner/B.CallB.PrintB gid=1 params="[[]], "
time=2024-10-28T23:12:07.416+08:00 level=INFO msg=***<-example/inner/B.CallB.PrintB gid=1
time=2024-10-28T23:12:07.416+08:00 level=INFO msg=**<-example/inner/A.CallA.PrintB gid=1
time=2024-10-28T23:12:07.416+08:00 level=INFO msg=*->example/inner/A.RecursionA gid=1 params="[[%!s(int=1) %!s(int=10)]], "
time=2024-10-28T23:12:07.416+08:00 level=INFO msg=**->example/inner/A.RecursionA gid=1 params="[[%!s(int=2) %!s(int=10)]], "
time=2024-10-28T23:12:07.416+08:00 level=INFO msg=***->example/inner/A.RecursionA gid=1 params="[[%!s(int=3) %!s(int=10)]], "
time=2024-10-28T23:12:07.416+08:00 level=INFO msg=****->example/inner/A.RecursionA gid=1 params="[[%!s(int=4) %!s(int=10)]], "
time=2024-10-28T23:12:07.416+08:00 level=INFO msg=*****->example/inner/A.RecursionA gid=1 params="[[%!s(int=5) %!s(int=10)]], "
time=2024-10-28T23:12:07.416+08:00 level=INFO msg=******->example/inner/A.RecursionA gid=1 params="[[%!s(int=6) %!s(int=10)]], "
time=2024-10-28T23:12:07.416+08:00 level=INFO msg=*******->example/inner/A.RecursionA gid=1 params="[[%!s(int=7) %!s(int=10)]], "
time=2024-10-28T23:12:07.416+08:00 level=INFO msg=********->example/inner/A.RecursionA gid=1 params="[[%!s(int=8) %!s(int=10)]], "
time=2024-10-28T23:12:07.416+08:00 level=INFO msg=*********->example/inner/A.RecursionA gid=1 params="[[%!s(int=9) %!s(int=10)]], "
time=2024-10-28T23:12:07.417+08:00 level=INFO msg=**********->example/inner/A.RecursionA gid=1 params="[[%!s(int=10) %!s(int=10)]], "
time=2024-10-28T23:12:07.417+08:00 level=INFO msg=***********<-example/inner/A.RecursionA gid=1
time=2024-10-28T23:12:07.417+08:00 level=INFO msg=**********<-example/inner/A.RecursionA gid=1
time=2024-10-28T23:12:07.417+08:00 level=INFO msg=*********<-example/inner/A.RecursionA gid=1
time=2024-10-28T23:12:07.417+08:00 level=INFO msg=********<-example/inner/A.RecursionA gid=1
time=2024-10-28T23:12:07.417+08:00 level=INFO msg=*******<-example/inner/A.RecursionA gid=1
time=2024-10-28T23:12:07.417+08:00 level=INFO msg=******<-example/inner/A.RecursionA gid=1
time=2024-10-28T23:12:07.417+08:00 level=INFO msg=*****<-example/inner/A.RecursionA gid=1
time=2024-10-28T23:12:07.417+08:00 level=INFO msg=****<-example/inner/A.RecursionA gid=1
time=2024-10-28T23:12:07.417+08:00 level=INFO msg=***<-example/inner/A.RecursionA gid=1
time=2024-10-28T23:12:07.417+08:00 level=INFO msg=**<-example/inner/A.RecursionA gid=1
time=2024-10-28T23:12:07.417+08:00 level=INFO msg=*<-main.main gid=1
```