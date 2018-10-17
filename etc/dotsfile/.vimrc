"安装plug-vim {{{
"$ curl -fLo ~/.vim/autoload/plug.vim --create-dirs https://raw.githubusercontent.com/junegunn/vim-plug/master/plug.vim
"
"安装vim各种插件
":PlugInstall
"
"}}}
"加载插件"{{{
call plug#begin()
Plug 'vim-scripts/tComment'
Plug 'gcmt/taboo.vim'
Plug 'vim-scripts/tComment'
Plug 'tpope/vim-fugitive'
Plug 'airblade/vim-gitgutter'
Plug 'vim-airline/vim-airline'
Plug 'vim-airline/vim-airline-themes'
Plug 'ntpeters/vim-better-whitespace'
Plug 'easymotion/vim-easymotion'

Plug 'vim-scripts/taglist.vim'
let g:Tlist_WinWidth=45
let g:Tlist_Exit_OnlyWindow=1
let g:Tlist_Enable_Fold_Column=0
nnoremap <F2> :TlistToggle<CR>

Plug 'scrooloose/nerdtree'
let g:NERDTreeWinPos="right"
nnoremap <F4> :NERDTreeToggle<CR>

Plug 'kien/ctrlp.vim'
let g:ctrlp_cmd='CtrlPBuffer'

Plug 'rking/ag.vim'
nnoremap <C-G> :Ag <C-R>=expand("<cword>")<CR><CR>

Plug 'junegunn/fzf', { 'dir': '~/.fzf', 'do': './install --all' }
nnoremap <C-N> :FZF -q <C-R>=expand("<cword>")<CR><CR>

Plug 'fatih/vim-go'
Plug 'chr4/nginx.vim'
Plug 'othree/html5.vim'
Plug 'godlygeek/tabular'
Plug 'plasticboy/vim-markdown'
Plug 'vim-scripts/python.vim'

Plug 'mbbill/echofunc'
Plug 'vim-syntastic/syntastic'
let g:syntastic_quiet_messages = { "regex": [
            \ "Missing module docstring",
            \ "Missing class docstring",
            \ "Missing method docstring",
            \ "Missing function docstring",
            \ "Wrong continued indentation",
            \ "Line Too Long",
            \ "Invalid variable name",
            \ "Too many instance attributes",
            \ "defined outside __init__",
            \ "Catching too general exception Exception",
            \ "Exactly one space required before assignment",
         \ ] }

Plug 'Valloric/YouCompleteMe'
let g:syntastic_enable_signs = 1
let g:ycm_confirm_extra_conf=0
nnoremap gd :YcmCompleter GoToDeclaration<CR>
nnoremap gD :YcmCompleter GoToReferences<CR>

Plug 'benmills/vimux'
let mapleader=";"
nnoremap <Leader>; :VimuxPromptCommand<CR>
" nnoremap <Leader>j :VimuxZoomRunner<CR>
" nnoremap <Leader>l :VimuxRunLastCommand<CR>
" nnoremap <Leader>vx :VimuxInterruptRunner<CR>

Plug 'vim-scripts/matrix.vim--Yang'
call plug#end()
"}}}
" 基本配置"{{{
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

set tabstop=4
set shiftwidth=4
set cindent
set expandtab
set backspace=indent,eol,start

set showmatch
set matchtime=2
set foldenable
set foldmethod=marker

set hlsearch
set incsearch
set nowrapscan
set smartcase
set ignorecase

set hidden
set autowrite
set encoding=utf-8
set mouse=a

colorscheme elflord
set t_Co=256
"}}}
"映射快捷键"{{{
nnoremap <C-H> <C-W>h
nnoremap <C-J> <C-W>j
nnoremap <C-K> <C-W>k
nnoremap <C-L> <C-W>l
nnoremap <Space> :

nnoremap j gj
nnoremap k gk

nnoremap df :FZF<CR>
inoremap df _
inoremap jk <Esc>
cnoremap jk <CR>
"}}}
" 编程配置{{{
set keywordprg=man\ -a
set splitbelow
set splitright

autocmd BufReadPost * normal `"
autocmd BufNewFile,BufReadPost *.shy set filetype=shy
autocmd BufNewFile,BufReadPost *.shy set commentstring=#%s
autocmd BufNewFile,BufReadPost *.conf set filetype=nginx

command! RR wa | source ~/.vimrc |e
source ~/.vim_local
"}}}
