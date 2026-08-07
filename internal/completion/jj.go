// Package completion generates shell integration for jj aliases.
package completion

import (
	"fmt"
	"io"
)

// WriteJJBridge writes a shell bridge that delegates `jj stack` completion to bb.
// It must be sourced after the normal jj and bb completion scripts.
func WriteJJBridge(out io.Writer, shell string) error {
	var script string
	switch shell {
	case "bash":
		script = bashBridge
	case "zsh":
		script = zshBridge
	case "fish":
		script = fishBridge
	default:
		return fmt.Errorf("jj alias completion supports bash, zsh, and fish; got %q", shell)
	}
	_, err := io.WriteString(out, script)
	return err
}

const bashBridge = `# Source after: source <(jj util completion bash); source <(bb completion bash)
if declare -F _jj >/dev/null && declare -F __start_bb >/dev/null; then
  eval "$(declare -f _jj | sed '1s/^_jj /_jj_blackbelt_original /')"
  _jj() {
    if [[ ${COMP_WORDS[1]} == stack ]]; then
      local -a _bb_saved_words=("${COMP_WORDS[@]}")
      local _bb_saved_cword=$COMP_CWORD
      COMP_WORDS=(bb stack "${COMP_WORDS[@]:2}")
      __start_bb
      COMP_WORDS=("${_bb_saved_words[@]}")
      COMP_CWORD=$_bb_saved_cword
    else
      _jj_blackbelt_original "$@"
    fi
  }
  complete -F _jj jj
fi
`

const zshBridge = `# Source after: source <(jj util completion zsh); source <(bb completion zsh)
if (( $+functions[_jj] && $+functions[_bb] )); then
  functions[_jj_blackbelt_original]=$functions[_jj]
  _jj() {
    if (( CURRENT >= 2 )) && [[ ${words[2]} == stack ]]; then
      local -a _bb_saved_words=("${words[@]}")
      local _bb_saved_current=$CURRENT
      words=(bb stack "${_bb_saved_words[3,-1]}")
      CURRENT=$_bb_saved_current
      _bb
      words=("${_bb_saved_words[@]}")
      CURRENT=$_bb_saved_current
    else
      _jj_blackbelt_original "$@"
    fi
  }
  compdef _jj jj
fi
`

const fishBridge = `# Source after: jj util completion fish | source; bb completion fish | source
function __bb_complete_jj_stack
    set -l tokens (commandline -opc)
    set -e tokens[1..2]
    bb __complete stack $tokens (commandline -ct) 2>/dev/null | string match -rv '^:'
end
complete -c jj -n 'test (count (commandline -opc)) -ge 2; and test (commandline -opc)[2] = stack' -f -a '(__bb_complete_jj_stack)'
`
