# erinyes

## Introduction

erinyes 在 serveless 场景下完成了溯源图的构建，并可以实现精确的攻击溯源，支持基于溯源图的下游异常检测任务。其核心功能包括：
- 基于宿主机采集日志，支持划分容器
- 流量层追踪请求链路
- 通过执行单元划分，解决依赖爆炸问题

## Graph Builder

溯源图构建的数据来源于 sysdig 采集工具生成的审计日志，目前支持以下系统调用解析：

| 三类关系解析     | 系统调用                   | 特点                        |
| ---------------- | -------------------------- | --------------------------- |
| process  syscall | fork、vfork、clone         | 返回值表示子进程pid         |
|                  | execve                     | 进程替换                    |
| file  syscall    | write、writev、read、readv | 两类数据流动方向            |
|                  | open、openat               | 可根据参数转换为read、write |
| network  syscall | sednto、recevfrom          | UDP协议的数据流动           |
|                  | write、read                | TCP协议的数据流动           |
|                  | bind、listen、connect      | socket相关系统调用          |

sysdig 采集指令如下：

```shell
sysdig -p"*%evt.datetime %proc.name %proc.pid %proc.vpid %evt.dir %evt.type %fd.name %proc.ppid %proc.exepath %evt.rawres %fd.lip %fd.rip %fd.lport %fd.rport %container.id %container.name %evt.info" "container.id!=652f0e0e767a and container.id!=host and container.name!=<N/A> and container.image!=registry.aliyuncs.com/google_containers/pause:3.2 and (evt.type=open or evt.type=openat or evt.type=read or evt.type=write or evt.type=sendto or evt.type=recvfrom or evt.type=execve or evt.type=fork or evt.type=clone or evt.type=bind or evt.type=listen or evt.type=connect or evt.type=accept or evt.type=accept4 or evt.type=chmod or evt.type=connect)"
```

