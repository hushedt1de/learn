# Redis String 类型

String（字符串）是 Redis 中最基础、最常用的数据类型，可以用来保存文本、整数、浮点数、JSON、序列化对象以及图片等二进制数据。

String 底层存储的是字节序列，并且是二进制安全的：数据中可以包含任意字节，因为它不依赖 `\0` 来判断字符串结尾。

虽然单个 String 值默认最大为 512 MB，但是大对象会增加网络传输、内存占用和阻塞时间，因此实践中通常应避免把过大的值放入 Redis：一般最多 KB 级到几十 KB。超过 100 KB 就要谨慎，超过 1 MB 一般不建议。

## 1. 基本用法

### 1.1 获取

```redis
GET user:1:name
```

`GET` 成功时返回键对应的值，键不存在时返回 nil。

| 命令 | 作用 |
| --- | --- |
| `GETSET key value` | 设置新值并返回旧值；为兼容保留，通常优先使用 `SET key value GET` |
| `GETDEL key` | 获取旧值并删除键 |

```redis
SET config:version "v2" GET
GETDEL one-time-token:xyz
```

`GETDEL` 会在一条原子命令中返回 String 的旧值并删除该键，适合消费一次性令牌。
如果键保存的不是 String 类型，执行 String 专用命令会返回 `WRONGTYPE` 错误。

### 1.2 设置/更新

```redis
SET user:1:name "Alice"
```

`SET` 成功时通常返回 `OK`。
推荐使用便于识别业务含义的键名，例如 `user:1:name`、`article:1001:views`。冒号只是命名约定，不代表 Redis 会自动建立层级结构。

| 命令 | 含义 |
| --- | --- |
| `SET user:1:name "Alice" NX` | 仅当键不存在时写入 |
| `SET verification:13800000000 "927461" EX 60` | 写入数据，并让它在 60 秒后过期 |
| `SET request:20260623:001 "processing" NX EX 30` | 仅当键不存在时写入，并在 30 秒后过期 |

官方文档通常将完整语法写成下面这种形式：

```text
SET key value [NX | XX] [GET] [EX seconds | PX milliseconds |
    EXAT unix-time-seconds | PXAT unix-time-milliseconds | KEEPTTL]
```

- `[]` 表示其中的内容可以不写。
- `|` 表示从几个选项中选择一个。例如 `NX | XX` 表示可以使用 `NX` 或 `XX`，但不能同时使用。
- `seconds`、`milliseconds` 等单词是参数占位符，使用时要替换成具体数字。
- 不需要某项功能时，直接省略对应选项。

| 选项 | 作用 |
| --- | --- |
| `NX` | 仅当键不存在时写入 |
| `XX` | 仅当键已经存在时写入 |
| `GET` | 写入新值并返回旧值 |
| `EX seconds` | 设置秒级过期时间 |
| `PX milliseconds` | 设置毫秒级过期时间 |
| `EXAT timestamp` | 设置秒级 Unix 时间戳作为过期时刻 |
| `PXAT timestamp` | 设置毫秒级 Unix 时间戳作为过期时刻 |
| `KEEPTTL` | 更新值时保留键原有的过期时间 |

`SET key value` 默认会覆盖旧值，并移除该键原有的过期时间。需要保留有效期时应显式使用 `KEEPTTL`。

### 1.3 删除

#### 使用 `DEL` 直接删除

删除一个键：

```redis
DEL user:1:name
```

也可以一次删除多个键：

```redis
DEL user:1:name user:1:city user:1:age
```

`DEL` 不仅能删除 String，也能删除 Hash、List 等其他类型的键。返回实际删除的键数量，不存在的键会被忽略。

#### 使用 `UNLINK` 异步释放内存

```redis
UNLINK cache:large-object
```

`UNLINK` 会先从键空间移除键，再由后台线程释放该键占用的内存。客户端会立刻无法读取这个键。它的返回值和 `DEL` 一样，是实际移除的键数量。

对于普通的小 String，通常直接使用 `DEL` 即可。删除占用内存较大的键或包含大量元素的复杂数据结构时，可以考虑 `UNLINK`，降低释放内存对 Redis 主线程的影响。

#### 根据旧值决定是否删除

Redis 8.4 起提供 `DELEX`，可以仅在 String 当前值满足条件时删除键。比较和删除由一条命令完成，不会被其他客户端的命令插入。

```redis
SET lock:order:1001 "client-a"
DELEX lock:order:1001 IFEQ "client-a"
```

