# Redis Sorted Set 类型

Sorted Set（有序集合，也常写作 ZSet）是 Redis 中用于保存多个不重复 member，并为每个 member 关联一个 score 的数据类型。

可以把 Sorted Set 理解为一组带分数的唯一成员：

```text
game:rank
  user:2 -> 9800
  user:1 -> 8600
  user:3 -> 7300
```

Sorted Set 中的 member 本质上是字符串，支持二进制安全数据；score 是双精度浮点数。member 不能重复，同一个 member 再次写入时会更新它的 score。

Sorted Set 会按 score 排序。score 相同时，Redis 会按 member 的字典序排序。因此，Sorted Set 适合处理排行榜、延时队列、按时间排序的数据、优先级任务、热门内容等场景。

## 1. 基本用法

### 1.1 添加或更新成员

添加一个或多个成员：

```redis
ZADD game:rank 8600 user:1 9800 user:2 7300 user:3
```

`ZADD` 会创建 Sorted Set 键，或向已有 Sorted Set 中加入成员。返回值表示本次新增的成员数量；如果只是更新已有成员的 score，不计入新增数量。

更新已有成员分数：

```redis
ZADD game:rank 9000 user:1
```

`user:1` 已存在时，新的 score 会覆盖旧 score。

`ZADD` 支持一些常用选项：

| 选项 | 作用 |
| --- | --- |
| `NX` | 只添加不存在的成员，不更新已有成员 |
| `XX` | 只更新已有成员，不添加新成员 |
| `GT` | 仅当新 score 大于当前 score 时才更新 |
| `LT` | 仅当新 score 小于当前 score 时才更新 |
| `CH` | 返回新增和被修改的成员总数 |
| `INCR` | 对单个成员执行增量加分，效果类似 `ZINCRBY` |

示例：

```redis
ZADD game:rank NX 6000 user:4
ZADD game:rank GT 9200 user:1
ZADD game:rank XX CH 9900 user:2
```

`NX`、`XX`、`GT`、`LT` 不能随意组合。实际使用前应根据 Redis 版本确认支持情况。

### 1.2 获取成员分数

```redis
ZSCORE game:rank user:1
```

`ZSCORE` 返回成员对应的 score。成员不存在或键不存在时返回 nil。

Redis 6.2 起支持一次获取多个成员的分数：

```redis
ZMSCORE game:rank user:1 user:2 user:404
```

`ZMSCORE` 会按请求 member 的顺序返回结果，不存在的成员在对应位置返回 nil。

### 1.3 统计成员数量

```redis
ZCARD game:rank
```

`ZCARD` 返回 Sorted Set 中的成员数量。

按 score 范围统计数量：

```redis
ZCOUNT game:rank 8000 10000
```

按字典序范围统计数量：

```redis
ZLEXCOUNT dict:words [a [z
```

| 命令 | 作用 |
| --- | --- |
| `ZCARD key` | 返回成员总数 |
| `ZCOUNT key min max` | 返回 score 在指定范围内的成员数量 |
| `ZLEXCOUNT key min max` | 当多个成员 score 相同时，按 member 字典序统计数量 |

### 1.4 增加或减少分数

```redis
ZINCRBY game:rank 100 user:1
ZINCRBY game:rank -50 user:3
```

`ZINCRBY` 会把指定成员的 score 增加给定值，并返回更新后的 score。成员不存在时，会先以 `0` 为初始分数创建该成员。

它适合处理积分、热度、播放量权重等增量变化。单条 `ZINCRBY` 是原子的，并发请求不会覆盖彼此的增量。

### 1.5 删除成员

删除一个或多个指定成员：

```redis
ZREM game:rank user:3
ZREM game:rank user:1 user:2
```

`ZREM` 返回实际删除的成员数量。不存在的成员会被忽略。

当一个 Sorted Set 中的最后一个成员被删除后，这个键也会被 Redis 自动删除。

删除整个 Sorted Set 键仍然使用通用命令：

```redis
DEL game:rank
UNLINK game:rank
```

`DEL` 同步删除键；`UNLINK` 会先从键空间移除键，再由后台线程释放内存。对于包含大量成员的大 Sorted Set，可以考虑使用 `UNLINK` 降低释放内存对主线程的影响。

### 1.6 获取排名

按 score 从小到大获取排名：

```redis
ZRANK game:rank user:1
```

按 score 从大到小获取排名：

