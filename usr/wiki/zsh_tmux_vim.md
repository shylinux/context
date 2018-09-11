## 简介

zsh 和bash一样，是一种终端的shell，但提供了更丰富的功能，更快捷的补全。

tmux 是一款高效的终端分屏器，可以在终端把一块屏幕分成多个小窗口，每个窗口都启动一个独立shell。

vim 是一款强大的编辑器，通过模式化快捷键提升编辑速度，通过灵活的脚本与插件扩展丰富的功能。

使用zsh+tmux+vim的工具链，根据自己的使用习惯进行个性化配置，可以极大的提升编程开发速度。

相关链接

- Mac包管理器: <https://brew.sh/>

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
如果Mac上没有brew，可以安装一下.
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
vim通过丰富的插件，可以扩展很多功能，定制出完全个性化的编辑器。
但大量的插件手动维护太复杂，可以下载一个[vim插件管理器vundle](https://github.com/VundleVim/Vundle.vim)。
```
$ git clone https://github.com/VundleVim/Vundle.vim.git ~/.vim/bundle/Vundle.vim
```
启用vundle插件管理。
```
$ vi ~/.vimrc
filetype off
set nocompatible
set rtp+=~/.vim/bundle/vundle/
call vundle#begin()
Plugin 'VundleVim/Vundle.vim'
call vundle#end()
filetype plugin indent on
```
### vim源码安装
参考博客: [vim源码安装](https://www.jianshu.com/p/3e606e31da5f)
```
$ sudo apt-get install python-dev
$ sudo apt-get install python3-dev
$ sudo apt-get install libncurses5-dev
$ git clone git@github.com:vim/vim.git && cd vim
$ sudo mkdir /usr/local/vim8
$ ./configure --with-features=huge\
				  --enable-pythoninterp\
				  --enable-python3interp\
				  --with-python-config-dir=/usr/lib/python2.7/config-x86_64-linux-gnu/\
				  --with-python3-config-dir=/usr/lib/python3.5/config-3.5m-x86_64-linux-gnu/\
				  --enable-luainterp\
				  --enable-perlinterp\
				  --enable-rubyinterp\
				  --enable-multibyte\
				  --prefix=/usr/local/vim8/
$ make
$ sudo make install
```
### Vundle安装
### YouCompleteMe安装
参考博客: [YouCompete安装](http://www.10tiao.com/html/263/201610/2652564254/1.html)
打开.vimrc配置文件，添加插件。
```
$ vim ~/.vimrc
Bundle 'vim-syntastic/syntastic'
Bundle 'Valloric/YouCompleteMe'
```
保存并关闭，重新打开vim，执行插件安装命令。
```
$ vim
:BundleInstall
```
插件安装成功后，进入目录进行编译。
```
$ sudo apt-get install pylint
$ sudo apt-get install cmake
$ cd ~/.vim/bundle/YouCompleteMe
$ ./install.py --clang-completer
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
<video id="video" poster="/static/public/player/playerbg.png" width="100%" height="auto" preload="metadata" controls="" src="blob:http://99vbkc.com/5aa889fc-0af4-4fdc-ac25-46d456b70028"></video>
- 终端录制: <https://asciinema.org/>
