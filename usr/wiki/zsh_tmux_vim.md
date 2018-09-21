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
Windows上安装tmux还是算了，太折腾了，放弃吧，兄弟。

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
重命令窗口
```
rename-window (renamew) [-t target-window] new-name
```
查找窗口
```
find-window (findw) [-CNT] [-F format] [-t target-window] match-string
```

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

分隔窗口
```
split-window (splitw) [-bdfhvP] [-c start-directory] [-F format] [-p percentage|-l size] [-t target-pane] [command]
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

```
last-pane (lastp) [-de] [-t target-window]
swap-pane (swapp) [-dDU] [-s src-pane] [-t dst-pane]
move-pane (movep) [-bdhv] [-p percentage|-l size] [-s src-pane] [-t dst-pane]
join-pane (joinp) [-bdhv] [-p percentage|-l size] [-s src-pane] [-t dst-pane]
kill-pane (killp) [-a] [-t target-pane]
resize-pane (resizep) [-DLMRUZ] [-x width] [-y height] [-t target-pane] [adjustment]
respawn-pane (respawnp) [-k] [-t target-pane] [command]

break-pane (breakp) [-dP] [-F format] [-s src-pane] [-t dst-window]
capture-pane (capturep) [-aCeJpPq] [-b buffer-name] [-E end-line] [-S start-line][-t target-pane]
pipe-pane (pipep) [-o] [-t target-pane] [command]
```

```
choose-client (capturep) [-aCeJpPq] [-b buffer-name] [-E end-line] [-S start-line][-t target-pane]
choose-session (capturep) [-aCeJpPq] [-b buffer-name] [-E end-line] [-S start-line][-t target-pane]
choose-window (capturep) [-aCeJpPq] [-b buffer-name] [-E end-line] [-S start-line][-t target-pane]
choose-buffer (capturep) [-aCeJpPq] [-b buffer-name] [-E end-line] [-S start-line][-t target-pane]
choose-tree (capturep) [-aCeJpPq] [-b buffer-name] [-E end-line] [-S start-line][-t target-pane]
select-layout (selectl) [-nop] [-t target-window] [layout-name]
select-window (selectw) [-lnpT] [-t target-window]
select-pane (selectp) [-DdegLlMmRU] [-P style] [-t target-pane]
```

#### tmux缓存管理
```
show-buffer (showb) [-b buffer-name]
load-buffer (loadb) [-b buffer-name] path
save-buffer (saveb) [-a] [-b buffer-name] path
paste-buffer (pasteb) [-dpr] [-s separator] [-b buffer-name] [-t target-pane]
set-buffer (setb) [-a] [-b buffer-name] [-n new-buffer-name] data
delete-buffer (deleteb) [-b buffer-name]
clear-history (clearhist) [-t target-pane]
copy-mode (confirm) [-p prompt] [-t target-client] command
```

#### tmux快捷键与命令行
```
bind-key (bind) [-cnr] [-t mode-table] [-R repeat-count] [-T key-table] key command [arguments]
unbind-key (unbind) [-acn] [-t mode-table] [-T key-table] key
send-keys (send) [-lRM] [-t target-pane] key ...
send-prefix (send) [-lRM] [-t target-pane] key ...
clock-mode (clearhist) [-t target-pane]
```

```
display-message (display) [-p] [-c target-client] [-F format] [-t target-pane] [message]
show-messages (showmsgs) [-JT] [-t target-client]
show-options (show) [-gqsvw] [-t target-session|target-window] [option]
set-option (set) [-agosquw] [-t target-window] option [value]
source-file (source) [-q] path
confirm-before (confirm) [-p prompt] [-t target-client] command
command-prompt (clearhist) [-t target-pane]
show-environment (showenv) [-gs] [-t target-session] [name]
set-environment (setenv) [-gru] [-t target-session] name [value]
run-shell (run) [-b] [-t target-pane] shell-command
if-shell (if) [-bF] [-t target-pane] shell-command command [command]
wait-for (wait) [-L|-S|-U] channel
set-hook (setenv) [-gru] [-t target-session] name [value]
show-hooks (showenv) [-gs] [-t target-session] [name]
```

#### tmux客户端与服务端
```
refresh-client (refresh) [-S] [-C size] [-t target-client]
suspend-client (suspendc) [-t target-client]
switch-client (switchc) [-Elnpr] [-c target-client] [-t target-session] [-T key-table]
detach-client (detach) [-P] [-a] [-s target-session] [-t target-client]
lock-client (lockc) [-t target-client]
```
```
server-info (info) 
start-server (start) 
kill-server (killp) [-a] [-t target-pane]
lock-server (lock) 
```

#### tmux服务配置
```
buffer-limit 20
default-terminal "screen"
escape-time 500
exit-unattached off
focus-events off
history-file ""
message-limit 100
quiet off
set-clipboard on
terminal-overrides "xterm*:XT:Ms=\E]52;%p1%s;%p2%s\007:Cs=\E]12;%p1%s\007:Cr=\E]112\007:Ss=\E[%p1%d q:Se=\E[2 q,screen*:XT"
```

#### tmux会话配置
```
mouse on
prefix C-s
prefix2 None

