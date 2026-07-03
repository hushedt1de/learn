# Redis Hash 类型

Hash（哈希）是 Redis 中用于保存字段和值映射的数据类型。一个 Hash 键内部可以包含多个 field-value，适合表示一个对象、一个配置项、一组计数器或一组紧密相关的小字段。

可以把 Hash 理解为 Redis key 下的一张小表：

```text
user:1
  name  -> Alice
  age   -> 20
  city  -> Shanghai
```

Hash 的 field 和 value 本质上都是字符串，支持二进制安全数据。和将整个对象序列化成一个 String 相比，Hash 可以独立读写某个字段，不必每次读取和写回整个对象。

## 1. 基本用法

### 1.1 设置字段

设置单个字段：

```redis
HSET user:1 name "Alice"
```

`HSET` 会创建 Hash 键，或更新已有 Hash 中的字段。返回值表示本次新增的字段数量；如果只是覆盖已有字段，返回值通常为 `0`。

也可以一次设置多个字段：

```redis
HSET user:1 name "Alice" age 20 city "Shanghai"
```

仅当字段不存在时才设置：

```redis
HSETNX user:1 email "alice@example.com"
```

`HSETNX` 只处理一个字段。字段已存在时不会覆盖原值，返回 `0`；字段不存在并写入成功时返回 `1`。

### 1.2 获取字段

获取单个字段：

```redis
HGET user:1 name
```

字段不存在或键不存在时，`HGET` 返回 nil。

一次获取多个字段：

```redis
HMGET user:1 name age city phone
```

`HMGET` 会按请求字段的顺序返回结果；不存在的字段在对应位置返回 nil。

获取整个 Hash：

```redis
HGETALL user:1
```

`HGETALL` 返回所有字段和值。它适合字段数量较少的 Hash；如果 Hash 很大，使用 `HGETALL` 会一次性返回大量数据，可能造成网络传输和 Redis 主线程阻塞。

### 1.3 判断字段和统计字段数量

```redis
HEXISTS user:1 email
HLEN user:1
HSTRLEN user:1 name
```

| 命令 | 作用 |
| --- | --- |
| `HEXISTS key field` | 判断字段是否存在 |
| `HLEN key` | 返回 Hash 中字段数量 |
| `HSTRLEN key field` | 返回字段值的字节长度 |

`HSTRLEN` 按字节统计，而不是按 Unicode 字符统计。中文等 UTF-8 字符通常占多个字节。

### 1.4 删除字段和删除整个 Hash

删除一个或多个字段：

```redis
HDEL user:1 phone
HDEL user:1 email city
```

`HDEL` 返回实际删除的字段数量。字段不存在会被忽略。

当一个 Hash 中的最后一个字段被删除后，这个 Hash 键也会被 Redis 自动删除。

删除整个 Hash 键仍然使用通用命令：

```redis
DEL user:1
UNLINK user:1
```

`DEL` 同步删除键；`UNLINK` 会先从键空间移除键，再由后台线程释放内存。对于包含大量字段的大 Hash，可以考虑使用 `UNLINK` 降低释放内存对主线程的影响。

### 1.5 数值字段计算

Hash 字段可以保存数字字符串，并通过原子命令进行计算：

```redis
HSET article:1001 title "Redis Hash" views 0 likes 0
HINCRBY article:1001 views 1
HINCRBY article:1001 likes 10

HSET product:1 price 19.5
HINCRBYFLOAT product:1 price 0.5
```

| 命令 | 作用 |
| --- | --- |
| `HINCRBY key field increment` | 将字段的整数值增加指定值，增量可以为负数 |
| `HINCRBYFLOAT key field increment` | 将字段的浮点数值增加指定值 |

字段不存在时，这些命令会先把字段值视为 `0`。如果字段已有值不能解析为对应数字，命令会失败。金额等需要精确计算的数据，建议使用整数保存最小货币单位，例如用“分”保存价格。

### 1.6 获取所有字段或所有值

```redis
HKEYS user:1
HVALS user:1
```

| 命令 | 作用 |
| --- | --- |
| `HKEYS key` | 返回 Hash 中所有字段 |
| `HVALS key` | 返回 Hash 中所有值 |
| `HGETALL key` | 返回 Hash 中所有字段和值 |

这些命令都会遍历整个 Hash。字段数量较大时，应避免在高峰流量中频繁使用。

### 1.7 增量遍历

对于字段较多的 Hash，推荐使用 `HSCAN` 增量遍历：

```redis
HSCAN user:1 0 MATCH a* COUNT 20
```

`HSCAN` 每次返回一个新的游标和一批字段值。客户端应使用返回的游标继续下一次扫描，直到游标变为 `0`。

```text
HSCAN key cursor [MATCH pattern] [COUNT count]
```

- `cursor` 是游标，第一次扫描传 `0`。
- `MATCH` 用于按模式过滤字段名。
- `COUNT` 是扫描工作量提示，不保证每次一定返回指定数量。

`HSCAN` 适合后台任务、管理工具和低优先级遍历。它不能提供强一致快照：扫描过程中新增、修改或删除字段时，结果可能反映这些变化，也可能出现重复，业务代码需要能容忍这种情况。

### 1.8 与过期时间配合

