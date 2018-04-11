function! PrototoolFormatEnable() abort
    silent! let g:prototool_format_enable = 1
endfunction

function! PrototoolFormatDisable() abort
    silent! unlet g:prototool_format_enable
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

function! PrototoolFormat() abort
    if exists('g:prototool_format_enable')
        silent! execute '!prototool format -w %'
        silent! edit
    endif
endfunction

autocmd BufEnter,BufWritePost *.proto :call PrototoolFormat()
