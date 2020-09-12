" Description: run prototool lint

function! ale_linters#proto#prototool_lint#GetCommand(buffer) abort
  return 'prototool lint %s'
endfunction

call ale#linter#Define('proto', {
    \   'name': 'prototool-lint',
    \   'lint_file': 1,
    \   'output_stream': 'stdout',
    \   'executable': 'prototool',
    \   'command': function('ale_linters#proto#prototool_lint#GetCommand'),
    \   'callback': 'ale#handlers#unix#HandleAsError',
    \})