```redis
ZREVRANK game:rank user:1
```

排名从 `0` 开始。返回 `0` 表示第一名，返回 `1` 表示第二名。成员不存在时返回 nil。

Redis 7.2 起，`ZRANK` 和 `ZREVRANK` 可以带 `WITHSCORE` 同时返回排名和分数：

```redis
ZREVRANK game:rank user:1 WITHSCORE
```

如果业务展示需要从 `1` 开始的名次，应用层通常把 Redis 返回的排名加 `1`。

### 1.7 按排名范围查询

按 score 从小到大返回指定排名区间：

```redis
ZRANGE game:rank 0 9
```

按 score 从大到小返回指定排名区间：

```redis
ZREVRANGE game:rank 0 9
```

带上分数：

```redis
ZRANGE game:rank 0 9 WITHSCORES
ZREVRANGE game:rank 0 9 WITHSCORES
```

下标从 `0` 开始，`0 9` 表示前 10 个成员。下标支持负数，`-1` 表示最后一个成员。

Redis 6.2 起推荐使用更统一的 `ZRANGE` 语法：

```redis
ZRANGE game:rank 0 9 REV WITHSCORES
```

上例等价于按 score 从大到小返回前 10 名，并携带 score。

### 1.8 按 score 范围查询

查询 score 在指定范围内的成员：

```redis
ZRANGEBYSCORE game:rank 8000 10000 WITHSCORES
```

从高到低查询：

```redis
ZREVRANGEBYSCORE game:rank 10000 8000 WITHSCORES
```

分页限制返回数量：

```redis
ZRANGEBYSCORE game:rank 8000 10000 WITHSCORES LIMIT 0 20
```

区间默认是闭区间。可以使用 `(` 表示开区间：

```redis
ZRANGEBYSCORE game:rank (8000 10000
```

也可以使用 `-inf` 和 `+inf` 表示负无穷和正无穷：

```redis
ZRANGEBYSCORE game:rank -inf +inf WITHSCORES
```

Redis 6.2 起也可以使用统一的 `ZRANGE` 语法：

```redis
ZRANGE game:rank 8000 10000 BYSCORE WITHSCORES
ZRANGE game:rank 10000 8000 BYSCORE REV WITHSCORES
```

### 1.9 按字典序范围查询

当多个成员的 score 相同时，可以按 member 的字典序范围查询：

```redis
ZADD dict:words 0 apple 0 banana 0 cherry 0 date
ZRANGEBYLEX dict:words [banana [date
```

`[` 表示包含边界，`(` 表示不包含边界：

```redis
ZRANGEBYLEX dict:words (banana [date
```

删除字典序范围内的成员：

```redis
ZREMRANGEBYLEX dict:words [banana [date
```

字典序查询通常要求参与比较的成员使用相同 score。如果 score 不同，Sorted Set 会先按 score 排序，字典序范围的结果可能不符合直觉。

### 1.10 删除范围内的成员

按排名范围删除：

```redis
ZREMRANGEBYRANK game:rank 0 -101
```

上例删除低分区间，只保留分数最高的 100 个成员。因为 Sorted Set 默认按 score 从小到大排名，所以要保留高分成员时，删除的是靠前的低分排名。

按 score 范围删除：

```redis
ZREMRANGEBYSCORE delay:queue -inf 1782552000
```

上例删除 score 小于等于指定时间戳的成员。

按字典序范围删除：

```redis
ZREMRANGEBYLEX dict:words [a [c
```

范围删除可能一次移除大量成员。对大 Sorted Set 执行前应评估耗时和阻塞影响。

### 1.11 弹出成员

弹出分数最低的成员：

```redis
ZPOPMIN game:rank
```

弹出分数最高的成员：

```redis
ZPOPMAX game:rank
```

一次弹出多个成员：

```redis
ZPOPMAX game:rank 3
```

`ZPOPMIN` 和 `ZPOPMAX` 会返回成员及其 score，并从原 Sorted Set 中删除这些成员。

需要阻塞等待元素时，可以使用阻塞弹出命令：

```redis
BZPOPMIN delay:queue 5
BZPOPMAX priority:jobs 5
```

`5` 表示最多阻塞等待 5 秒。阻塞弹出适合简单队列场景，但复杂的可靠队列通常还需要处理中途失败、重试和确认机制。

### 1.12 增量遍历

