## 简介

zsh 和bash一样，是一种终端的shell，但提供了更丰富的功能，更快捷的补全。

tmux 是一款高效的终端分屏器，可以在终端把一块屏幕分成多个小窗口，每个窗口都启动一个独立shell。

docker 是一种容器软件，像虚拟机一样为应用软件提供一个完整独立的运行环境，但以一种更加轻量简捷的方式实现。

git 是代码的版本控制软件，用来记录代码各种变化。

vim 是一款强大的编辑器，通过模式化快捷键提升编辑速度，通过灵活的脚本与插件扩展丰富的功能。

使用zsh+tmux+vim的工具链，根据自己的使用习惯进行个性化配置，可以极大的提升编程开发速度。

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
### YouCompleteMe安装
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
### vim源码安装
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
## 基本功能使用
### zsh使用
### tmux使用
### docker使用
### git使用
### vim使用
## 个性化配置
## 源码解析
