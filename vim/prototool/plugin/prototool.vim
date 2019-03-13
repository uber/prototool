" For use in your .vimrc
" nnoremap <silent> <leader>f :call PrototoolFormat()<CR>
function! PrototoolFormat() abort
    silent! execute '!prototool format -w %'
    silent! edit
endfunction

" For use in your .vimrc
" nnoremap <silent> <leader>f :call PrototoolFormatFix()<CR>
function! PrototoolFormatFix() abort
    silent! execute '!prototool format --fix -w %'
    silent! edit
endfunction

" auto functions

function! PrototoolFormatEnable() abort
    silent! let g:prototool_format_enable = 1
    silent! let g:prototool_format_fix_flag = ''
endfunction

function! PrototoolFormatDisable() abort
    silent! unlet g:prototool_format_enable
    silent! let g:prototool_format_fix_flag = ''
endfunction

function! PrototoolFormatFixEnable() abort
    silent! let g:prototool_format_enable = 1
    silent! let g:prototool_format_fix_flag = '--fix '
endfunction

function! PrototoolFormatToggle() abort
    if exists('g:prototool_format_enable')
        call PrototoolFormatDisable()
        execute 'echo "prototool format DISABLED"'
    else
        call PrototoolFormatEnable()
        execute 'echo "prototool format ENABLED"'
    endif
endfunction

function! PrototoolFormatFixToggle() abort
    if exists('g:prototool_format_enable')
        call PrototoolFormatDisable()
        execute 'echo "prototool format DISABLED"'
    else
        call PrototoolFormatFixEnable()
        execute 'echo "prototool format --fix ENABLED"'
    endif
endfunction

function! PrototoolFormatOnSave() abort
    if exists('g:prototool_format_enable')
        silent! execute '!prototool format ' . g:prototool_format_fix_flag . '-w %'
        silent! edit
    endif
endfunction

function! PrototoolCreateEnable() abort
    silent! let g:prototool_create_enable = 1
endfunction

function! PrototoolCreateDisable() abort
    silent! unlet g:prototool_create_enable
endfunction

function! PrototoolCreateToggle() abort
    if exists('g:prototool_create_enable')
        call PrototoolCreateDisable()
        execute 'echo "prototool create DISABLED"'
    else
        call PrototoolCreateEnable()
        execute 'echo "prototool create ENABLED"'
    endif
endfunction

function! PrototoolCreateOnSave() abort
    if exists('g:prototool_create_enable')
        silent! execute '!prototool create %'
        silent! edit
    endif
endfunction

function! PrototoolCreateReadPostOnSave() abort
    if exists('g:prototool_create_enable')
      if line('$') == 1 && getline(1) == ''
        silent! execute '!prototool create %'
        silent! edit
      endif
    endif
endfunction

" default functionality below

let g:prototool_format_fix_flag = '--fix '
call PrototoolFormatDisable()
call PrototoolCreateEnable()

autocmd BufEnter,BufWritePost *.proto :call PrototoolFormatOnSave()
autocmd BufNewFile *.proto :call PrototoolCreateOnSave()
autocmd BufReadPost *.proto :call PrototoolCreateReadPostOnSave()