对于成员较多的 Sorted Set，推荐使用 `ZSCAN` 增量遍历：

```redis
ZSCAN game:rank 0 MATCH user:* COUNT 20
```

```text
ZSCAN key cursor [MATCH pattern] [COUNT count]
```

- `cursor` 是游标，第一次扫描传 `0`。
- `MATCH` 用于按模式过滤 member。
- `COUNT` 是扫描工作量提示，不保证每次一定返回指定数量。

客户端应使用返回的新游标继续下一次扫描，直到游标变为 `0`。

`ZSCAN` 不能提供强一致快照。扫描过程中新增、修改或删除成员时，结果可能反映这些变化，也可能出现重复，业务代码需要能容忍这种情况。

## 2. 集合运算

Sorted Set 支持带权重的并集和交集运算，也支持差集运算。

### 2.1 交集

交集表示同时存在于多个 Sorted Set 中的成员。结果成员的 score 会根据输入集合的 score 计算得出。

```redis
ZINTER 2 user:1:tags user:2:tags WITHSCORES
```

把交集结果保存到目标 Sorted Set：

```redis
ZINTERSTORE common:tags:1:2 2 user:1:tags user:2:tags
```

可以指定权重和聚合方式：

```redis
ZINTER 2 rank:a rank:b WEIGHTS 1 2 AGGREGATE SUM WITHSCORES
```

`WEIGHTS 1 2` 表示第一个集合的 score 乘以 1，第二个集合的 score 乘以 2。`AGGREGATE` 可以使用 `SUM`、`MIN` 或 `MAX`。

### 2.2 并集

并集表示多个 Sorted Set 中出现过的所有成员：

```redis
ZUNION 2 rank:a rank:b WITHSCORES
```

把并集结果保存到目标 Sorted Set：

```redis
ZUNIONSTORE rank:merged 2 rank:a rank:b
```

并集适合合并多个榜单、多个推荐来源或多个热度来源。使用权重可以调整不同来源的影响：

```redis
ZUNION 2 rank:views rank:likes WEIGHTS 1 10 AGGREGATE SUM WITHSCORES
```

### 2.3 差集

差集表示存在于第一个 Sorted Set、但不存在于后续 Sorted Set 的成员：

```redis
ZDIFF 2 all:items hidden:items WITHSCORES
```

把差集结果保存到目标 Sorted Set：

```redis
ZDIFFSTORE visible:items 2 all:items hidden:items
```

差集适合从候选列表中排除已读、已购买、已屏蔽或黑名单成员。

### 2.4 集合运算注意事项

| 命令 | 作用 |
| --- | --- |
| `ZINTER numkeys key [key ...]` | 返回多个 Sorted Set 的交集 |
| `ZINTERSTORE destination numkeys key [key ...]` | 将交集保存到目标 Sorted Set |
| `ZUNION numkeys key [key ...]` | 返回多个 Sorted Set 的并集 |
| `ZUNIONSTORE destination numkeys key [key ...]` | 将并集保存到目标 Sorted Set |
| `ZDIFF numkeys key [key ...]` | 返回第一个 Sorted Set 相对其他 Sorted Set 的差集 |
| `ZDIFFSTORE destination numkeys key [key ...]` | 将差集保存到目标 Sorted Set |

集合运算可能读取、计算和返回大量成员。参与运算的 Sorted Set 很大时，应评估耗时、内存和网络开销。需要反复读取的结果可以保存到临时键，并为临时键设置合理过期时间。

在 Redis Cluster 中，多键 Sorted Set 运算通常要求相关键位于同一个哈希槽。可以通过哈希标签设计键名，例如 `rank:{game}:daily` 和 `rank:{game}:weekly`。

## 3. 应用场景

### 3.1 排行榜

Sorted Set 最典型的场景是排行榜。member 保存用户 ID，score 保存积分、分数或热度：

```redis
ZADD game:rank 8600 user:1 9800 user:2 7300 user:3
ZINCRBY game:rank 120 user:1
ZRANGE game:rank 0 9 REV WITHSCORES
ZREVRANK game:rank user:1
```

`ZRANGE ... REV WITHSCORES` 可以取前 N 名，`ZREVRANK` 可以查某个用户的名次。

如果只需要保留前 1000 名，可以定期裁剪：

```redis
ZREMRANGEBYRANK game:rank 0 -1001
```

