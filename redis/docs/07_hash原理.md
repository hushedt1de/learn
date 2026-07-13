# Redis Hash 实现原理

## Redis Hash 的逻辑类型与内部编码

Hash 是 Redis 暴露给用户的逻辑类型，用来保存一组 field-value 映射。每个 field 和 value 本质上都是二进制安全的字符串。

在 Redis 内部，Hash 并不总是用同一种结构保存。Redis 会根据字段数量、字段和值的大小，在不同内部编码之间选择：

- **listpack**：紧凑连续内存结构，适合字段少、字段和值都比较小的 Hash。
- **hashtable**：真正的哈希表结构，适合字段多、字段或值较大的 Hash。

较老版本 Redis 中，小 Hash 曾使用 `ziplist` 编码。Redis 7 之后，`ziplist` 已逐步被 `listpack` 替代。学习原理时可以把二者都理解为“紧凑型连续存储”，但实际观察当前版本时更常见的是 `listpack`。

可以使用 `OBJECT ENCODING` 观察 Hash 当前的内部编码：

```redis
HSET user:1 name Alice age 20 city Shanghai
OBJECT ENCODING user:1
```

## listpack：小 Hash 的紧凑表示

当 Hash 中字段数量较少，并且每个 field、value 都不大时，Redis 会优先使用 `listpack`。

`listpack` 可以理解为一段连续内存，里面顺序保存 field 和 value：

```text
field1, value1, field2, value2, field3, value3 ...
```

例如：

```text
user:1
  name -> Alice
  age  -> 20
```

在 listpack 中大致会按如下顺序保存：

```text
name, Alice, age, 20
```

这种结构的优点是内存紧凑：

- 不需要为每个 field-value 单独分配哈希表节点；
- 连续内存有更好的缓存局部性；
- 对大量小对象来说，可以明显减少元数据开销。

它的缺点是查找、更新、删除字段时通常需要从头顺序扫描。字段数量很小时，这个成本可以接受；字段变多后，线性扫描会变慢。

因此，listpack 的设计目标不是让单次查找达到最快，而是在“小 Hash”场景下用更少内存换取足够好的性能。

## hashtable：大 Hash 的哈希表表示

当 Hash 变大，或者某个 field/value 超过紧凑编码阈值时，Redis 会把 Hash 转换为 `hashtable` 编码。

hashtable 编码下，Hash 内部会使用字典结构保存 field 到 value 的映射：

```text
dict
  name -> Alice
  age  -> 20
  city -> Shanghai
```

字典通过哈希函数把 field 映射到哈希表槽位。理想情况下，`HGET`、`HSET`、`HEXISTS`、`HDEL` 这类按字段访问的操作可以接近 O(1)。

hashtable 的优点是字段查找和更新更快，适合大 Hash。缺点是内存开销更高：

- 哈希表本身需要桶数组；
- 每个字段和值通常还需要对象或字符串结构；
- 哈希表节点、指针、内存分配器对齐都会带来额外成本。

所以，Redis 不会一开始就给所有 Hash 使用 hashtable，而是让小 Hash 先使用更省内存的 listpack。

## 编码转换条件

Hash 从 listpack 转换为 hashtable，通常受两个配置项影响：

```text
hash-max-listpack-entries
hash-max-listpack-value
```

常见默认含义如下：

| 配置项 | 作用 |
| --- | --- |
| `hash-max-listpack-entries` | listpack 编码允许的最大字段数量 |
| `hash-max-listpack-value` | listpack 编码允许的最大 field 或 value 字节长度 |

只要超过其中任意一个阈值，Redis 就可能把该 Hash 转换为 hashtable。

例如：

```redis
CONFIG GET hash-max-listpack-entries
CONFIG GET hash-max-listpack-value
```

需要注意：

- 编码转换通常是单向的。Hash 从 listpack 转为 hashtable 后，即使后来删除字段，也不会自动转回 listpack。
- 阈值是为了平衡内存和性能，不应只为了“看起来省内存”盲目调大。
- 如果 listpack 阈值设置过大，大 Hash 的字段查找可能因为线性扫描而变慢。

## 字典、哈希冲突与渐进式 rehash

hashtable 编码背后使用 Redis 的字典结构。字典的核心目标是通过哈希值快速定位 field。

当多个 field 被映射到同一个槽位时，会发生哈希冲突。Redis 会在槽位中保存冲突元素，并继续比较实际 field 内容来确认目标字段。只要哈希函数分布较好、哈希表容量合适，冲突带来的成本通常可控。

随着字段不断增加，哈希表需要扩容；随着字段大量删除，哈希表也可能缩容。Redis 为了避免一次性 rehash 阻塞主线程，会使用渐进式 rehash：

1. 准备一张新的哈希表；
2. 后续每次执行字典相关操作时，顺带搬迁一小部分旧数据；
3. 搬迁完成后，释放旧哈希表。

