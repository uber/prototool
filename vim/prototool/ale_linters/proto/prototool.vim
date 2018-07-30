" Description: run the prototool linter

call ale#Set('proto_prototool_command', 'all')

function! ale_linters#proto#prototool#GetCommand(buffer) abort
    let l:command = ale#Var(a:buffer, 'proto_prototool_command')
    if l:command is? 'all'
      " Compile the file, then do generation, then lint
      return 'prototool all --disable-format %s'
    elseif l:command is? 'compile'
      " Compile the file only, doing no generation
      return 'prototool compile %s'
    elseif l:command is? 'lint'
      " Compile the file and then lint, doing no generation
      return 'prototool lint %s'
    else
      " Sensible default, would be better if we could return error
      " Is there a way to return error?
      return 'prototool all --disable-format %s'
    endif
endfunction

call ale#linter#Define('proto', {
    \   'name': 'prototool',
    \   'lint_file': 1,
    \   'output_stream': 'stdout',
    \   'executable': 'prototool',
    \   'command_callback': 'ale_linters#proto#prototool#GetCommand',
    \   'callback': 'ale#handlers#unix#HandleAsError',
    \})