base-index 1
renumber-windows on
history-limit 50000

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
word-separators " -_@"

display-panes-active-colour red
display-panes-colour blue
display-panes-time 5000
display-time 5000
repeat-time 500
assume-paste-time 1
key-table "root"

destroy-unattached off
detach-on-destroy on
set-remain-on-exit off
default-command ""
default-shell "/bin/zsh"
lock-after-time 0
lock-command "lock -np"
update-environment "DISPLAY SSH_ASKPASS SSH_AUTH_SOCK SSH_AGENT_PID SSH_CONNECTION WINDOWID XAUTHORITY"
```

#### tmux窗口配置
```
set-window-option (setw) [-agoqu] [-t target-window] option [value]
show-window-options (showw) [-gv] [-t target-window] [option]
```

```
aggressive-resize off
alternate-screen on
monitor-activity off
monitor-silence 0
synchronize-panes off

wrap-search on
xterm-keys off
remain-on-exit off

mode-keys vi
mode-style fg=black,bg=yellow
allow-rename off
automatic-rename on
automatic-rename-format "#{?pane_in_mode,[tmux],#{pane_current_command}}#{?pane_dead,[dead],}"
pane-base-index 1


force-height 0
force-width 0
main-pane-height 24
main-pane-width 80
other-pane-height 0
other-pane-width 0

clock-mode-style 24
clock-mode-colour blue

pane-border-style default
pane-border-status off
pane-active-border-style fg=green
pane-border-format "#{?pane_active,#[reverse],}#{pane_index}#[default] "#{pane_title}""

window-style default
window-active-style default


window-status-separator " "
window-status-style default
window-status-bell-style reverse
window-status-last-style default
window-status-current-style default
window-status-activity-style reverse
window-status-format "#I:#W#{?window_flags,#{window_flags}, }"
window-status-current-format "#I:#W#{?window_flags,#{window_flags}, }"
```

### docker使用

- [Windows版docker下载](https://store.docker.com/editions/community/docker-ce-desktop-windows)
- [Mac版docker下载](https://store.docker.com/editions/community/docker-ce-desktop-mac)

#### docker镜像管理

- 查看镜像 docker image ls
- 下载镜像 docker image pull

刚安装docker后，查看镜像列表，如下为空。
```
$ docker image ls
REPOSITORY          TAG                 IMAGE ID            CREATED             SIZE
```
像github管理代码仓库一样，docker hub上也存放了很多镜像，用户可以自由的下载与上传各种镜像。
如下示例，下载一个busybox镜像。busybox是将Unix下的常用命令经过挑选裁剪集成一个程序中，搭配Linux内核就可以做出一个小型的操作系统，在嵌入式领域应用广泛。
体积很小不到1M，下载很快，所以这里用做示例。更多信息参考[busybox官网](https://busybox.net/)
```
$ docker image pull busybox
```
下载完成后，再查看镜像列表，就会看到busybox镜像相关的信息。
```
$ docker image ls
REPOSITORY          TAG                 IMAGE ID            CREATED             SIZE
busybox             latest              e1ddd7948a1c        6 weeks ago         1.16MB
```
#### docker容器管理

- 查看容器 docker ps
- 启动容器 docker run
- 停止容器 docker exec
- 停止容器 docker stop

如下示例，查看容器列表，因为还没启动任何容器，所以这里为空。
```
$ docker ps
CONTAINER ID        IMAGE               COMMAND             CREATED             STATUS              PORTS               NAMES
```
如下示例，用busybox:latest镜像，启动一个容器，并调用sh命令。
```
$ docker run --name demo -dt busybox:latest sh  
29ff6b8343c4a2c57eab297e74e62422ab9bbd481d69f5ebf108f4aa23ae835c
```
其中，-d 指用守护的方式启动，与交互式 -i 不同，守护式启动，容器可以一直运行，不会因为终端容器关闭而停止。
--name参数，指定容器的名字为demo，docker中标识容器有两种方式，一是通过ID查找容器，二是通过NAMES查找容器，为了方便记忆与查找，建议启动容器时加上名字参数。

如下示例，再次查看容器列表，看到容器已经启动。
```
$ docker ps
CONTAINER ID        IMAGE               COMMAND             CREATED             STATUS              PORTS               NAMES
29ff6b8343c4        busybox:latest      "sh"                4 minutes ago       Up 4 minutes                            demo
```

连接容器demo，调用命令解析器sh。这样就连接上了容器的命令行，可以执行各种命令。
```
$ docker exec -it demo sh
#
```

容器的停止，退出连接后，容器依然在后台运行，可以反复被连接。如果想停止容器的运行就用stop命令。
```
$ docker stop demo
```

#### 挂载文件
之前启动的容器都是与本机之间没有什么交互，是一个完全独立的运行环境。
如果需要容器与本机交互一些文件，就可以在启动容器时指定文件参数。
```
$ docker run --name demo1 -v/Users/shaoying:/home/shaoying  -dt busybox:latest sh
```

#### 端口映射

### git入门
Mac上自带git，不需要安装。
Windows上安装了的git-scm，也集成了git，也不需要单独安装。
但Ubuntu需要自己安装一下。
```
$ sudo apt-get install git
```
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
$ ./install.py --clang-completer
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