渐进式 rehash 能把一次大的迁移拆成很多小步骤，减少单次命令的延迟尖刺。

不过，渐进式 rehash 并不表示所有大 Key 操作都没有风险。`HGETALL`、`HKEYS`、`HVALS`、删除大 Hash、持久化和复制大 Hash，仍然可能带来明显压力。

## 常见命令复杂度

Hash 命令的复杂度与内部编码和访问范围有关。

| 命令 | 典型时间复杂度 | 说明 |
| --- | --- | --- |
| `HGET`、`HEXISTS` | O(1) | hashtable 下接近 O(1)；listpack 下可能需要线性扫描 |
| `HSET` | O(1) | hashtable 下接近 O(1)；触发编码转换或 rehash 时成本会增加 |
| `HDEL` | O(1) | 删除单个字段通常较快；删除大量字段与字段数量有关 |
| `HLEN` | O(1) | 字段数量保存在元数据中 |
| `HSTRLEN` | O(1) | 找到字段后可直接读取值长度 |
| `HMGET` | O(N) | N 为请求的字段数量 |
| `HGETALL`、`HKEYS`、`HVALS` | O(N) | N 为 Hash 中字段数量 |
| `HSCAN` | O(1) 单次调用，完整遍历 O(N) | 单次只扫描一部分，完整扫完仍与字段数量相关 |
| `HINCRBY`、`HINCRBYFLOAT` | O(1) | 找到字段后解析并写回数值 |

复杂度里的 O(1) 不代表没有成本。大字段值仍然有内存复制、网络传输、AOF 记录和复制传播成本。

## Hash 与大 Key 问题

Hash 常被用来减少 Redis key 的数量，例如把一个用户的多个属性放到同一个 Hash 中。这通常是合理的，但不能把 Hash 当成无限容量的表。

一个很大的 Hash 会带来多方面问题：

- `HGETALL`、`HKEYS`、`HVALS` 一次返回大量数据，阻塞主线程并占用网络带宽；
- RDB 保存、AOF 重写、主从复制需要处理大量字段；
- 删除大 Hash 时释放内存可能造成延迟尖刺；
- 单个 key 过大时，Cluster 中数据分布也会不均衡。

如果一个 Hash 可能持续增长，应考虑按业务维度拆分，例如按用户、日期、分片编号或业务状态拆成多个 key。

```text
user:1001:profile
user:1001:counter
article:2026-07:views
article:2026-08:views
```

拆分的目标不是机械地让每个 key 很小，而是避免任何一个 key 成为读写、删除、迁移和恢复时的瓶颈。

## 字段级过期的影响

普通 Redis 过期时间作用于整个 key。Hash 字段级过期是 Redis 7.4 起提供的新能力，可以给单个 field 设置 TTL。

从原理上看，字段级过期意味着 Redis 需要额外记录 field 的过期元数据，并在访问、主动过期扫描或后台维护过程中清理过期字段。

这带来两个实践影响：

- 字段级过期可以减少拆 key 的需求，适合字段生命周期不同但仍属于同一对象的场景。
- 过期元数据本身也占内存；如果字段数量巨大且 TTL 变化频繁，仍要评估内存和清理成本。

如果业务字段生命周期完全一致，给整个 Hash key 设置 `EXPIRE` 通常更简单，额外开销也更低。

## 与 String 保存 JSON 的对比

同一个对象既可以保存为 Hash，也可以序列化成 JSON 后保存为 String。

Hash 的优势：

- 可以独立读写某个字段；
- 多个小字段能共享一个 Redis key，减少键空间元数据；
- 计数字段可以直接使用 `HINCRBY` 原子递增。

String JSON 的优势：

- 对应用层结构更直观；
- 一次读写整个对象更简单；
- 适合字段总是整体读取、整体更新的场景。

选择时可以用一个简单原则：

- 经常按字段读写，字段数量有限，使用 Hash；
- 经常整体读写，结构嵌套复杂，使用 String JSON；
- 需要按对象内部字段查询，不应只依赖 Redis Hash，需要额外索引或数据库查询能力。

## 实践建议

- 小对象、少字段、字段独立读写时，Hash 很合适。
- 不要把持续增长的数据无限塞进单个 Hash，避免形成大 Key。
- 读取少量字段时使用 `HGET`、`HMGET`，不要为了一个字段使用 `HGETALL`。
- 大 Hash 遍历使用 `HSCAN`，并让业务逻辑容忍重复和非快照结果。
- 可以用 `OBJECT ENCODING` 观察内部编码，但业务逻辑不应依赖具体编码。
- 调整 `hash-max-listpack-entries` 和 `hash-max-listpack-value` 前，应结合压测观察内存、延迟和 CPU。
- 生命周期一致的数据优先使用 key 级过期；生命周期不同且 Redis 版本支持时，再考虑字段级过期。
