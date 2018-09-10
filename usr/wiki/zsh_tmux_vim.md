## 简介

zsh 和bash一样，是一种终端的shell，但提供了更丰富的功能，更快捷的补全。

tmux 是一款高效的终端分屏器，可以在终端把一块屏幕分成多个小窗口，每个窗口都启动一个独立shell。

vim 是一款强大的编辑器，通过模式化快捷键提升编辑速度，通过灵活的脚本与插件扩展丰富的功能。

使用zsh+tmux+vim的工具链，根据自己的使用习惯进行个性化配置，可以极大的提升编程开发速度。

### zsh安装
Mac上自带zsh，不用安装，但Ubuntu上需要自己安装一下。
```
$ sudo apt-get install zsh
```
将zsh设置为默认的shell。
```
$ chsh -s /usr/bin/zsh
```
原生的zsh不是很好用，可以安装一个插件管理器。
```
$ curl -L https://raw.github.com/robbyrussell/oh-my-zsh/master/tools/install.sh | sh
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
如果Mac上没有brew，可以安装一下 [Mac 包管理器 brew](https://brew.sh/)
```
$ ruby -e "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/master/install)"
```
启动tmux
```
$ tmux
```
### vim安装
Mac上自带vim，不需要安装，但Ubuntu需要自己安装一下。
```
$ sudo apt-get install vim
```
vim有很丰富的插件，可以下载一个插件管理器。
```
$ git clone https://github.com/VundleVim/Vundle.vim.git ~/.vim/bundle/Vundle.vim
```
## 基本快捷键
### zsh使用
### tmux使用
### vim使用
## 个性化配置
## 源码解析
Mac上安装pip
```
$ sudo easy_install pip
$ sudo pip install termtosvg
$ brew install ttygif

$ sudo apt-get install software-properties-common
$ sudo apt-add-repository ppa:zanchey/asciinema
$ sudo apt-get update
$ sudo apt-get install asciinema
$ sudo apt-get install python3-pip
$ sudo pip install TermRecord

```
[终端录制](https://asciinema.org/)
