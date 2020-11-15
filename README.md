# WSL-SSH-Agent-Proxy

Windows 10 now has a [built in SSH agent] available; however, it does not work
correctly within WSL2 VMs.  This is a proxy on the WSL2 side to forward all
requests to the Windows SSH agent.

[built in SSH agent]: https://docs.microsoft.com/en-us/windows-server/administration/openssh/openssh_keymanagement

## Usage

1. Start the SSH agent:

    ```powershell
    # Set the SSH agent to start automatically; also consider going through
    # the service UI and setting it to delayed start instead.
    Get-Service -Name ssh-agent | Set-Service -StartupType Automatic
    # Start it for this session
    Start-Service ssh-agent
    ```

2. [Download] the executable and save it somewhere.

    [Download]: https://github.com/mook/wsl-ssh-agent-proxy/releases/latest

3. Set it to automatically start by appending to your `~/.bashrc`:

    ```bash
    export SSH_AUTH_SOCK=~/.ssh/agent.sock # or anywhere else
    setsid ssh-agent-proxy &
    ```

## Related Projects

- [rupor-github/wsl-ssh-agent] sounds awesome, but it seems like it [doesn't
  work with WSL 2].
- This is basically a packaged version of [anaisbetts/ssh-agent-relay] and other
  [npiperelay] based solutions that does not require installing a separate
  `socat` within the WSL environment.

[rupor-github/wsl-ssh-agent]: https://github.com/rupor-github/wsl-ssh-agent
[doesn't work with WSL 2]: https://github.com/microsoft/WSL/issues/5961
[anaisbetts/ssh-agent-relay]: https://github.com/anaisbetts/ssh-agent-relay
[npiperelay]: https://github.com/jstarks/npiperelay
