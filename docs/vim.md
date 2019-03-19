# Vim Integration

[Back to README.md](README.md)

This repository is a self-contained plugin for use with the
[ALE Lint Engine](https://github.com/w0rp/ale).

Prototool is also enabled as a maker within
[Neomake](https://github.com/neomake/neomake/blob/master/autoload/neomake/makers/ft/proto.vim).

The Vim integration will currently compile, provide lint errors, do generation of your stubs, and
format your files on save. It will also optionally create new files from a template when opened.

The plugin is under [vim/prototool](../vim/prototool), so your plugin manager needs to point there
instead of the base of this repository. Assuming you are using
[vim-plug](https://github.com/junegunn/vim-plug), copy/paste the following into your `.vimrc` and
you should be good to go. If you are using [Vundle](https://github.com/VundleVim/Vundle.vim), just
replace `Plug` with `Vundle` below.

```vim
Plug 'w0rp/ale'
Plug 'uber/prototool', { 'rtp':'vim/prototool' }
let g:ale_linters = {
\   'go': ['golint'],
\   'proto': ['prototool-lint'],
\}
let g:ale_lint_on_text_changed = 'never'
" <leader>f will format and fix your current file.
" Change to PrototoolFormat to only format and not fix.
nnoremap <silent> <leader>f :call PrototoolFormatFix()<CR>
```

A longer explanation:

```vim
" Prototool must be installed as a binary for the Vim integration to work.

" Add ale and prototool with your package manager.
" Note that Plug downloads from dev by default. There may be minor changes
" to the Vim integration on dev between releases, but this won't be common.
" To make sure you are on the same branch as your Prototool install, set
" the branch field in the options for uber/prototool per the vim-plug
" documentation. Vundle does not allow setting branches, so on Vundle,
" go into plug directory and checkout the branch of the release you are on.
Plug 'w0rp/ale'
Plug 'uber/prototool', { 'rtp':'vim/prototool' }

" We recommend setting just this for Golang, as well as the necessary set for proto.
" Note the 'prototool' linter is still available, but deprecated in favor of individual linters.
" Use the 'prototool-compile' linter to just compile, 'prototool-lint' to compile and lint,
" 'prototool-all' to compile, do generation of your stubs, and then lint.
let g:ale_linters = {
\   'go': ['golint'],
\   'proto': ['prototool-lint'],
\}
" We recommend you set this.
let g:ale_lint_on_text_changed = 'never'

" We generally have <leader> mapped to ",", uncomment this to set leader.
"let mapleader=","

" ,f will format and fix your current file.
" Change to PrototoolFormat to only format and not fix.
nnoremap <silent> <leader>f :call PrototoolFormatFix()<CR>
" ,e will toggle formatting and fixing on and off.
" Change to PrototoolFormatToggle to toggle with only format and not fix instead.
nnoremap <silent> <leader>e :call PrototoolFormatToggle()<CR>
" ,c will toggle create on and off.
nnoremap <silent> <leader>c :call PrototoolCreateToggle()<CR>

" Uncomment this to enable formatting and fixing by default.
" Change to PrototoolFormatEnable to format and not fix by default.
"call PrototoolFormatFixEnable()
" Uncomment this to disable creating Protobuf files from a template by default.
"call PrototoolCreateDisable()
```
