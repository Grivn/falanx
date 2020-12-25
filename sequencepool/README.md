# Falanx: sequence pool
## Introduction
序列池，用于保存来自共识节点或客户端的日志，并进行日志序列的维护
## Component
序列池主要包含四部分，分别为：交易池、本地日志序列、客户端请求序列、副本日志序列
## Construction
### transaction
交易，其中包含：交易哈希、交易内容
- 交易哈希（hash）：交易的哈希值
- 交易内容（content）：交易内容本身
### non-select quorum-txList
还未被选中的合法交易队列，表示已经有n-f个节点完成排序的交易集合，其中包含：交易哈希列表
- 交易哈希列表（hash list）：交易的哈希值，使用list
### selected quorum-txs
已选中的合法交易，表示已选中的交易，其中包含：交易的哈希值，使用map
### ordered request
有序请求，日志以请求为粒度，一个请求中可以包含多个交易，其中包含：序号、请求内容
- 序号（sequence）：客户端发送的请求顺序，不可重复，递增处理
- 请求内容（content）：交易哈希列表
### ordered log
有序日志，日志以交易为粒度，其中包含：序号、交易
- 序号（sequence）：每个节点的日志需要排序
### transaction pool
交易池，其中包含：交易、交易哈希、已排序节点、候选节点、超时计时器
- 交易（transaction）：单笔交易
- 交易哈希（hash）：交易哈希取值，用作交易的标记
- 已排序节点（ordered_replica）：已经对该交易进行排序的节点，使用list
- 候选节点（candidates）：将参与关系图生成的节点组合，使用list
- 超时计时器（timer）：当已排序节点数目达到n-f个时开启，用于进行crash行为的检测
### local logs
本地日志序列，其中包含：有序日志
- 有序日志（ordered_log）：经过排序的节点日志，表示当前节点对于交易的处理序列
### client requests
客户端请求序列，为每个新客户端生成相应的客户端请求序列，其中包含：有序请求，缓存堆
- 有序请求（ordered_req）：表示客户端请求顺序，为顺序排列的请求，使用list
- 缓存堆（cached_heap）：用于缓存来自客户端的请求，按照请求序号进行堆排序，使用heap
### replica logs
副本日志序列，为每个共识节点生成相应的日志序列，其中包括：有序日志，缓存堆
- 有序日志（ordered_log）：表示每个副本对交易的排序情况，为顺序排序的日志，使用list
- 缓存堆（cached_heap）：用于缓存来自副本的日志，按照日志序号进行堆排序，使用heap
## Interface
### 接收来自客户端的请求
ReceiveTransactions(sr *cTypes.OrderedRequest)
- 缓存客户端的请求
- 验证客户端请求的序列
- 生成本地的排序日志
### 接收来自副本的日志
ReceiveReplicaLogs(sl *types.SequenceLog)
- 缓存来自副本的排序日志
- 验证来自副本的排序日志
### 读取序列池
LocalLogOrder() *types.SequenceLog