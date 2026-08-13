---
title: "Shortcode 组件演示"
date: 2026-08-13
categories: ["测试"]
tags: ["shortcode"]
draft: true
description: "details / gallery / tabs 组件集成验证页"
---

## Details 折叠

{{< details "答案" >}}
这里是被折叠的**重点内容**，点击标题展开。
{{< /details >}}

## Gallery 图廊

{{< gallery >}}
![示例一](/img/demo-a.png)
![示例二](/img/demo-b.png "第二张")
{{< /gallery >}}

## Tabs 标签页

{{< tabs "Go" "Python" >}}
=== Go ===
```go
package main

func main() {
    fmt.Println("hello")
}
```
=== Python ===
```python
def main():
    print("hello")
```
{{< /tabs >}}

结尾段落。
