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

function! PrototoolCreateEnable() abort
    silent! let g:prototool_create_enable = 1
endfunction

function! PrototoolCreateDisable() abort
    silent! unlet g:prototool_create_enable
endfunction

call PrototoolCreateEnable()

function! PrototoolCreateToggle() abort
    if exists('g:prototool_create_enable')
        call PrototoolCreateDisable()
        execute 'echo "prototool create DISABLED"'
    else
        call PrototoolCreateEnable()
        execute 'echo "prototool create ENABLED"'
    endif
endfunction

function! PrototoolCreate() abort
    if exists('g:prototool_create_enable')
        silent! execute '!prototool create %'
        silent! edit
    endif
endfunction

autocmd BufNewFile *.proto :call PrototoolCreate()
