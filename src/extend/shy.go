package main

import (
	// 数据层
	_ "contexts/mdb" //数据中心
	_ "contexts/nfs" //存储中心
	_ "contexts/ssh" //集群中心
	_ "contexts/tcp" //网络中心
	// 控制层
	_ "contexts/gdb" //调试中心
	_ "contexts/lex" //词法中心
	_ "contexts/log" //日志中心
	_ "contexts/yac" //语法中心
	// 服务层
	_ "contexts/aaa" //认证中心
	_ "contexts/cli" //管理中心
	c "contexts/ctx" //模块中心
	_ "contexts/web" //应用中心

	// 应用层
	_ "examples/chat" //会议中心
	_ "examples/code" //代码中心
	_ "examples/mall" //交易中心
	_ "examples/team" //团队中心
	_ "examples/wiki" //文档中心

	// 应用层
	_ "examples/chat/feishu" //飞书
)

func main() {
	c.Start()
}