注意这里的排名按从小到大计算。删除 `0 -1001` 表示移除低分成员，只保留分数最高的 1000 名。

### 3.2 延时队列

可以把任务 ID 作为 member，把任务到期时间戳作为 score：

```redis
ZADD delay:queue 1782552000 order:1001
ZRANGE delay:queue -inf 1782552000 BYSCORE LIMIT 0 10
ZREM delay:queue order:1001
```

消费者定时查询 score 小于等于当前时间的任务，成功获取后再删除。为了避免多个消费者重复处理，查询和删除通常需要用 Lua 脚本、Redis Functions 或事务保证原子性。

Sorted Set 适合实现轻量延时调度。对可靠性、重试、死信队列和消费确认要求较高时，应考虑专门的消息队列。

### 3.3 热门内容

可以把文章、视频或商品 ID 作为 member，把热度值作为 score：

```redis
ZINCRBY hot:articles 1 article:1001
ZINCRBY hot:articles 5 article:1002
ZRANGE hot:articles 0 19 REV WITHSCORES
```

热度可以来自阅读、点赞、评论、收藏等行为。不同事件可以使用不同增量，也可以结合时间衰减定期降低旧内容的 score，避免历史内容长期霸榜。

### 3.4 按时间排序的数据

把时间戳作为 score，可以按时间范围查询数据：

```redis
ZADD user:1001:timeline 1782552000 post:1 1782555600 post:2
ZRANGE user:1001:timeline 1782550000 1782560000 BYSCORE
```

这种方式适合保存最近动态、浏览记录、操作日志索引等。实际内容通常保存在 String、Hash 或数据库中，Sorted Set 只保存 ID 和排序信息。

对于长期增长的数据，应定期删除过旧成员：

```redis
ZREMRANGEBYSCORE user:1001:timeline -inf 1780000000
```

### 3.5 优先级任务

可以把任务 ID 作为 member，把优先级作为 score：

```redis
ZADD priority:jobs 10 job:low 100 job:high
ZPOPMAX priority:jobs
```

`ZPOPMAX` 会取出 score 最高的任务并删除，适合简单的优先级消费模型。如果任务处理失败，需要把任务重新放回队列，或维护处理中队列和重试次数。

## 4. 使用建议

- Sorted Set 适合需要排序、排名、范围查询的唯一成员集合。
- member 唯一，score 可以重复。score 相同时，Redis 会按 member 字典序排序。
- 大 Sorted Set 避免频繁执行无边界范围查询，例如 `ZRANGE key 0 -1`。
- 排行榜通常用 `ZRANGE key 0 N REV WITHSCORES` 获取前 N 名，用 `ZREVRANK` 获取个人排名。
- 延时队列中“取到期任务并删除”应保证原子性，避免多个消费者重复处理同一任务。
- score 是浮点数。金额、精确计数等数据尽量用整数表达，避免浮点精度问题。
- 多键集合运算在 Redis Cluster 中通常要求键位于同一个哈希槽。
- 对大 Sorted Set 做范围删除、集合运算或全量扫描前，应评估对 Redis 主线程、持久化和复制的影响。

## 5. 常用命令速查

| 分类 | 命令 |
| --- | --- |
| 写入 | `ZADD`、`ZINCRBY` |
| 读取分数 | `ZSCORE`、`ZMSCORE` |
| 统计 | `ZCARD`、`ZCOUNT`、`ZLEXCOUNT` |
| 排名 | `ZRANK`、`ZREVRANK` |
| 范围查询 | `ZRANGE`、`ZREVRANGE`、`ZRANGEBYSCORE`、`ZREVRANGEBYSCORE`、`ZRANGEBYLEX` |
| 删除 | `ZREM`、`ZREMRANGEBYRANK`、`ZREMRANGEBYSCORE`、`ZREMRANGEBYLEX`、`DEL`、`UNLINK` |
| 弹出 | `ZPOPMIN`、`ZPOPMAX`、`BZPOPMIN`、`BZPOPMAX` |
| 集合运算 | `ZINTER`、`ZINTERSTORE`、`ZUNION`、`ZUNIONSTORE`、`ZDIFF`、`ZDIFFSTORE` |
| 增量遍历 | `ZSCAN` |

> 不同 Redis 版本支持的命令和选项可能不同。使用前可通过 `COMMAND INFO <command>` 查看当前实例的命令信息，并结合所用版本的官方文档确认行为。