普通 `EXPIRE`、`TTL`、`PERSIST` 等命令作用于整个 Hash 键：

```redis
HSET session:abc userId 1001 role "admin"
EXPIRE session:abc 1800
TTL session:abc
```

上例表示整个 `session:abc` 会在 1800 秒后过期。Redis 的普通键过期时间不针对 Hash 中的单个字段。

Redis 7.4 起支持 Hash 字段级过期，可为指定字段设置 TTL：

```redis
HSET cart:1001 item:1 2 item:2 1
HEXPIRE cart:1001 3600 FIELDS 1 item:1
HTTL cart:1001 FIELDS 1 item:1
HPERSIST cart:1001 FIELDS 1 item:1
```

| 命令 | 作用 |
| --- | --- |
| `HEXPIRE` / `HPEXPIRE` | 设置字段的秒级或毫秒级相对过期时间 |
| `HEXPIREAT` / `HPEXPIREAT` | 设置字段的秒级或毫秒级绝对过期时刻 |
| `HTTL` / `HPTTL` | 查询字段剩余过期时间 |
| `HPERSIST` | 移除字段的过期时间 |

字段级过期是较新的能力。实际使用前应确认 Redis 版本和客户端库是否支持；如果目标环境不支持，仍需要拆分键或通过应用逻辑处理字段生命周期。

## 2. 应用场景

### 2.1 保存对象属性

Hash 很适合保存字段数量有限、字段经常独立读写的对象：

```redis
HSET user:1001 id 1001 name "Alice" city "Shanghai" level 3
HGET user:1001 name
HSET user:1001 city "Beijing"
```

相比把用户对象整体保存为 JSON String，Hash 修改单个字段更方便，也能减少序列化和反序列化开销。

但 Hash 不等于关系型数据库中的表。Redis 不能直接按 Hash 内部字段建立通用索引，也不能通过 `city = Shanghai` 这类条件查询所有用户。需要按字段查询时，应额外维护索引结构，或仍由关系型数据库承担查询职责。

### 2.2 保存对象多个计数器

一个业务对象常有多种计数，例如阅读量、点赞数、收藏数和评论数：

```redis
HSET article:1001 views 0 likes 0 favorites 0 comments 0
HINCRBY article:1001 views 1
HINCRBY article:1001 likes 1
```

把这些计数放在同一个 Hash 中，可以减少键数量，并让相关数据更集中。单条 `HINCRBY` 是原子的，并发请求不会覆盖彼此的增量。

如果需要“计数加一后再判断上限、再修改其他键”，应使用 Lua 脚本、Redis Functions 或事务来保证多步逻辑的整体一致性。

### 2.3 缓存部分可变的对象

商品、用户资料、配置项等缓存对象如果经常只改其中一两个字段，可以使用 Hash：

```redis
HSET cache:product:1001 name "keyboard" price 29900 stock 120
HGET cache:product:1001 price
HSET cache:product:1001 stock 119
EXPIRE cache:product:1001 300
```

这种方式避免每次更新库存都读出完整 JSON、修改、再整体写回。实际系统仍要处理缓存与数据库之间的一致性，以及缓存穿透、击穿和雪崩等问题。

### 2.4 保存用户临时状态

Hash 可以保存同一用户、会话或连接的一组短期状态：

```redis
HSET online:user:1001 device "ios" ip "10.0.0.8" lastSeen 1782552000
EXPIRE online:user:1001 300
```

如果整组状态生命周期一致，给 Hash 键设置 TTL 即可。如果每个字段生命周期不同，需要确认是否可以使用字段级过期；否则应将生命周期不同的数据拆成多个独立键。

## 3. 实践建议

- Hash 适合保存字段数量有限、经常按字段读写的一组相关数据。
- 不要把无限增长的数据塞进一个 Hash。字段过多会形成大 Key，影响网络传输、持久化、复制、删除和故障恢复。
- 读取少量字段时优先使用 `HGET` 或 `HMGET`，避免为了一个字段执行 `HGETALL`。
- 大 Hash 遍历优先使用 `HSCAN`，并让业务逻辑能容忍重复和非快照结果。
- 普通过期时间属于整个键；字段级过期需要 Redis 7.4+ 支持。
- Hash 不能代替数据库表查询。需要按字段条件搜索时，应设计额外索引或回到数据库查询。
- 字段名也会占用内存。大量小对象使用 Hash 时，应在可读性和内存占用之间做权衡。

## 4. 常用命令速查

| 分类 | 命令 |
| --- | --- |
| 写入 | `HSET`、`HSETNX` |
| 读取 | `HGET`、`HMGET`、`HGETALL` |
| 判断与统计 | `HEXISTS`、`HLEN`、`HSTRLEN` |
| 删除 | `HDEL`、`DEL`、`UNLINK` |
| 数值计算 | `HINCRBY`、`HINCRBYFLOAT` |
| 全量获取 | `HKEYS`、`HVALS`、`HGETALL` |
| 增量遍历 | `HSCAN` |
| 生命周期 | `EXPIRE`、`TTL`、`PERSIST`、`HEXPIRE`、`HTTL`、`HPERSIST` |

> 不同 Redis 版本支持的命令和选项可能不同。使用前可通过 `COMMAND INFO <command>` 查看当前实例的命令信息，并结合所用版本的官方文档确认行为。