上例仅当锁的当前值仍为 `client-a` 时才删除。这样可以避免客户端持有的锁过期后，误删其他客户端后来创建的锁。

`DELEX` 支持以下条件：

| 条件 | 删除条件 |
| --- | --- |
| `IFEQ value` | 当前值等于指定值 |
| `IFNE value` | 当前值不等于指定值 |
| `IFDEQ digest` | 当前值的摘要等于指定摘要 |
| `IFDNE digest` | 当前值的摘要不等于指定摘要 |

命令返回 `1` 表示键已删除，返回 `0` 表示键不存在或条件不满足。`DELEX` 只能处理 String；键属于其他类型时返回 `WRONGTYPE` 错误。

摘要可以通过 Redis 8.4 起提供的 `DIGEST` 命令获得：

```redis
DIGEST cache:large-object
DELEX cache:large-object IFDEQ 0123456789abcdef
```

摘要是 String 内容的 16 位十六进制 XXH3 哈希值。比较大 String 时，使用摘要可以避免把完整旧值再次发送给 Redis；示例中的摘要需要替换为 `DIGEST` 实际返回的结果。

### 1.4 批量操作

```redis
MSET user:1:name "Alice" user:1:city "Shanghai"
MGET user:1:name user:1:city user:2:name
```

- `MSET` 一次写入多个键值对。
- `MGET` 一次读取多个键；不存在的键在对应位置返回 nil。
- `MSETNX` 仅在所有指定键都不存在时才写入；只要一个键已存在，就不会写入任何键。

批量命令可以减少网络往返次数。在 Redis Cluster 中，多键命令通常要求相关键位于同一个哈希槽，可通过哈希标签设计键名，例如 `user:{1}:name` 和 `user:{1}:city`。

### 1.5 修改和截取字符串

| 命令 | 作用 | 示例 |
| --- | --- | --- |
| `APPEND key value` | 将内容追加到值末尾，返回追加后的字节长度 | `APPEND log:1 " world"` |
| `STRLEN key` | 获取值的字节长度 | `STRLEN user:1:name` |
| `GETRANGE key start end` | 按字节下标读取闭区间，支持负数下标 | `GETRANGE message 0 4` |
| `SETRANGE key offset value` | 从指定字节偏移量开始覆盖内容 | `SETRANGE message 6 "Redis"` |

`STRLEN`、`GETRANGE` 和 `SETRANGE` 都按字节处理，而不是按 Unicode 字符处理。UTF-8 中文通常占多个字节，直接按偏移截取可能得到无效文本。

### 1.6 数值操作

String 可以保存十进制整数或浮点数字符串，并通过原子命令进行计算：

```redis
SET article:1001:views 0
INCR article:1001:views
INCRBY article:1001:views 10
DECR article:1001:views
DECRBY article:1001:views 3

SET product:1:price 19.5
INCRBYFLOAT product:1:price 0.5
```

| 命令 | 作用 |
| --- | --- |
| `INCR key` | 整数加 1 |
| `INCRBY key increment` | 整数增加指定值，增量可以为负数 |
| `DECR key` | 整数减 1 |
| `DECRBY key decrement` | 整数减少指定值 |
| `INCRBYFLOAT key increment` | 浮点数增加指定值 |

键不存在时，这些命令会先将其视为 `0`。整数运算要求值能够表示为 64 位有符号整数；值格式错误或运算溢出时命令会失败。金额等需要精确计算的数据，建议以最小货币单位保存为整数，例如用“分”而不是浮点数保存价格。

### 1.7 位操作

String 也可以看作位数组：

| 命令 | 作用 |
| --- | --- |
| `SETBIT key offset value` | 设置指定偏移位，`value` 只能是 `0` 或 `1` |
| `GETBIT key offset` | 获取指定偏移位 |
| `BITCOUNT key [start end]` | 统计值中为 `1` 的位数 |
| `BITOP operation destkey key...` | 对一个或多个 String 执行位运算 |
| `BITPOS key bit` | 查找第一个指定 bit 的位置 |
| `BITFIELD` / `BITFIELD_RO` | 将部分位区间作为整数读写 |

例如，用用户编号作为偏移量记录每日签到：

```redis
SETBIT sign:2026-06-23 1001 1
GETBIT sign:2026-06-23 1001
BITCOUNT sign:2026-06-23
```

位操作非常节省空间，但偏移量过大且稀疏时仍可能创建很大的 String，应预先评估最大编号和内存占用。

### 1.8 与过期时间配合

