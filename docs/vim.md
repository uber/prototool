# Vim Integration

This repository is a self-contained plugin for use with the [ALE Lint Engine](https://github.com/w0rp/ale). It should be similarly easy to add support for Syntastic, Neomake, etc.

The Vim integration will currently compile, provide lint errors, do generation of your stubs, and format your files on save. It will also optionally create new files from a template when opened.

The plugin is under [vim/prototool](../vim/prototool), so your plugin manager needs to point there instead of the base of this repository. Assuming you are using [vim-plug](https://github.com/junegunn/vim-plug), copy/paste the following into your vimrc and you should be good to go. If you are using [Vundle](https://github.com/VundleVim/Vundle.vim), just replace `Plug` with `Vundle` below.

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

" ,f will toggle formatting on and off.
" Change to PrototoolFormatFixToggle to toggle with --fix instead.
nnoremap <silent> <leader>f :call PrototoolFormatToggle()<CR>
" ,c will toggle create on and off.
nnoremap <silent> <leader>c :call PrototoolCreateToggle()<CR>

" Uncomment this to enable formatting by default.
"call PrototoolFormatEnable()
" Uncomment this to enable formatting with --fix by default.
"call PrototoolFormatFixEnable()
" Uncomment this to disable creating Protobuf files from a template by default.
"call PrototoolCreateDisable()
```

The recommended setup in short:

```vim
Plug 'w0rp/ale'
Plug 'uber/prototool', { 'rtp':'vim/prototool' }
let g:ale_linters = {
\   'go': ['golint'],
\   'proto': ['prototool-lint'],
\}
let g:ale_lint_on_text_changed = 'never'
call PrototoolFormatFixEnable()
```
