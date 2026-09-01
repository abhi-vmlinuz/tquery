package completions

import _ "embed"

//go:embed tq.bash
var BashCompletion string

//go:embed tq.fish
var FishCompletion string

//go:embed tq.zsh
var ZshCompletion string
