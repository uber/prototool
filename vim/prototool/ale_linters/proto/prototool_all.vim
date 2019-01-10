" Description: run prototool all

function! ale_linters#proto#prototool_all#GetCommand(buffer) abort
  return 'prototool all --disable-format %s'
endfunction

call ale#linter#Define('proto', {
    \   'name': 'prototool-all',
    \   'lint_file': 1,
    \   'output_stream': 'stdout',
    \   'executable': 'prototool',
    \   'command_callback': 'ale_linters#proto#prototool_all#GetCommand',
    \   'callback': 'ale#handlers#unix#HandleAsError',
    \})
