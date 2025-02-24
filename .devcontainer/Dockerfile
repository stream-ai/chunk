# ------------------------------------------------------
# Stage 1: Build and install Go + tools
# ------------------------------------------------------
  ARG GO_VERSION=1.24.0
  FROM golang:${GO_VERSION} AS build
  
  RUN  go install -v github.com/air-verse/air@latest \
    && go install -v github.com/aurc/loggo@latest \
    && go install -v mvdan.cc/gofumpt@latest \
    && go install -v golang.org/x/tools/cmd/goimports@latest \
    && go install -v golang.org/x/tools/gopls@latest
  
  # ------------------------------------------------------
  # Stage 2: Final devcontainer image
  # ------------------------------------------------------
  FROM mcr.microsoft.com/devcontainers/base:ubuntu
  
  ARG AWS_CLI_VERSION=2.23.5
  ARG AWS_CDK_VERSION=2.177.0
  ARG NODE_VERSION=22.13.1
  ARG PYTHON_VER=3.10
  
  ENV AWS_CLI_VERSION=${AWS_CLI_VERSION}
  ENV AWS_CDK_VERSION=${AWS_CDK_VERSION}
  ENV NODE_VERSION=${NODE_VERSION}
  ENV PYTHON_VER=${PYTHON_VER}
  
  # [NEW] Dependencies for Homebrew (Linuxbrew)
  RUN apt-get update \
    && apt-get install -y --no-install-recommends \
    ca-certificates \
    curl \
    gnupg \
    lsb-release \
    unzip \
    vim \
    inotify-tools \
    iputils-ping \
    socat \
    fswatch \
    expect \
    build-essential \
    file \
    git \
    && rm -rf /var/lib/apt/lists/*
  
  # ----------------------------
  # Install Docker CLI + Docker Compose plugin
  # ----------------------------
  RUN sudo curl -fsSL https://download.docker.com/linux/ubuntu/gpg \
    | gpg --dearmor --batch --yes -o /usr/share/keyrings/docker-archive-keyring.gpg \
    && echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] \
    https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" \
    > /etc/apt/sources.list.d/docker.list \
    && apt-get update \
    && apt-get install -y docker-ce-cli docker-compose-plugin
  
  # ----------------------------
  # Install AWS CLI v2
  # ----------------------------
  RUN cd /tmp \
    && curl -sSL "https://awscli.amazonaws.com/awscli-exe-linux-x86_64-${AWS_CLI_VERSION}.zip" -o awscliv2.zip \
    && unzip awscliv2.zip \
    && ./aws/install \
    && rm -rf ./aws* \
    && cd -
  
  # ----------------------------
  # Install Python
  # ----------------------------
  RUN apt-get update \
    && apt-get install -y python${PYTHON_VER} python${PYTHON_VER}-venv python${PYTHON_VER}-dev \
    && ln -sf /usr/bin/python${PYTHON_VER} /usr/local/bin/python \
    && rm -rf /var/lib/apt/lists/*
  
  # Add the 'vscode' user to the 'docker' group
  RUN groupadd docker \
    && usermod -aG docker vscode
  
  # ----------------------------
  # Install FNM and Node.js
  # ----------------------------
  RUN curl -fsSL https://fnm.vercel.app/install | bash -s -- --install-dir /usr/local/fnm/bin --skip-shell
  ENV FNM_DIR=/usr/local/fnm
  ENV PATH=/usr/local/fnm/bin:${PATH}
  RUN fnm install ${NODE_VERSION} && fnm default ${NODE_VERSION}
  ENV PATH=/usr/local/fnm/aliases/default/bin:${PATH}
  
  # ----------------------------
  # Install AWS CDK
  # ----------------------------
  RUN npm install -g aws-cdk@${AWS_CDK_VERSION} 
  
  # ------------------------------------------------------
  # Install Homebrew (Linuxbrew), jless
  # ------------------------------------------------------
  RUN /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)" \
    && echo 'eval "$(/home/linuxbrew/.linuxbrew/bin/brew shellenv)"' >> /etc/bash.bashrc \
    && eval "$(/home/linuxbrew/.linuxbrew/bin/brew shellenv)" \
    && brew install jless 
  
  # Make brew, jless, and fx available to the 'vscode' shell environment
  RUN echo 'eval "$(/home/linuxbrew/.linuxbrew/bin/brew shellenv)"' >> /home/vscode/.bashrc \
    && echo 'eval "$(/home/linuxbrew/.linuxbrew/bin/brew shellenv)"' >> /home/vscode/.zshrc
  
  # --------------------------------------------------------
  # Copy Go toolchain & installed binaries from build stage
  # --------------------------------------------------------
  COPY --from=build /usr/local/go/ /usr/local/go/
  COPY --from=build /go /go
  RUN chown -R vscode:vscode /go
  
  # Add Go to path
  ENV GOPATH=/go
  ENV PATH=/usr/local/go/bin:/go/bin:$PATH
  
  # ------------------------------------------------------
  # Setup Workspace
  # ------------------------------------------------------
  WORKDIR /workspace
  
  RUN ln -s /workspace/.scripts/alias.sh /home/vscode/alias.sh \
    && echo "source /home/vscode/alias.sh" >> /home/vscode/.bashrc \
    && echo "source /home/vscode/alias.sh" >> /home/vscode/.zshrc
  