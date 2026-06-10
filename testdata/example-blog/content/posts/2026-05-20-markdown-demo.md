---
title: "Markdown 功能演示"
date: 2026-05-20
categories: ["技术"]
tags: ["Markdown", "演示"]
series: "镇海开发日志"
toc: true
---

## 表格演示

| 特性 | 状态 | 说明 |
|------|------|------|
| GFM 表格 | ✅ | 已支持 |
| 代码高亮 | ✅ | Chroma 引擎 |
| 数学公式 | ✅ | KaTeX 渲染 |

## 代码块

```python
def fibonacci(n):
    """计算斐波那契数列"""
    if n <= 1:
        return n
    return fibonacci(n-1) + fibonacci(n-2)

print(fibonacci(10))  # 输出: 55
```

## 提示框

> [!note]
> 这是一个普通的提示框。

> [!warning]
> 这是一个警告提示框，需要注意！

> [!tip]
> 小技巧：使用 `Ctrl+K` 打开搜索。

## 图片

![镇海](https://example.com/zhenhai.webp "我的秘书舰")

一段普通的正文内容，包含了**粗体**、*斜体*和`行内代码`。
