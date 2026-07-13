# Dev container toolchain for vizb ("vizber").
#
# The Microsoft Go image ships a non-root user named `vscode`. We rename it to
# `vizber` (same UID/GID 1000) so the OS login matches the project identity.
# Node is installed here (not via a Dev Container feature) so feature scripts
# cannot reintroduce a `vscode` user after the rename.
FROM mcr.microsoft.com/devcontainers/go:1.26-bookworm

USER root

# ── Rename vscode → vizber (UID/GID stay 1000 for host bind-mount ownership) ──
RUN set -eux; \
    groupmod --new-name vizber vscode; \
    usermod --login vizber --move-home --home /home/vizber vscode; \
    echo 'vizber ALL=(root) NOPASSWD:ALL' > /etc/sudoers.d/vizber; \
    chmod 0440 /etc/sudoers.d/vizber; \
    # Drop any leftover vscode identity
    if getent passwd vscode >/dev/null 2>&1; then userdel -r vscode || true; fi; \
    if getent group vscode >/dev/null 2>&1; then groupdel vscode || true; fi; \
    getent passwd vizber; \
    test ! -d /home/vscode; \
    test -d /home/vizber

# Tell Dev Container tooling the remote user is vizber (overrides image metadata).
LABEL devcontainer.metadata='[{"remoteUser":"vizber"}]'

# Base image pins GOTOOLCHAIN=local and ships 1.26.4; go.mod needs 1.26.5.
# auto = download the newer toolchain when the module requires it.
ENV HOME=/home/vizber \
    PATH="/go/bin:/home/vizber/go/bin:/usr/local/go/bin:${PATH}" \
    GOTOOLCHAIN=auto
# Writable dirs for editor / package managers
RUN mkdir -p \
      /home/vizber/.local/share/pnpm/store \
      /home/vizber/.local/bin \
      /home/vizber/.host-local-bin \
      /home/vizber/.cache \
      /home/vizber/go \
  && chown -R vizber:vizber /home/vizber

# Node 22 + corepack (no node feature — keeps user rename intact)
RUN curl -fsSL https://deb.nodesource.com/setup_22.x | bash - \
  && apt-get install -y --no-install-recommends nodejs \
  && corepack enable \
  && corepack prepare pnpm@10.33.2 --activate \
  && node -v && pnpm -v \
  && rm -rf /var/lib/apt/lists/*

# Task (taskfile.dev)
RUN sh -c "$(curl --location https://taskfile.dev/install.sh)" -- -d -b /usr/local/bin \
  && task --version

USER vizber
WORKDIR /home/vizber

# pnpm: home-local store only (repo .pnpm-store can be root-owned from old Docker runs)
RUN printf '%s\n' \
      'store-dir=/home/vizber/.local/share/pnpm/store' \
      'confirm-modules-purge=false' \
      > /home/vizber/.npmrc

# Pre-fetch go1.26.5 (go.mod) so first builds do not wait on download.
RUN GOTOOLCHAIN=go1.26.5 go version

# golangci-lint v2 (matches .golangci.yml version: "2")
RUN curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh \
  | sh -s -- -b "$(go env GOPATH)/bin" \
  && golangci-lint --version
# Prompt: basename only via zsh/bash built-ins (%1~ / \W). No custom precmd.
# Appended after the image oh-my-zsh setup so it wins over the default theme.
RUN printf '%s\n' \
      '' \
      '# vizber: %1~ = last path segment only (project name); full path is still `pwd`' \
      "PROMPT='%F{magenta}vizber%f %F{cyan}%1~%f %# '" \
      >> /home/vizber/.zshrc \
  && printf '%s\n' \
      '' \
      '# vizber: \W = basename only' \
      'PS1="\[\e[35m\]vizber\[\e[0m\] \[\e[36m\]\W\[\e[0m\] \$ "' \
      >> /home/vizber/.bashrc
