# Dev container toolchain for vizb.
# Node 22 is added via the node feature in devcontainer.json.
#
# Keep the base image user `vscode` (UID 1000). Renaming it breaks Zed/VS Code
# feature installs and `docker exec -u …` (passwd entry missing). The cool
# project identity is the container name in devcontainer.json, not the OS user.
FROM mcr.microsoft.com/devcontainers/go:1.26-bookworm

ENV PATH="/go/bin:/home/vscode/go/bin:${PATH}"

USER root

# Ensure vscode owns ~/.local so the remote editor (Zed/VS Code) can write
# ~/.local/share. Do NOT bind-mount over ~/.local/bin — that makes Docker create
# ~/.local as root and breaks the remote server (exit status 1).
RUN mkdir -p /home/vscode/.local/share /home/vscode/.local/bin /home/vscode/.host-local-bin \
  && chown -R vscode:vscode /home/vscode/.local /home/vscode/.host-local-bin

# Task (taskfile.dev) — install prebuilt binary (faster/more reliable than go install).
RUN sh -c "$(curl --location https://taskfile.dev/install.sh)" -- -d -b /usr/local/bin \
  && task --version

# golangci-lint v2 (matches .golangci.yml version: "2") — prebuilt binary
USER vscode
RUN curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh \
  | sh -s -- -b "$(go env GOPATH)/bin" \
  && golangci-lint --version
