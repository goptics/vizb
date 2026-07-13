# Dev container toolchain for vizb.
# Node 22 is added via the node feature in devcontainer.json.
#
# Keep the base image user `vscode` (UID 1000). Renaming it breaks Zed/VS Code
# feature installs and `docker exec -u …` (passwd entry missing). The cool
# project identity is the container *name* ("vizber") and shell prompt, not the
# OS login name.
FROM mcr.microsoft.com/devcontainers/go:1.26-bookworm

ENV PATH="/go/bin:/home/vscode/go/bin:${PATH}"

USER root

# Writable dirs for editor / package managers (must not be root-owned).
RUN mkdir -p \
      /home/vscode/.local/share/pnpm/store \
      /home/vscode/.local/bin \
      /home/vscode/.host-local-bin \
      /home/vscode/.cache \
  && chown -R vscode:vscode /home/vscode/.local /home/vscode/.host-local-bin /home/vscode/.cache

# Task (taskfile.dev) — install prebuilt binary (faster/more reliable than go install).
RUN sh -c "$(curl --location https://taskfile.dev/install.sh)" -- -d -b /usr/local/bin \
  && task --version

USER vscode

# pnpm: use a home-local store (never repo .pnpm-store — that can be root-owned
# from old Docker runs and causes "readonly database"). Pin pnpm 10.x — pnpm 11
# treats ignored dependency build scripts as hard errors and breaks `pnpm dev`.
RUN printf '%s\n' \
      'store-dir=/home/vscode/.local/share/pnpm/store' \
      'confirm-modules-purge=false' \
      > /home/vscode/.npmrc

# Shell branding: show "vizber" in the prompt while the OS user stays vscode.
# Appended after the image's oh-my-zsh setup so it wins over the default theme.
RUN printf '%s\n' \
      '' \
      '# vizber Dev Container branding (OS user remains vscode for tooling compatibility)' \
      'export DEVCONTAINER_NAME=vizber' \
      'if [ -n "$ZSH_VERSION" ]; then' \
      '  PROMPT="%F{magenta}vizber%f %F{blue}%~%f %# "' \
      'elif [ -n "$BASH_VERSION" ]; then' \
      '  PS1="\[\e[35m\]vizber\[\e[0m\] \[\e[34m\]\w\[\e[0m\] \$ "' \
      'fi' \
      >> /home/vscode/.zshrc \
  && printf '%s\n' \
      '' \
      'export DEVCONTAINER_NAME=vizber' \
      'PS1="\[\e[35m\]vizber\[\e[0m\] \[\e[34m\]\w\[\e[0m\] \$ "' \
      >> /home/vscode/.bashrc

# golangci-lint v2 (matches .golangci.yml version: "2") — prebuilt binary
RUN curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh \
  | sh -s -- -b "$(go env GOPATH)/bin" \
  && golangci-lint --version
