# powerline-shell-go

Just like [powerline-shell](https://github.com/milkbikis/powerline-shell), but written in Go - which means it's a lot faster!

This is forked from [bitbucket.org:devsamurais/powerline-shell-go](https://bitbucket.org/devsamurais/powerline-shell-go).

## Instalation

1. Download Go compiler
2. tar -zxvf go1.2.linux-amd64.tar.gz
3. Add following to `.profile.:

    ```
    #!bash
    export PATH=$PATH:/opt/go/bin
    export GOPATH=$HOME/go
    export GOROOT=/opt/go
    ```

5. `go get`
6. `go build`
7. Open ~/.bashrc
8. Add 

    ```
    #!bash

    # powerline-shell-go
    function _update_ps1() {
        export PS1="$(~/bin/powerline-shell-go/powerline-shell-go $?)"
    }
    export PROMPT_COMMAND="_update_ps1"

    ```

9. ???
10. PROFIT