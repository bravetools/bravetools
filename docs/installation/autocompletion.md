---
layout: default
description: Instructions for installing Bravetools on macOS
keywords: autocomplete, cli, installation, install, shell
title: Enabling autocompletion
parent: Install Bravetools
nav_order: 4
---

# Autocompletion

bravetools supports autocompletion of commands and unit/image names through the `brave completion` command.

```
brave completion

Generate the autocompletion script for brave for the specified shell.
See each sub-command's help for details on how to use the generated script.

Usage:
  brave completion [command]

Available Commands:
  bash        Generate the autocompletion script for bash
  fish        Generate the autocompletion script for fish
  powershell  Generate the autocompletion script for powershell
  zsh         Generate the autocompletion script for zsh

Flags:
  -h, --help   help for completion
```


The following shells are currently supported:
- [bash](#bash)
- [fish](#fish)
- [powershell](#powershell)
- [zsh](#zsh)

## bash
This depends on the 'bash-completion' package.
If it is not installed already, you can install it via your OS's package manager.

To load completions in your current shell session:

        source <(brave completion bash)

To load completions for every new session, execute once:

Linux:

        brave completion bash > /etc/bash_completion.d/brave

macOS:

        brave completion bash > $(brew --prefix)/etc/bash_completion.d/brave

You will need to start a new shell for this setup to take effect.

## fish
To load completions in your current shell session:

        brave completion fish | source

To load completions for every new session, execute once:

        brave completion fish > ~/.config/fish/completions/brave.fish

You will need to start a new shell for this setup to take effect.

## powershell
To load completions in your current shell session:

        brave completion powershell | Out-String | Invoke-Expression

To load completions for every new session, add the output of the above command
to your powershell profile.

## zsh

If shell completion is not already enabled in your environment you will need
to enable it.  You can execute the following once:

        echo "autoload -U compinit; compinit" >> ~/.zshrc

To load completions in your current shell session:

        source <(brave completion zsh); compdef _brave brave

To load completions for every new session, execute once:

Linux:

        brave completion zsh > "${fpath[1]}/_brave"

macOS:

        brave completion zsh > $(brew --prefix)/share/zsh/site-functions/_brave

You will need to start a new shell for this setup to take effect.