过期时间属于键，而不是 String 值本身，相关命令适用于多种 Redis 数据类型：

```redis
EXPIRE session:abc 1800
PEXPIRE session:abc 1800000
TTL session:abc
PTTL session:abc
PERSIST session:abc
```

- `EXPIRE` 给键设置或更新 `秒级` 过期时间，`PEXPIRE` 给键设置或更新 `毫秒级` 过期时间。
- `TTL` 返回剩余秒数，`PTTL` 返回剩余毫秒数。返回 `-1` 表示键存在但没有过期时间，返回 `-2` 表示键不存在。
- `PERSIST` 可以移除过期时间，使键永久保存。

## 2. 应用场景

### 2.1 缓存对象或页面片段

可以将 JSON 或序列化后的对象保存为 String：

```redis
SET cache:product:1001 '{"id":1001,"name":"keyboard","price":29900}' EX 300
```

这种方式读取整个对象很方便，但修改单个字段通常需要先反序列化再整体写回。如果经常独立读写多个字段，可以考虑 Hash；如果对象很大，则应评估拆分、压缩以及“大 Key”问题。

常见缓存流程是：先读取缓存，未命中时查询数据库，再回填 Redis。实际系统还需要考虑缓存穿透、击穿、雪崩以及缓存和数据库之间的一致性。

### 2.2 计数器

`INCR` 系列命令适用于阅读量、点赞数、下载次数和接口调用次数等计数。单条命令是原子的，并发客户端不会发生普通“读取—加一—写回”导致的覆盖问题。

如果计数还需要同时判断上限或修改其他键，应使用 Lua 脚本、Redis Functions 或事务实现整体逻辑，不能假设多条独立命令天然具有整体原子性。

### 2.3 分布式锁

最基本的加锁操作可以使用带过期时间的条件写入：

```redis
SET lock:order:1001 "550e8400-e29b-41d4-a716-446655440000" NX PX 30000
```

值应是每个锁持有者唯一的随机标识。释放锁时必须“比较标识并删除”作为一个原子操作执行，通常使用 Lua 脚本：

```lua
if redis.call('GET', KEYS[1]) == ARGV[1] then
    return redis.call('DEL', KEYS[1])
end
return 0
```

不能直接执行 `DEL lock:order:1001`，否则当前客户端超时后，可能误删另一个客户端新获取的锁。生产级分布式锁还要评估任务超时、续期、主从切换、网络分区和 fencing token 等问题；对正确性要求很高的场景，不应把一个简单 Redis 锁当作完整的协调协议。

### 2.4 Session、验证码和一次性令牌

String 配合过期时间适合保存登录 Session、短信验证码、密码重置令牌和幂等键。创建时同时设置值和过期时间，可以避免 `SET` 成功但 `EXPIRE` 因客户端故障未执行而留下永久键。

一次性令牌可用 `GETDEL` 原子地读取并删除，避免两个请求同时消费成功。

### 2.5 限流

固定窗口限流可以用计数器和过期时间实现。但把 `INCR` 与首次设置过期时间拆成两条命令会存在中途失败的风险，通常应使用 Lua 脚本将计数和设置 TTL 合并为一个原子操作。

固定窗口在窗口边界可能出现突发流量。需要更平滑的限制时，可以考虑滑动窗口、令牌桶等算法，并根据算法选择 Sorted Set、Hash 或专用模块。

### 2.6 状态标记和位图统计

简单布尔状态可以用 `SET key 1 EX ...` 表示；用户签到、在线状态、活跃用户统计等大量密集布尔值可使用位操作，通常比“每个状态一个键”更节省内存。

## 3. 实践建议

- 将相互关联的键使用统一前缀组织，但避免过长键名造成额外内存浪费。
- 缓存和临时数据应在写入时一并设置 TTL，并为大量缓存的过期时间增加适度随机抖动，降低同时失效造成的压力。
- 优先使用 `MGET`、`MSET` 或 pipeline 减少网络往返，但控制单批次大小，避免一次处理过多数据阻塞服务。
- 避免大 Key。大值会放大网络、持久化、复制、删除和故障恢复成本；必要时拆分对象并监控实际内存占用。
- 不要将“单条命令原子”误解为“整个业务流程原子”，跨键或多步骤约束应显式使用脚本、函数或事务。
- 使用 `SET ... NX EX/PX` 代替已能由 `SET` 选项覆盖的旧式组合命令，以便在一次原子操作中完成条件写入和设置有效期。
