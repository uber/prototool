" Description: run prototool compile

function! ale_linters#proto#prototool_compile#GetCommand(buffer) abort
  return 'prototool compile %s'
endfunction

call ale#linter#Define('proto', {
    \   'name': 'prototool-compile',
    \   'lint_file': 1,
    \   'output_stream': 'stdout',
    \   'executable': 'prototool',
    \   'command': function('ale_linters#proto#prototool_compile#GetCommand'),
    \   'callback': 'ale#handlers#unix#HandleAsError',
    \})
