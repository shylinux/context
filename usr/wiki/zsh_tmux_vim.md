## 《终端工具链》简介
终端工具链，就是对编程开发过程中所用到的各种命令行的工具进行高效的组合，不断的提升编程速度与开发效率。

- 在主流的系统中，Ubuntu的命令行最为强大，有丰富的命令行工具，可以很容易组合出自己的终端工具链；
- 其次是MacOSX，命令行也很丰富，再搭配上苹果电脑的硬件与系统，可以组合出很流畅的终端工具链；
- 最后是Windows，命令行功能弱的可以忽略，但可以安装一个shell工具[git-scm](https://git-scm.com/downloads)，使用一些基本的命令，如果需要更丰富的命令行工具，可以本地安装虚拟机或是远程连接云主机，使用Ubuntu。

命令行终端，与图形界面不同，是以一种文本化的方式与系统进行交互。
可以很直接、很高效执行各种系统操作，同时各种重复性的操作，都可以很方便的写成程序脚本，和系统命令一样直接调用，不断的提升操作效率。

在终端里，有大量丰富的命令可以使用，不可能全部掌握，一些基本的命令会使用即可。
但在开发流程，想提升编程速度与开发效率，就需要深入理解与熟练掌握这几个工具：zsh、tmux、docker、git、vim。

- **zsh** 和系统默认自带的bash一样，也是一种shell，不断的解析用户或脚本的输入，执行各种命令。但提供了更丰富的特性，如各种补全，命令补全、文件补全、历史补全，可以极大的提升操作效率。
- **tmux** 是一款高效的终端分屏器，可以在终端把一块屏幕分成多个小窗口，每个窗口都启动一个独立shell，这样就可以充分的利用屏幕，同时执行多个命令。
- **docker** 是一种容器软件，像虚拟机一样为应用软件提供一个完整独立的运行环境，但以一种更加轻量简捷的方式实现，极大的简化的软件的部署与分发。
- **git** 是代码的版本控制软件，用来管理代码的每次变化，分支与版本，本地与远程代码仓库，可以实现多人协作开发。
- **vim** 是一款极其强大的编辑器，通过模式化快捷键提升编辑速度，通过灵活的脚本与插件扩展丰富的功能。

使用zsh+tmux+vim的工具链，根据自己的使用习惯进行个性化配置，就可以极大的提升编程速度与开发效率。

## 基础入门
每个系统上打开终端的方式都不一样，根据自己的系统进行操作。

- 在Ubuntu中，按Ctrl+Alt+T，可以直接打开终端。
- 在Mac中，打开Finder，然后打开，应用->实用工具->终端。
- 在Windows里，先下载一个应用：[git-scm](https://git-scm.com/downloads)，按步骤安装即可，然后打开应用Git bash。

打开终端后，你就打开了一个全新的世界，通过命令行，你就可以自由自在的控制你自己的电脑，并可以直接与世界上成千上万的计算机进行各种交互。

先来体验一下几个基本的命令吧。

输入"date"，并按回车，即可查看当前日期与时间。
```
$ date
Wed Sep 12 09:32:53 CST 2018
```
输入"pwd"，并按回车，即可查看当前所在目录。
```
$ pwd
/Users/shaoying
```
在Mac上输入"open"可以打开各种应用，如访问网页。
```
$ open http://www.baidu.com
```
在Mac上将文字转换成语音播放。
```
say hello
```
查看电脑开机时长。
```
$ uptime
19:31  up 26 days, 21:21, 3 users, load averages: 2.00 1.96 1.98
```
查看主机名。
```
$ hostname
shy-MacBook-Pro.local
```
下载文件，使用wget命令，参数输入下载链接地址，即可下载文件到当前目录。
```
$ wget http://www.baidu.com
```

如果Mac上没有brew，可以安装一下[Mac包管理器](https://github.com/Homebrew/brew)。更多信息参考[HomeBrew官网](https://brew.sh/)
```
$ ruby -e "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/master/install)"
```
### zsh使用
Mac上自带zsh，不用安装，但Ubuntu上需要自己安装一下。
```
$ sudo apt-get install zsh
```
原生的zsh不是很好用，可以安装一个[zsh插件管理器](https://github.com/robbyrussell/oh-my-zsh)。
更多信息可以查看[ohmyzsh官网](https://ohmyz.sh/)。
```
$ sh -c "$(curl -fsSL https://raw.github.com/robbyrussell/oh-my-zsh/master/tools/install.sh)"
```
如果在Ubuntu上没有安装curl，可以安装一下。

在终端常用的搜索命令有find与grep。
ag就是取代grep，对指定目录下的所有文件的内容进行全文搜索。
度h
```
$ sudo apt-get install curl
```
```
$ brew install the_silver_searcher
```
```
$ sudo apt-get install silversearcher-ag
```
### tmux使用
在开发与测试的过程中，经常会同时处理多个事情，就用到多个终端窗口，可以使用SecureCRT、PuTTY、Terimal、iTerm、Konsole等软件去管理这些终端。
不过窗口一多，在一堆标签中来回切换窗口，简直是恶梦一样，经常会翻来翻去，搞半天才能找到自己想用的终端窗口。
而且尤其是连接远程服务器，网络一但断开，所有的窗口就要重新连接，而且终端里的操作历史基本上就废了，很多工作就要重新操作一遍了。

tmux就是一款极其强大的终端管理软件。
它不仅可以把一个窗口分隔成任意多个小窗口，每个窗口都是一个独立的终端，像IDE一样把几个相关的终端窗口放在一个界面中。
而且即使网络中断也不会丢失任何数据，重新连接后可以继续工作，不会感觉出现场环境有任何变化。
仅凭这两点，tmux就可以极大的提高工作的效率和工作连续性。

Ubuntu上安装tmux
```
$ sudo apt-get install tmux
```
Mac上安装tmux
```
$ brew install tmux
```
Windows上安装tmux还是算了，太折腾了，放弃吧。

启动或连接tmux。
```
$ tmux
```
每次在终端运行tmux命令时，tmux命令本身作为一个客户端，会去连接后台服务，如果服务进程不存在，就会创建一个后台服务进程。
所以tmux是以CS的服务模式管理终端的，即使网络中断了，客户端退出了，所有的运行终端都完整的保存在服务进程中。

服务启动后，和普通终端一样，就可以在tmux的窗口中执行各种命令行了。
tmux有三种交互方式，快捷键，命令行，配置文件。

#### tmux快捷键体验
tmux默认的控制键是Ctrl+b，然后再输入命令字符。就可以对tmux进行各种控制。

示例如下，按下Ctrl+b，然后再按引号键，就可以将当前终端窗口分隔成上下两个终端。
```
Ctrl+b "
```
示例如下，按下Ctrl+b，然后再按百分号键，就可以将当前终端窗口分隔成左右两个终端。
```
Ctrl+b %
```
示例如下，按下Ctrl+b，然后再按字母o，就可以在两个终端窗口间来回切换。
```
Ctrl+b o
```

示例如下，按下Ctrl+b，然后再按字母c，就可以创建一个新窗口。
```
Ctrl+b c
```
示例如下，按下Ctrl+b，然后再按字母n，就可以切换到后一个窗口。
```
Ctrl+b n
```
示例如下，按下Ctrl+b，然后再按字母p，就可以切换到前一个窗口。
```
Ctrl+b p
```
示例如下，按下Ctrl+b，然后再按字母d，就会断开当前连接，回到原始的终端窗口，但tmux会话中的所有终端都还在运行。
```
Ctrl+b d
```
断开连接后，在终端中再次执行tmux命令，tmux会重新连接会话，之前的窗口都会原样打开。
```
$ tmux
```

示例如下，按下Ctrl+b，然后再按问号?，就可以查看所有快捷键。可以按方向键上下翻页，按字母q可以退出查看。
```
Ctrl+b ?
```

#### tmux命令行体验
tmux的控制方式除了方便的快捷键，还有丰富的命令行。

如下命令，查看有多少个窗口。
```
$ tmux list-windows
```
如下命令，创建一个新窗口。再调用list-windows，就可以看到新的窗口。
```
$ tmux new-window
```
如下命令，查看有多少个会话。
```
$ tmux list-sessions
```
如下命令，连接最近的会话。
```
$ tmux attach-session
```

除了在外部终端命令行中执行tmux命令，还可以在tmux中执行命令，按Ctrl+b然后按冒号，就进入底行模式，然后就可以输入各种命令，最后按回车执行。
```
Ctrl+b : split-window <Enter>
```
显示时间，按任意键退出。
```
Ctrl+b : clock-mode <Enter>
```
显示状态条
```
Ctrl+b : set status on <Enter>
```
隐藏状态条
```
Ctrl+b : set status on <Enter>
```

#### tmux功能详解
每次运行tmux都会启动一个客户端client，每个客户端client会连接到服务端server下的某一个会话session。
当然多个客户端client可以连接同一个会话session，所以client与session是多对一的关系。

每一个会话session下面可以管理多个窗口window，每个窗口window可以被分隔成多个窗格pane，每个窗格pane就是一个独立的终端teriminal。
所以从server->session->window->pane就形成一个树状的结构，每个叶子节点就是pane，即一个终端。

命令如下，查看客户端列表。
```
$ tmux list-clients
/dev/ttys000: context [158x42 xterm-256color] (utf8)
```
其中"/dev/ttys000"就是客户端设备, "context"是会话名，"158x42"是窗口宽高，utf8是字符集。

命令的详细定义可以查man手册，"man tmux"，可以查看每条命令的详细定义。
```
list-clients (lsc) [-F format] [-t target-session]
```
其中"list-clients"是命令，lsc是命令简写，
"-F format"，可以指定输出内容格式，
"-t target-session"，可以查看指定会话下有多少客户端连接，默认查看所有会话下的客户端。

查看所有会话session
```
list-sessions (ls) [-F format]
```

查看某会话或所有会话下的所有窗口window
```
list-windows (lsw) [-a] [-F format] [-t target-session]
```

查看某窗口或所有窗口下的所有空格pane
```
list-panes (lsp) [-as] [-F format] [-t target-window]
```
除了终端管理，tmux还提供了缓存管理，方便在终端间复制文字。tmux会保存每次复制的内容。
```
list-buffers (lsb) [-F format]
```

查看所有命令行
```
list-commands (lscm) [-F format]
```
查看所有快捷键
```
list-keys (lsk) [-t mode-table] [-T key-table]
```

#### tmux会话管理
新建会话
```
new-session (new) [-AdDEP] [-c start-directory] [-F format] [-n window-name] [-s session-name] [-t target-session] [-x width] [-y height] [command]
```
- "-s session-name" 指定会话名
- "-n window-name" 指定初始窗口的名字
- "-c start-directory" 设定起始目录
- "-d" 会话创建成功后，不自动连接
    - "-x width" 指定窗口宽度
    - "-y width" 指定窗口调试
- "-A" 如果会话存在就自动连接
    - "-D" 自动连接会话时，断开会话与其它客户端连接
- "-t target-session" 创建共享会话，新会话session-name与已存在的target-session共享所有窗口
- "-P format" 指定命令输出的格式
- "command" 窗口创建成功后执行的shell命令


会话是否存在
```
has-session (has) [-t target-session]
```
重命名会话
```
rename-session (rename) [-t target-session] new-name
```
连接会话
```
attach-session (attach) [-dEr] [-c working-directory] [-t target-session]
```
锁定会话
```
lock-session (locks) [-t target-session]
```
删除会话
```
kill-session (killp) [-a] [-t target-pane]
```

#### tmux窗口管理
创建窗口
```
new-window (neww) [-adkP] [-c start-directory] [-F format] [-n window-name] [-t target-window] [command]
```
- -n 指定新窗口的名字
- -t 指定新窗口的位置
    - -a 新创建的窗口在-t指定窗口后面的位置
    - -k -t指定的位置如果存在窗口，则删除此窗口，并插入新的窗口
- -c 指定新窗口的当前目录
- -d 新窗口创建后，不切换为当前窗口
- -P 命令执行成功后，返回字符串
    - -F 返回字符串的格式
- command 窗口启动后执行的命令

分割窗口，将一个窗口分成上下或左右两个窗口
```
split-window (splitw) [-bdfhvP] [-c start-directory] [-F format] [-p percentage|-l size] [-t target-pane] [command]
```
- -t 被分割的窗口
- -h 分割成左右两个窗口
- -v 分割成上下两个窗口
- -b 指定新窗口的位置在左边或是上边
- -f 新窗口高度占满整个窗口
- -p 指定窗口宽度或高度百分比
- -l 指定窗口的宽度或高度
- -P 命令执行成功后返回字符串
    - -F 返回字符串的格式
- -d 新窗口不切换为当前窗口
- -c 新窗口的当前目录
- command 执行的命令

重命令窗口
```
rename-window (renamew) [-t target-window] new-name
```
查找窗口
```
find-window (findw) [-CNT] [-F format] [-t target-window] match-string
```
- -N 从窗口名字中匹配
- -C 从窗口内容中匹配
- -T 从窗口标题中匹配
- match-string 匹配字符串

切换最近使用的窗口
```
last-window (last) [-t target-session]
```
切换下一个的窗口
```
next-window (next) [-a] [-t target-session]
```
切换上一个的窗口
```
previous-window (prev) [-a] [-t target-session]
```

交换两个窗口的位置
```
swap-window (swapw) [-d] [-s src-window] [-t dst-window]
```
移动窗口到指定位置
```
move-window (movew) [-dkr] [-s src-window] [-t dst-window]
```
镜像出一个窗口
```
link-window (linkw) [-dk] [-s src-window] [-t dst-window]
```
删除镜像
```
unlink-window (unlinkw) [-k] [-t target-window]
```
删除窗口
```
kill-window (killw) [-a] [-t target-window]
```
激活窗口
```
respawn-window (respawnw) [-k] [-t target-window] [command]
```

循环移动窗口位置
```
rotate-window (rotatew) [-DU] [-t target-window]
```

窗口切换到下一种布局
```
next-layout (nextl) [-t target-window]
```
窗口切换到上一种布局
```
previous-layout (prevl) [-t target-window]
```

显示所有窗口的序号
```
display-panes (displayp) [-t target-client]
```

切换到上一次选中的面板
```
last-pane (lastp) [-de] [-t target-window]
```

交换两个面板的位置
```
swap-pane (swapp) [-dDU] [-s src-pane] [-t dst-pane]
```
合并面板，把"-t dst-pane"的面板分割成两个， 把-s指定的面板移到新分割的位置中，-t与-s位于相同的窗口中
```
move-pane (movep) [-bdhv] [-p percentage|-l size] [-s src-pane] [-t dst-pane]
```
合并面板，把"-t dst-pane"的面板分割成两个， 把-s指定的面板移到新分割的位置中，-t与-s位于不同的窗口中
```
join-pane (joinp) [-bdhv] [-p percentage|-l size] [-s src-pane] [-t dst-pane]
```
把某个面板，移动到一个新的窗口
```
break-pane (breakp) [-dP] [-F format] [-s src-pane] [-t dst-window]
```
调整大小
```
resize-pane (resizep) [-DLMRUZ] [-x width] [-y height] [-t target-pane] [adjustment]
```
- -Z 占满全屏或恢复原大小
- -U 扩大上边缘位置
- -D 扩大下边缘位置
- -L 扩大左边缘位置
- -R 扩大右边缘位置

```
pipe-pane (pipep) [-o] [-t target-pane] [command]
```

捕获面板内容
```
capture-pane (capturep) [-aCeJpPq] [-b buffer-name] [-E end-line] [-S start-line][-t target-pane]
```
- -e 保留转义格式
- -p 输出内容到标准输出
- -b 输出内容到缓存
- -S start-line 起始行
- -E end-line 结束行
- -t 目标面板

重新激活面板
```
respawn-pane (respawnp) [-k] [-t target-pane] [command]
```

删除一个面板
```
kill-pane (killp) [-a] [-t target-pane]
```

切换当前窗口
```
select-window (selectw) [-lnpT] [-t target-window]
```

切换当前面板
```
select-pane (selectp) [-DdegLlMmRU] [-P style] [-t target-pane]
```
- -d 禁止面板输入
- -e 使能面板输入
- -m 标记面板，被标记的面板作为join-pane swap-pane swap-window命令-s的默认参数
- -M 取消标记
- -g 显示终端属性
- -P style 设置终端属性
- -t 目标面板
- -D -U -L -R 目标面板为-t指定面板的相邻面板

选择窗口布局
```
select-layout (selectl) [-nop] [-t target-window] [layout-name]
```

交互式选择客户端
```
choose-client (capturep) [-aCeJpPq] [-b buffer-name] [-E end-line] [-S start-line][-t target-pane]
```
交互式选择会话
```
choose-session (capturep) [-aCeJpPq] [-b buffer-name] [-E end-line] [-S start-line][-t target-pane]
```
交互式选择窗口
```
choose-window (capturep) [-aCeJpPq] [-b buffer-name] [-E end-line] [-S start-line][-t target-pane]
```
交互式选择会话或窗口
```
choose-tree (capturep) [-aCeJpPq] [-b buffer-name] [-E end-line] [-S start-line][-t target-pane]
```
交互式选择缓存
```
choose-buffer (capturep) [-aCeJpPq] [-b buffer-name] [-E end-line] [-S start-line][-t target-pane]
```

#### tmux缓存管理
进入复制模式
```
copy-mode [-t target-pane]
```
复制模式，用户就可以浏览此窗口的输出历史，并可以复制内容到缓存

清空面板输出的历史记录
```
clear-history (clearhist) [-t target-pane]
```

粘贴缓存内容到指定面板
```
paste-buffer (pasteb) [-dpr] [-s separator] [-b buffer-name] [-t target-pane]
```
设置缓存内容
```
set-buffer (setb) [-a] [-b buffer-name] [-n new-buffer-name] data
```
- -a 保存缓存内容，并追加到后面
- -b 指定缓存的名字
- -n 缓存的新名字
- data 新加内容

显示某条缓存的内容
```
show-buffer (showb) [-b buffer-name]
```
将某条缓存内容保存到文件
```
save-buffer (saveb) [-a] [-b buffer-name] path
```
将文件内容加载到某条缓存
```
load-buffer (loadb) [-b buffer-name] path
```
删除某条缓存
```
delete-buffer (deleteb) [-b buffer-name]
```

#### tmux快捷键与命令行

发送前缀键到面板
```
send-prefix [-t target-pane]
```
发送字符到面板
```
send-keys (send) [-lRM] [-t target-pane] key ...
```
快捷键映射到命令行
```
bind-key (bind) [-cnr] [-t mode-table] [-R repeat-count] [-T key-table] key command [arguments]
```
取消映射
```
unbind-key (unbind) [-acn] [-t mode-table] [-T key-table] key
```

```
clock-mode (clearhist) [-t target-pane]
```

输出信息
```
display-message (display) [-p] [-c target-client] [-F format] [-t target-pane] [message]
```
- -p 作为命令的输出，否则输出到状态行
- -c target-client 指定客户端
- -t target-pane 指定面板
- message 输出的消息

查看输出历史
```
show-messages (showmsgs) [-JT] [-t target-client]
```

执行命令前，让用户选择一下是否执行
```
confirm-before (confirm) [-p prompt] [-t target-client] command
```
执行命令前，让用户输出一些参数
```
command-prompt [-p prompts] [-I inputs] [-t target-client] [template]
```
- -p 提示信息
- -I 参数默认值
- -t 指定客户端
- template 命令模板
```

执行Shell命令
```
run-shell (run) [-b] [-t target-pane] shell-command
```
执行完Shell命令，根据执行结果，选择执行tmux的命令
```
if-shell (if) [-bF] [-t target-pane] shell-command command [command]
```

显示环境变量
```
show-environment (showenv) [-gs] [-t target-session] [name]
```
设置环境变量
```
set-environment (setenv) [-gru] [-t target-session] name [value]
```

显示配置
```
show-options (show) [-gqsvw] [-t target-session|target-window] [option]
```
- -s 服务级配置
- -w 窗口级配置
- -g 显示全局配置
- -v 只显示配置值
- -t 目标会话或窗口
- option 配置项

修改配置
```
set-option (set) [-agosquw] [-t target-window] option [value]
```

加载脚本，
可以将tmux命令写文件中，并随时加载执行，tmux服务在启动时，会默认加载~/.tmux.conf，所以可以将一些默认的配置写此文件中
```
source-file (source) [-q] path
```

同步机制，
tmux提供了一种同步机制，shell命令可以等待tmux条件，tmux可触发条件
```
wait-for (wait) [-L|-S|-U] channel
```

其它命令
```
set-hook (setenv) [-gru] [-t target-session] name [value]
show-hooks (showenv) [-gs] [-t target-session] [name]
```

#### tmux客户端与服务端
刷新客户端
```
refresh-client (refresh) [-S] [-C size] [-t target-client]
```
挂起客户端
```
suspend-client (suspendc) [-t target-client]
```
切换客户端与会话的关系
```
switch-client (switchc) [-Elnpr] [-c target-client] [-t target-session] [-T key-table]
```
断开客户端与会话的连接
```
detach-client (detach) [-P] [-a] [-s target-session] [-t target-client]
```
锁定客户端
```
lock-client (lockc) [-t target-client]
```
查看服务信息
```
server-info (info) 
```
启动服务进程
```
start-server (start) 
```
结束服务进程
```
kill-server
```
锁定所有客户端
```
lock-server (lock) 
```

#### tmux服务配置
常规配置
```
buffer-limit 20
message-limit 100
exit-unattached off
set-clipboard on

default-terminal "screen"
escape-time 500
focus-events off
history-file ""
quiet off
terminal-overrides "xterm*:XT:Ms=\E]52;%p1%s;%p2%s\007:Cs=\E]12;%p1%s\007:Cr=\E]112\007:Ss=\E[%p1%d q:Se=\E[2 q,screen*:XT"
```

#### tmux会话配置
使用set命令，可以修改tmux一些配置，让tmux更加个性化

启用鼠标
```
mouse on
```
设置命令前缀
```
prefix C-s
prefix2 None
```
设置状态栏
```
status on
status-keys vi
status-interval 15
status-justify left
status-position bottom
status-style fg=black,bg=green
status-left "[#S] "
status-left-length 10
status-left-style default
status-right " "#{=21:pane_title}" %H:%M %d-%b-%y"
status-right-length 40
status-right-style default
message-style fg=black,bg=yellow
message-command-style fg=yellow,bg=black
visual-activity off
visual-bell off
visual-silence off
bell-action any
bell-on-alert off
set-titles off
set-titles-string "#S:#I:#W - "#T" #{session_alerts}"
```

窗口显示的配置
```
display-panes-active-colour red
display-panes-colour blue
display-panes-time 5000
display-time 5000
```

启动或结束的窗口时的配置
```
detach-on-destroy on
destroy-unattached off
set-remain-on-exit off
default-command ""
default-shell "/bin/zsh"
lock-after-time 0
lock-command "lock -np"
update-environment "DISPLAY SSH_ASKPASS SSH_AUTH_SOCK SSH_AGENT_PID SSH_CONNECTION WINDOWID XAUTHORITY"
```

常规配置
```
base-index 1
renumber-windows on
history-limit 50000
word-separators " -_@"
key-table "root"
repeat-time 500
assume-paste-time 1
```

#### tmux窗口配置
使用set -w命令或是set-window-option命令，可以设置窗口的一些配置
窗口布局设置
```
force-height 0
force-width 0
main-pane-height 24
main-pane-width 80
other-pane-height 0
other-pane-width 0
```

窗口样式
```
pane-border-style default
pane-border-status off
pane-active-border-style fg=green
pane-border-format "#{?pane_active,#[reverse],}#{pane_index}#[default] "#{pane_title}""

window-style default
window-active-style default

clock-mode-style 24
clock-mode-colour blue
```

状态栏样式
```
window-status-separator " "
window-status-style default
window-status-bell-style reverse
window-status-last-style default
window-status-current-style default
window-status-activity-style reverse
window-status-format "#I:#W#{?window_flags,#{window_flags}, }"
window-status-current-format "#I:#W#{?window_flags,#{window_flags}, }"
```

常规配置
```
mode-keys vi
mode-style fg=black,bg=yellow
allow-rename off
automatic-rename on
automatic-rename-format "#{?pane_in_mode,[tmux],#{pane_current_command}}#{?pane_dead,[dead],}"
pane-base-index 1
```

其它配置
```
aggressive-resize off
alternate-screen on
monitor-activity off
monitor-silence 0
synchronize-panes off

wrap-search on
xterm-keys off
remain-on-exit off
```

### docker使用

程序在运行的过程中，很多都需要依赖外部系统环境，如配置文件、环境变量、缓存文件等。
现在软件往往将这些文件打包成软件包，交付给用户，并部署到生产环境。但这种方式还存在很多问题。

- 测试环境的一些文件或环境变量没有放到软件包中
- 生产环境中程序与程序之间使用相同的资源，产生冲突
- 在测试环境中，测试运行没问题，但一部署到生产环境，就出现了问题。
- 当系统的程序一多，使用时间一长，就会积累下一堆垃圾文件，尤其是Windows和Android。

docker以容器的形式，将程序运行的所有环境，打包成一个独立的镜像，进行交付与部署。
应用程序运行的容器中，有一个独立的根文件系统，并可以监听任意端口，设置环境变量，就好像独自占有一个操作系统。
这样就可以保证程序运行环境的完整性与独立性，测试环境与生产环境不会有差异。
在外部看来就只是一个镜像文件，很方便的添加与删除，也不会在系统残留任何文件。
并且这个容器中的运行环境，可以像git管理代码一样，记录下各个历史版本，并可以随时切换。

与VMware或是VirtualBox之类的虚拟机相比，docker更加轻量，docker中的进程和本机进程一样，docker中的文件和本机文件一样，docker只是将它们组织在一起。
而不像虚拟机一样，需要虚拟出一个完整的操作系统，并提供一堆设备驱动程序，一台普通的电脑开几个虚拟机资源就不足了，但docker却像进程一样占有很少的资源，可以运行很多容器。

docker分为企业版EE，社区版CE，对于个人使用社区版即可。更多信息参考：[docker官网](https://docs.docker.com/)

- [Windows版docker下载](https://store.docker.com/editions/community/docker-ce-desktop-windows)
- [Mac版docker下载](https://store.docker.com/editions/community/docker-ce-desktop-mac)

Ubuntu上安装docker，还需要将docker官网加到软件源中
```
$ curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo apt-key add -
$ sudo add-apt-repository \
    "deb [arch=amd64] https://download.docker.com/linux/ubuntu \
    $(lsb_release -cs) \
    stable"
$ sudo apt-get update
$ sudo apt-get install docker-ce
```

#### 入门体验
如下示例，run命令用busybox镜像，启动了docker的一个容器。
```
$ docker run -it busybox
Unable to find image 'busybox:latest' locally
latest: Pulling from library/busybox
8c5a7da1afbc: Pull complete
Digest: sha256:cb63aa0641a885f54de20f61d152187419e8f6b159ed11a251a09d115f_f9bd
Status: Downloaded newer image for busybox:latest
/ #
```
这里本机并没有busybox的镜像，所以docker自动从dockhub上，下载了busybox的镜像，并用这个镜像启动了一个容器。
像github共享代码一样，dockhub上也共享了各种系统镜像，用户可以自由的下载与上传。

run命令需要一个参数，指定镜像的名称与版本，这里镜像的名字为busybox，默认的版本为latest，即最新版。
busybox是将Unix下的常用命令经过挑选裁剪集成到一个程序中，搭配Linux内核就可以做出一个小型的操作系统，在嵌入式领域应用广泛。
体积小到只有1M左右，下载很快，所以这里用做示例。更多信息参考[busybox官网](https://busybox.net/)

"/ #"，看到最后一行的的命令提示符，就知道容器已经启动了，就可以在这个容器的shell中，执行各种命令与操作了。最后用Ctrl+D或关闭终端窗口，就可以结束容器。
再次运行run，又可以重新启动容器。



如下示例，除了交互式启动容器，还可以用守护的方式启动容器。
```
$ docker run -dt --name demo busybox
d71c8e37bcc153db239f8b1eccb5fa53d202df84d3ffa7ae4e7f8c051d0d481a
```
- "-dt"，指定容器守护的方式运行，即使终端窗口关闭了，守护式的容器会在后台一直运行，可以被反复连接使用。
- "--name demo"，指定了容器的名字为demo，每个容器启动后docker会生成一个sha256的哈希值，如这里的"d71c8e37bcc153db239f8b1eccb5fa53d202df84d3ffa7ae4e7f8c051d0d481a"
在之后的命令中，参数中需要指定容器的地方，都可以这个sha256的哈希值，也可以只写出前几位。但为了方便记忆，可以给容器指定名字。
- "busybox"，指定镜像名字与版本

如下示例，ps命令可以查看正在运行的容器。
```
$ docker ps
CONTAINER ID        IMAGE               COMMAND             CREATED             STATUS              PORTS               NAMES
d71c8e37bcc1        busybox             "sh"                14 seconds ago      Up 13 seconds                           demo
```
如下示例，top命令可以查看某容器中运行的进程
```
$ docker top demo
PID                 USER                TIME                COMMAND
4163                root                0:00                sh
```

如下示例，exec可以在容器中运行各种命令，并将命令输出到当前终端。
```
$ docker exec demo uname
Linux
$ docker exec demo hostname
d71c8e37bcc1
```
如下示例，还可以连接容器，启动一个交互shell。Ctrl+D或是关闭终端窗口，只会结束当前shell，容器依然还在后台继续运行。还可以被反复连接。
```
$ docker exec -it demo sh
/ #
```
如下示例，还可以对容器进行重命名。可以用ps命令查看，新名字已经生效。
```
$ docker rename demo demo1
```

容器中的根文件系统与本机的文件系统是完全隔离的，所以才能提供给容器中应用一个独立的运行环境。

如下示例，可以将本机文件复制到容器中。
```
$ docker cp ~/.vimrc demo1:/root
$ docker exec demo1 ls -a /root
.
..
.vimrc
```
如下示例，可以将容器中文件复制到本机。
```
$ docker cp demo1:/root vimrc
$ ls
vimrc
```

如下示例，停止容器
```
$ docker stop demo1
```

ps命令默认只查看正在运行的容器，如果要查看已经停止的容器可以加参数-a
```
$ docker ps -a
CONTAINER ID        IMAGE               COMMAND             CREATED             STATUS                         PORTS               NAMES
513a5e1fc6f4        busybox             "sh"                4 minutes ago       Exited (137) 3 seconds ago                         demo1
```

如下示例，已经停止的容器，还可以再启动，继续运行。
```
$ docker start demo1
```

如下示例，如果确定已经停止的容器，不会再次启动使用，可以删除掉。
```
$ docker rm demo1
```

#### 镜像管理
```
docker images
image       Manage images

images      List images
search      Search the Docker Hub for images
pull        Pull an image or a repository from a registry
push        Push an image or a repository to a registry
save        Save one or more images to a tar archive (streamed to STDOUT by default)
load        Load an image from a tar archive or STDIN
rmi         Remove one or more images

history     Show the history of an image
build       Build an image from a Dockerfile
tag         Create a tag TARGET_IMAGE that refers to SOURCE_IMAGE
```

#### 容器管理
```
stats       Display a live stream of container(s) resource usage statistics
attach      Attach local standard input, output, and error streams to a running container
container   Manage containers
commit      Create a new image from a container's changes
diff        Inspect changes to files or directories on a container's filesystem
kill        Kill one or more running containers
wait        Block until one or more containers stop, then print their exit codes
create      Create a new container
pause       Pause all processes within one or more containers
unpause     Unpause all processes within one or more containers
export      Export a container's filesystem as a tar archive
import      Import the contents from a tarball to create a filesystem image
logs        Fetch the logs of a container
restart     Restart one or more containers
update      Update configuration of one or more containers
volume      Manage volumes
network     Manage networks
```

#### 系统管理
```
deploy      Deploy a new stack or update an existing stack
version     Show the Docker version information
system      Manage Docker
info        Display system-wide information
login       Log in to a Docker registry
logout      Log out from a Docker registry
inspect     Return low-level information on Docker objects
secret      Manage Docker secrets
trust       Manage trust on Docker images
```

#### 集群管理
```
config      Manage Docker configs
checkpoint  Manage checkpoints
plugin      Manage plugins
service     Manage services
swarm       Manage Swarm
node        Manage Swarm nodes
stack       Manage Docker stacks
```

### git使用
软件的开发是不断的迭代，不断的优化，不断的升级，是一个循序渐进的过程。
所以需要对代码进行版本管理，记录每次提交的代码，可以随时查看变化与切换版本。

软件开发有时需要同时进行多个任务，如在开发新功能时，需要修复线上bug，所以需要分支管理。
一般一个项目中至少会有三种分支：master、feature、bugfix。

此外软件开发的项目都一般是由团队完成，要多人协作共同完成功能开发，所以需要仓库管理，管理多人的代码。

git就是这样一种开源的代码管理工具，可以用来管理代码的版本、分支、仓库等。

Mac上自带git，不需要安装。
Windows上安装了的git-scm，也集成了git，也不需要单独安装。
但Ubuntu需要自己安装一下。
```
$ sudo apt-get install git
```
#### git 基础入门
git help tutorial
git help everyday
git help workflows
git help glossary

git是一个命令集合，有很多子命令，可以通过help命令来查看。
```
$ git help
```

git init
git status
git diff
git add
git mv
git rm
git reset
git commit
git checkout
git blame
git log
git tag

git branch
git merge
git rebase

git config
git clone
git remote
git revert
git fetch
git pull
git push

git bisect
git grep
git show


add                       merge-octopus
add--interactive          merge-one-file
am                        merge-ours
annotate                  merge-recursive
apply                     merge-resolve
archimport                merge-subtree
archive                   merge-tree
bisect                    mergetool
bisect--helper            mktag
blame                     mktree
branch                    mv
bundle                    name-rev
cat-file                  notes
check-attr                p4
check-ignore              pack-objects
check-mailmap             pack-redundant
check-ref-format          pack-refs
checkout                  patch-id
checkout-index            prune
cherry                    prune-packed
cherry-pick               pull
citool                    push
clean                     quiltimport
clone                     read-tree
column                    rebase
commit                    receive-pack
commit-tree               reflog
config                    relink
count-objects             remote
credential                remote-ext
credential-cache          remote-fd
credential-cache--daemon  remote-ftp
credential-osxkeychain    remote-ftps
credential-store          remote-http
cvsexportcommit           remote-https
cvsimport                 remote-testsvn
cvsserver                 repack
daemon                    replace
describe                  request-pull
diff                      rerere
diff-files                reset
diff-index                rev-list
diff-tree                 rev-parse
difftool                  revert
difftool--helper          rm
fast-export               send-email
fast-import               send-pack
fetch                     sh-i18n--envsubst
fetch-pack                shell
filter-branch             shortlog
fmt-merge-msg             show
for-each-ref              show-branch
format-patch              show-index
fsck                      show-ref
fsck-objects              stage
gc                        stash
get-tar-commit-id         status
grep                      stripspace
gui--askpass              submodule
hash-object               submodule--helper
help                      subtree
http-backend              svn
http-fetch                symbolic-ref
http-push                 tag
imap-send                 unpack-file
index-pack                unpack-objects
init                      update-index
init-db                   update-ref
instaweb                  update-server-info
interpret-trailers        upload-archive
log                       upload-pack
ls-files                  var
ls-remote                 verify-commit
ls-tree                   verify-pack
mailinfo                  verify-tag
mailsplit                 web--browse
merge                     whatchanged
merge-base                worktree
merge-file                write-tree
merge-index

### vim入门
Mac上自带vim，不需要安装。
Windows上安装了的git-scm，也集成了vim，也不需要单独安装。
但Ubuntu默认只安装了vi，vim需要自己安装一下。
```
$ sudo apt-get install vim
```
vim是最高效的编辑器，没有之一。熟练掌握它的使用，会极大的提升文本或代码的编辑速度。

当然与常见的其它编辑器不同，vim有独特的操作模式，刚开始使用会有些奇怪。
不过一但你理解了它的运行逻辑，适应了它的操作习惯，你这辈子都不会再想用其它的编辑器。

输入命令vim，如果带有参数则会打开此文件，如果没有参数，则直接打开一个空文件。
```
$ vim hi.txt
```
按字母"i"，进入编辑模式，然后就可以输入任意文本。
```
Hello Vim World!
Vim is best and fast.
You can use vim to input text or code into complute in a free style.
```
输入完内容后，按左上角的"Esc"键，就可以退出编辑模式，回到命令模式。
在命令模式下可以对文件内容，进行各种查看、搜索、修改等操作。vim很多高效的操作都是在命令模式下执行的。

最后，需要保存文件并退出时，按冒号键":"，进入底行模式，再输入"wq"并按回车。
```
:wq
```
#### vim的常用模式
与其它编辑器不同，vim是一种模式化编辑器，即处在不同模式下，每个按键都会不同的功能。
之所以vim是最高效的编辑器，这就是其中的原因之一。

常见的模式有：命令模式、编辑模式、底行模式。
启动vim后，默认的模式是命令模式，其它模式都是以命令模式为基准中心进行相互切换。
即由命令模式切到到编辑模式，由编辑模式切换到命令模式；由命令模式切换到底行模式，由底行模式切换到命令模式。

- 命令模式: 通过各种快捷键，对文件内容进行各种快速的查看、搜索、修改等操作。
- 编辑模式: 和其它编辑器一样，各种字母数字按键会当成文件的内容直接输入。
- 底行模式: 通过输入各种命令行，调用vim各种函数或脚本对文件内容进行复杂的处理。

不同模式下按键的功能是不一样的，所以要知道当前的模式，并能自由切换。如果不知道自己所在的模式，就按\<Esc\>总会回到命令模式的。

- 命令模式->底行模式: 输入冒号键":"，从命令模式切换到底行模式。
- 命令模式<-底行模式: 输入回车键\<Enter\>执行命令，或输入退出键\<Esc\>不执行命令，然后从底行模式，返回到命令模式。
- 命令模式->编辑模式: 输入"i"、"I"、"a"、"A"、"o"、"O"、"s"、"S"，其中任何一个键都可以进行编辑模式。
	- "i": 在光标左边的位置开始插入内容
	- "a": 在光标右边的位置开始插入内容
	- "o": 在光标下面新加一空白并开始插入内容
	- "O": 在光标上面新加一空白并开始插入内容
	- "A": 在光标当前行尾的位置开始插入内容
	- "I": 在光标当前行首的位置开始插入内容
	- "s": 删除光标下的字符，并在光标当前的位置开始插入内容
	- "S": 删除光标当前行的所有字符，并在光标当前行开始插入内容
- 命令模式<-编辑模式: 输入退出键\<Esc\>，从编辑模式返回到命令模式。

#### vim的帮助文档
vim有很灵活的快捷键，大量的配置与命令，和丰富的插件，所以除了要多练习，还需要多看文档。vim自带的帮助文档就很丰富很完整。

英文不好也没关系，vim帮助文档中的英语单词都比较简单，很容易阅读与理解。
当然也有很多复杂的单词，不过那也无所谓，就算每篇文档只看懂一半，也能学到很多有用的技巧。
因为英文的文档，是最接近作者的思维，功能讲解的也更加全面，所以直接查看英文文档是一个优秀程序的必备修养。
所以要克服自己的心理，刚开始看懂多少是多少，慢慢养成直接看英文文档的习惯，文档看的多了，自然看懂的就会越来越多。
阅读英文文档不仅可以提高英文水平，而且对你写文档和写代码的水平也会有很大的帮助。

打开vim，在命令模式下输入":help"并按回车，就可以进入文档中心，查看各种文档。
```
:help
```

***学VIM第一要掌握的四个快捷键就是hjkl***。
把双手自然的放在键盘上，右手离的最近的四个键就是hjkl，所以这么黄金的位置当然分配给了使用最频繁的光标移动了。

- "h"键在最左边，所以按"h"当然就是把光标左移
- "l"键在最右边，所以按"l"当然就是把光标右移
- "j"键在食指下面，是按键最快的，所以按"j"当然就是把光标向下移。
- 最后是"k"键，当然就是把光标向上移。

在帮助文档中，可以练习一下光标移动的这四个键，多体验一下，你就会感觉到vim的强大。
在vim鼠标就是最影响编辑速度的瓶颈，练习过一段时间的快捷键后，再碰鼠标，你就会很明显的感觉到鼠标严重拖慢你的操作速度。
再练习过一段时间后，别说鼠标了，就是那四个方向键和"Esc"键你会懒得去用，因为那太慢太慢了。
最终使用vim的效果就是你的手腕不会有任何移动，只有手指在26个字母和10个数字键上噼里啪啦的在敲击。

在vim有tags标签，就像网页中的超链接一样，可以点击访问另外一个文档。
使用的方式就是，把光标移动到有特殊标记的单词上，按下两个组合键快捷"Ctrl+]"，就可以跳到另外一篇文档中。
看完那篇文档后，按下组合键"Ctrl+T"，就可以返回原来的文档。

另外浏览文件常用的快捷键还有。

- Ctrl+F 向下翻页
- Ctrl+B 向上翻页
- Ctrl+D 向下翻半页
- Ctrl+U 向上翻半页
- gg 跳到文件第一行
- G 跳到文件最后一行

#### vim的文本修改
除了前面的光标移动与翻页功能，命令模式下还有很多操作命令。

- cc 删除当前行，并进入编辑模式
- dd 删除当前行
- yy 复制当前行
- p 在光标之后粘贴复制的内容
- x 删除光标所在的字符
- r 替换光标所在的字符，输入字符r后，接着输入新字符
- u 取消操作
- Ctrl+R 恢复取消的操作

#### vim的单词搜索
除了hjkl，移动光标外，还有更快捷移动光标的方式，就是搜索。

- 行内搜索: "f"与"t"，当用h或l进行左右移动时，经常需要按多好次键才能移动到目标位置，所以输入f加上目标位置所在的字符就可以直接跳过去。t与f相同也是跳转，只是t是跳到目标字符的左边一个位置。
- 匹配搜索: "\*"与"#"，按一下"\*"，vim就会用光标所在的单词，向下搜索，直接跳到当前单词出现的下一位置。#与\*相反，是向上搜索。
- 全文搜索: "?"与"/"，如果相搜索任意单词，在命令模式下输入"/"，然后输入想要搜索的单词，最后输入回车即可向下搜索并移动光标到单词所在位置。?与/相反是向上搜索。

#### vim的命令组合
#### vim的常用配置
除了前面那一堆高效的快捷键外，vim另一强大的原因就是灵活的配置。
你完全可以根据自己的习惯，修改各种各样的配置，让编辑器更加得心应手。

vim的配置命令是set，在命令模式中输入":set "，再加上需要修改的配置。

***显示行号***，输入":set number"，即可在窗口左边显示文件的所有行号，就可以很清楚的知道当前的位置。
```
:set number
```
***显示相对行号***，很多时候目标位置距当前位置相隔很多行，还要去目测或一行行去数相对位置。
设置显示相对行号后，就可以直接看到窗口中所有行相对于当前行的相对行号。
```
:set relativenumber
```

***显示光标纬线***，有时屏幕屏幕太大，往往对不准同一列的字符或同一行的内容，就可以显示经纬经。
```
:set cursorline
```
***显示光标经线***
```
:set cursorcolumn
```
状态行，也可以显示很多用有的信息。
***显示光标位置***
```
:set ruler
```
***显示当前命令***
```
:set showcmd
```
***显示当前模式***
```
:set showmode
```
#### vim的启动脚本
vim有大量的配置与命令，不可能每次启动vim都要手动去输入一遍。
vim具有脚本解析的功能，并且会在启动的时候会加载启动脚本文件。
所以就可以把一些常用的命令放到启动脚本中，每次打开文件时，vim会自动的首先执行这些命令。

vim默认的启动脚本文件是在家目录下的.vimrc，如下打开启动脚本文件，并把之前的那些命令写到脚本中。
因为vim会把脚本文件中的每一行当成一条命令来解析并执行，所以行首不需要再专门输入":"。
```
$ vim ~/.vimrc
set number
set relativenumber
set cursorline
set cursorcolumn
set ruler
set showcmd
set showmode
set cc=80
set nowrap
set scrolloff=3
```
#### vim的扩展插件
除的vim自带的配置与命令，还有大量丰富的插件，可以扩展很多功能。
但大量的插件手动维护太复杂，可以下载一个[vim插件管理器](https://github.com/junegunn/vim-plug)。
执行如下命令，下载plug-vim。
```
$ curl -fLo ~/.vim/autoload/plug.vim --create-dirs \
    https://raw.githubusercontent.com/junegunn/vim-plug/master/plug.vim
```
下载完成后，还需要在启动脚本文件中，加入一些命令启用此插件管理器。
打开~/.vimrc，并添加以下第2行及以后的内容。
```
$ vi ~/.vimrc
call plug#begin()
Plug 'vim-scripts/tComment'
call plug#end()
```
以后如果需要添加新的插件，就可以在"call plug#begin()"与"call plug#end()"之间插入Plug命令。
如安装插件tComment，就插入"Plug 'vim-scripts/tComment'"。
重新加载启动脚本。
```
:source ~/.vimrc
```
执行":PlugInstall"命令。plug-vim就会从github上，下载tComment插件。
```
:PlugInstall
```
重新加载启动脚本。
```
:source ~/.vimrc
```
输入":help tComment"，即可查看此插件的帮助文档。
```
:help tComment
```
***tComment插件***可以对代码进行快速注释或取消注释。在编写代码尤其是调试代码时，经常会遇到需要暂时注释掉一段代码，但稍后又取消掉注释。
tComment通过简单的命令就可以很快的实现此功能，不再需要手动的去插入一堆注释的符号。

- "gcc" 注释或取消注释当前行的代码。


## 进阶指南
### zsh技巧
在Mac上，将zsh设置为默认的shell。
```
$ chsh -s /bin/zsh
```
在Ubuntu上，将zsh设置为默认的shell。
```
$ chsh -s /usr/bin/zsh
```
### tmux技巧
### docker技巧
### git技巧
### vim技巧
#### YouCompleteMe安装
vim只是编辑器，如果需要语法检查与补全功能可以安装插件[YouCompleteMe](https://github.com/Valloric/YouCompleteMe)。
打开.vimrc配置文件，添加插件。
```
$ vim ~/.vimrc
Plugin 'vim-syntastic/syntastic'
Plugin 'Valloric/YouCompleteMe'
```
保存配置文件，重新打开vim，并执行安装命令。
```
$ vim
:PlugInstall
```
由于网络原因可能下载不了，可以手动下载插件。
```
$ git clone https://github.com/Valloric/YouCompleteMe ~/.vim/bundle/YouCompleteMe
$ cd ~/.vim/bundle/YouCompleteMe
$ git submodule update --init --recursive
```
不论是用vundle安装或手动下载，都需要进入插件目录进行编译安装。
```
$ cd ~/.vim/bundle/YouCompleteMe
$ ./install.py --clang-completer --gocompleter
```
Ubuntu上如果没有cmake还需要安装一下。
```
$ sudo apt-get install cmake
```

## 源码解析
### zsh源码解析
[zsh源码](https://github.com/zsh-users/zsh)
### tmux源码解析
[tmux源码](https://github.com/tmux/tmux)
### docker源码解析
[docker源码](https://github.com/docker/docker-ce)
### git源码解析
[git源码](https://github.com/git/git)
### vim源码解析
#### vim源码安装
vim默认不支持python的语法补全，如果需要用到python，可以下载[vim源码](https://github.com/vim/vim)，编译安装。
```
$ sudo apt-get install python
$ sudo apt-get install python-pip
$ sudo apt-get install python-dev
$ sudo apt-get install libncurses5-dev
$ sudo apt-get install build-essential
$ git clone https://github.com/vim/vim.git && cd vim
$ ./configure --with-features=huge\
				  --enable-pythoninterp\
				  --with-python-config-dir=/usr/lib/python2.7/config-x86_64-linux-gnu/\
				  --enable-multibyte\
				  --prefix=/usr/local/vim8/
$ make -j8
$ sudo mkdir /usr/local/vim8
$ sudo make install
```

```
$ sudo apt-get install build-essential
```

