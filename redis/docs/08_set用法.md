# Redis Set 类型

Set（集合）是 Redis 中用于保存多个不重复元素的数据类型。一个 Set 键内部可以包含多个 member，每个 member 都是唯一的。

可以把 Set 理解为一组无重复值：

```text
user:1:tags
  redis
  backend
  database
```

Set 中的元素本质上是字符串，支持二进制安全数据。Set 不保证元素顺序，因此不能依赖返回结果的排列顺序。如果需要按分数排序，应使用 Sorted Set；如果需要按插入顺序保存重复元素，应使用 List。

Set 适合处理去重、标签、关注关系、共同好友、黑名单、抽奖候选池等场景。

## 1. 基本用法

### 1.1 添加元素

添加一个或多个元素：

```redis
SADD user:1:tags redis backend database
```

`SADD` 会创建 Set 键，或向已有 Set 中加入元素。返回值表示本次实际新增的元素数量；如果元素已经存在，不会重复添加，也不会计入返回值。

再次添加已有元素：

```redis
SADD user:1:tags redis
```

如果 `redis` 已经存在，返回值为 `0`。

### 1.2 获取所有元素

```redis
SMEMBERS user:1:tags
```

`SMEMBERS` 会返回 Set 中的全部元素。由于 Set 无序，返回顺序不稳定，业务代码不应依赖它。

当 Set 很大时，`SMEMBERS` 会一次性返回大量数据，可能增加网络传输和 Redis 主线程压力。大集合推荐使用 `SSCAN` 增量遍历。

### 1.3 判断元素是否存在

判断单个元素是否存在：

```redis
SISMEMBER user:1:tags redis
```

返回 `1` 表示存在，返回 `0` 表示不存在。

Redis 6.2 起支持一次判断多个元素：

```redis
SMISMEMBER user:1:tags redis mysql backend
```

`SMISMEMBER` 会按请求元素的顺序返回多个结果，每个结果都是 `1` 或 `0`。

### 1.4 统计元素数量

```redis
SCARD user:1:tags
```

`SCARD` 返回 Set 中元素数量。它不需要返回所有元素，适合用来统计去重后的数量。

例如统计一篇文章的去重点赞用户数：

```redis
SADD article:1001:likes user:1 user:2 user:3
SCARD article:1001:likes
```

### 1.5 删除元素

删除一个或多个指定元素：

```redis
SREM user:1:tags database
SREM user:1:tags redis backend
```

`SREM` 返回实际删除的元素数量。不存在的元素会被忽略。

当一个 Set 中的最后一个元素被删除后，这个 Set 键也会被 Redis 自动删除。

删除整个 Set 键仍然使用通用命令：

```redis
DEL user:1:tags
UNLINK user:1:tags
```

`DEL` 同步删除键；`UNLINK` 会先从键空间移除键，再由后台线程释放内存。对于包含大量元素的大 Set，可以考虑使用 `UNLINK` 降低释放内存对主线程的影响。

### 1.6 随机获取元素

随机返回一个元素，但不删除：

```redis
SRANDMEMBER lottery:users
```

随机返回多个元素：

```redis
SRANDMEMBER lottery:users 3
```

随机弹出并删除一个元素：

```redis
SPOP lottery:users
```

随机弹出并删除多个元素：

```redis
SPOP lottery:users 3
```

| 命令 | 作用 |
| --- | --- |
| `SRANDMEMBER key [count]` | 随机返回元素，不删除 |
| `SPOP key [count]` | 随机返回元素，并从 Set 中删除 |

`SRANDMEMBER key count` 中的 `count` 为正数时，返回元素不会重复，最多返回 Set 的元素总数；`count` 为负数时，返回结果可能包含重复元素。

`SPOP` 会修改原 Set，适合抽奖后从候选池中移除中奖者。

### 1.7 移动元素

```redis
SMOVE source:set target:set member
```

示例：

```redis
SADD todo:users user:1 user:2 user:3
SMOVE todo:users done:users user:1
```

`SMOVE` 会把元素从源 Set 移动到目标 Set。移动过程是原子的，适合把任务、用户或资源从一个状态集合切换到另一个状态集合。

如果源 Set 中不存在该元素，命令返回 `0`，目标 Set 不会变化。

### 1.8 增量遍历

对于元素较多的 Set，推荐使用 `SSCAN` 增量遍历：

```redis
SSCAN user:1:tags 0 MATCH b* COUNT 20
```

```text
SSCAN key cursor [MATCH pattern] [COUNT count]
```

- `cursor` 是游标，第一次扫描传 `0`。
- `MATCH` 用于按模式过滤元素。
- `COUNT` 是扫描工作量提示，不保证每次一定返回指定数量。

客户端应使用返回的新游标继续下一次扫描，直到游标变为 `0`。

`SSCAN` 不能提供强一致快照。扫描过程中新增或删除元素时，结果可能反映这些变化，也可能出现重复，业务代码需要能容忍这种情况。

## 2. 集合运算

Set 的一个核心能力是集合运算，包括交集、并集和差集。

