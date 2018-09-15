## 《终端工具链》简介
终端工具链，就是对编程开发过程中所用到的各种命令行的工具进行高效的组合，不断的提升编程速度与开发效率。

- 在主流的系统中，Ubuntu的命令行最为强大，有丰富的命令行工具，可以很容易组合出自己的终端工具链；
- 其次是MacOSX，命令行也很丰富，再搭配上苹果电脑的硬件与系统，可以组合出流畅的终端工具链；
- 最后是Windows，命令行功能弱的可以忽略，但可以安装一个shell工具[git-scm](https://git-scm.com/downloads)，使用一些基本的命令，如果需要更丰富的命令行工具，可以本地安装虚拟机或是远程连接云主机，使用Ubuntu。

命令行终端，与图形界面不同，是以一种文本化的方式与系统进行交互。
可以很直接、很高效执行各种系统操作，同时各种重复性的操作，都可以很方便的写成程序脚本，和系统命令一样直接调用，不断的提升操作效率。

在终端里，有大量丰富的命令可以使用，不可能全部掌握，一些基本的命令会使用即可。
但在开发流程，想提升编程速度与开发效率，就需要深入理解与熟练掌握这几个工具：zsh、tmux、docker、git、vim。

- **zsh** 和系统默认自带的bash一样，也是一种shell，不断的解析用户或脚本的输入，执行各种命令。但提供了更丰富的特性，如各种补全，命令补全、文件补全、历史补全，可以极大的提升操作效率。
- **tmux** 是一款高效的终端分屏器，可以在终端把一块屏幕分成多个小窗口，每个窗口都启动一个独立shell，这样就可以充分的利用屏幕，同时执行多个命令。
- **docker** 是一种容器软件，像虚拟机一样为应用软件提供一个完整独立的运行环境，但以一种更加轻量简捷的方式实现，极大的简化的软件的部署与分发。
- **git** 是代码的版本控制软件，用来管理代码的每次变化，分支与版本，本地与远程代码仓库，可以实现多人协作开发。
- **vim** 是一款强大的编辑器，通过模式化快捷键提升编辑速度，通过灵活的脚本与插件扩展丰富的功能。

使用zsh+tmux+vim的工具链，根据自己的使用习惯进行个性化配置，就可以极大的提升编程速度与开发效率。

## 入门指南
每个系统上打开终端的方式都不一样，根据自己的系统进行操作。

- 在Ubuntu中，按Ctrl+Alt+T，可以直接打开终端。
- 在Mac中，打开Finder，然后打开，应用->实用工具->终端。
- 在Windows里，先下载一个应用：[git-scm](https://git-scm.com/downloads)，按步骤安装即可，然后打开应用Git bash。

打开终端后，你就打开了一个全新的世界，通过命令行，你就可以自由自在的控制你自己的电脑，并可以与世界成千上万的计算机进行各种交互。

先来体验一下基本的几个命令吧。

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
### zsh使用
### tmux使用
### docker使用
### git使用
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

与其它编辑器不同，vim是一种模式化编辑器，即处在不同模式下，每个按键都会不同的功能。
之所以vim是最高效的编辑器，这就是其中的原因之一。<br/>
常见的模式有：命令模式、编辑模式、底行模式。
启动vim后，默认的模式是命令模式，其它模式都是以命令模式为基准中心进行相互切换。即由命令模式切到到编辑模式，由编辑模式切换到命令模式；由命令模式切换到底行模式，由底行模式切换到命令模式。

- 命令模式: 通过各种快捷键，对文件内容进行各种快速的查看、搜索、修改等操作。
- 编辑模式: 和其它编辑器一样，各种字母数字按键会当成文件的内容直接输入。
- 底行模式: 通过输入各种命令行，调用vim各种函数或脚本对文件内容进行复杂的处理。

## 个性化配置
### zsh安装
Mac上自带zsh，不用安装，但Ubuntu上需要自己安装一下。
```
$ sudo apt-get install zsh
```
在Mac上，将zsh设置为默认的shell。
```
$ chsh -s /bin/zsh
```
在Ubuntu上，将zsh设置为默认的shell。
```
$ chsh -s /usr/bin/zsh
```
原生的zsh不是很好用，可以安装一个[zsh插件管理器](https://github.com/robbyrussell/oh-my-zsh)。
更多信息可以查看[ohmyzsh官网](https://ohmyz.sh/)。
```
$ sh -c "$(curl -fsSL https://raw.github.com/robbyrussell/oh-my-zsh/master/tools/install.sh)"
```
如果在Ubuntu上没有安装curl，可以安装一下。
```
$ sudo apt-get install curl
```
### tmux安装
[tmux源码](https://github.com/tmux/tmux)
Ubuntu上安装
```
$ sudo apt-get install tmux
```
Mac上安装
```
$ brew install tmux
```
如果Mac上没有brew，可以安装一下[Mac包管理器](https://github.com/Homebrew/brew)。更多信息参考[HomeBrew官网](https://brew.sh/)
```
$ ruby -e "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/master/install)"
```
### docker安装
[docker下载](https://www.docker.com/products/docker-desktop)

[docker源码](https://github.com/docker/docker-ce)
### git安装
Mac上自带git，不需要安装，但Ubuntu需要自己安装一下。
```
$ sudo apt-get install git
```
### vim安装
Mac上自带vim，不需要安装，但Ubuntu需要自己安装一下。
```
$ sudo apt-get install vim
```
vim通过丰富的插件，可以扩展很多功能，定制出完全个性化的编辑器。
但大量的插件手动维护太复杂，可以下载一个[vim插件管理器](https://github.com/VundleVim/Vundle.vim)。
```
$ git clone https://github.com/VundleVim/Vundle.vim.git ~/.vim/bundle/Vundle.vim
```
启用vundle插件管理：打开~/.vimrc，并添加以下第2行及以后的内容。
```
$ vi ~/.vimrc
filetype off
set nocompatible
set rtp+=~/.vim/bundle/Vundle.vim
call vundle#begin()
Plugin 'VundleVim/Vundle.vim'
call vundle#end()
filetype plugin indent on
```
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
#### vim源码安装
vim默认不支持python的语法补全，如果需要用到python，可以下载[vim源码](https://github.com/vim/vim)，编译安装。更多信息查看[vim官网](https://www.vim.org/)
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
## 源码解析