### 2.1 交集

交集表示同时存在于多个 Set 中的元素：

```redis
SINTER user:1:follows user:2:follows
```

常见用途是查找共同关注、共同好友、共同标签。

只返回交集数量：

```redis
SINTERCARD 2 user:1:follows user:2:follows
```

`SINTERCARD` 适合只关心数量、不需要返回具体元素的场景，可以减少网络传输。

把交集结果保存到新 Set：

```redis
SINTERSTORE common:follows:1:2 user:1:follows user:2:follows
```

### 2.2 并集

并集表示多个 Set 中出现过的所有元素，结果仍然去重：

```redis
SUNION user:1:tags user:2:tags
```

把并集结果保存到新 Set：

```redis
SUNIONSTORE all:tags:1:2 user:1:tags user:2:tags
```

并集适合合并多个来源的候选集合，例如合并多个推荐来源、多个用户分组或多个标签集合。

### 2.3 差集

差集表示存在于第一个 Set、但不存在于后续 Set 的元素：

```redis
SDIFF user:1:follows user:2:follows
```

上例可以理解为：用户 1 关注了，但用户 2 没有关注的对象。

把差集结果保存到新 Set：

```redis
SDIFFSTORE only:user:1:follows user:1:follows user:2:follows
```

差集适合排除已读、已推荐、已购买、黑名单等元素。

### 2.4 集合运算注意事项

| 命令 | 作用 |
| --- | --- |
| `SINTER key [key ...]` | 返回多个 Set 的交集 |
| `SINTERCARD numkeys key [key ...]` | 返回交集元素数量 |
| `SUNION key [key ...]` | 返回多个 Set 的并集 |
| `SDIFF key [key ...]` | 返回第一个 Set 相对其他 Set 的差集 |
| `SINTERSTORE destination key [key ...]` | 将交集保存到目标 Set |
| `SUNIONSTORE destination key [key ...]` | 将并集保存到目标 Set |
| `SDIFFSTORE destination key [key ...]` | 将差集保存到目标 Set |

集合运算可能会读取和返回大量元素。参与运算的 Set 很大时，应评估耗时、内存和网络开销。只需要数量时，优先考虑 `SCARD` 或 `SINTERCARD` 这类不会返回完整元素列表的命令。

在 Redis Cluster 中，多键集合运算通常要求相关键位于同一个哈希槽。可以通过哈希标签设计键名，例如 `user:{1}:follows` 和 `user:{1}:fans`。

## 3. 应用场景

### 3.1 去重计数

Set 天然去重，适合统计去重后的用户、设备或 IP。

```redis
SADD page:1001:uv user:1
SADD page:1001:uv user:2
SADD page:1001:uv user:1
SCARD page:1001:uv
```

即使 `user:1` 多次访问，也只会被记录一次。

如果数据量非常大，只需要估算基数而不需要保存具体元素，可以考虑 HyperLogLog。

### 3.2 点赞和收藏

```redis
SADD article:1001:likes user:1
SISMEMBER article:1001:likes user:1
SCARD article:1001:likes
SREM article:1001:likes user:1
```

Set 可以方便地处理“是否点赞”“点赞人数”“取消点赞”等操作。单条 `SADD` 和 `SREM` 是原子的，并发请求不会把同一个用户重复计数。

### 3.3 标签和分类

```redis
SADD article:1001:tags redis cache backend
SADD tag:redis:articles article:1001 article:1002
SADD tag:backend:articles article:1001 article:1003
SINTER tag:redis:articles tag:backend:articles
```

可以用 Set 保存对象拥有的标签，也可以反向保存某个标签下的对象集合。多个标签组合查询时，可以通过交集找出同时拥有这些标签的对象。

### 3.4 好友和关注关系

```redis
SADD user:1:follows user:2 user:3 user:4
SADD user:2:follows user:3 user:5

SINTER user:1:follows user:2:follows
SDIFF user:1:follows user:2:follows
```

交集可以找共同关注，差集可以找一方关注但另一方未关注的对象。对于社交关系较大的系统，需要结合分页、缓存结果和异步任务控制单次运算规模。

### 3.5 抽奖候选池

```redis
SADD lottery:2026 users:1 users:2 users:3 users:4
SPOP lottery:2026 2
```

`SPOP` 会随机取出并删除元素，适合不允许重复中奖的抽奖。需要保留原候选池时，可以先复制到临时 Set，或使用 `SRANDMEMBER` 随机读取后由业务逻辑处理结果。

## 4. 使用建议

- Set 是无序结构，不要依赖 `SMEMBERS`、`SINTER` 等命令的返回顺序。
- Set 自动去重，适合保存唯一成员，不适合保存重复值。
- 大 Set 避免频繁使用 `SMEMBERS` 和大型集合运算，优先使用 `SSCAN`、计数命令或异步任务。
- 需要排序、排名、分数时使用 Sorted Set，而不是 Set。
- 需要精确查询成员是否存在时，Set 很合适；只需要低成本概率判断时，可以考虑 Bloom Filter。
