#!/bin/sh

init_config() {
    jq_cmd='.'

    if [ -n "$RCON_PASS" ]; then
      jq_cmd="${jq_cmd} | .rcon_pass = \"$RCON_PASS\""
      echo "Factorio rcon password is '$RCON_PASS'"
    fi

    jq_cmd="${jq_cmd} | .sq_lite_database_file = \"/opt/fsm-data/sqlite.db\""
    jq_cmd="${jq_cmd} | .log_file = \"/opt/fsm-data/factorio-server-manager.log\""

    jq "${jq_cmd}" /opt/fsm/conf.json >/opt/fsm-data/conf.json
}

random_pass() {
    LC_ALL=C tr -dc 'a-zA-Z0-9' </dev/urandom | fold -w 24 | head -n 1
}

install_game() {
    echo "Downloading Factorio ${FACTORIO_VERSION}..."
    if ! curl --fail --location "https://www.factorio.com/get-download/${FACTORIO_VERSION}/headless/linux64" \
         --output /tmp/factorio_${FACTORIO_VERSION}.tar.xz; then
        echo "Failed to download Factorio"
        return 1
    fi

    echo "Extracting Factorio..."
    if ! tar -xf /tmp/factorio_${FACTORIO_VERSION}.tar.xz -C /opt; then
        echo "Failed to extract Factorio"
        rm -f /tmp/factorio_${FACTORIO_VERSION}.tar.xz
        return 1
    fi

    rm -f /tmp/factorio_${FACTORIO_VERSION}.tar.xz
    echo "Factorio installed successfully"
    return 0
}

if [ ! -f /opt/fsm-data/conf.json ]; then
    init_config
fi

# Only download if Factorio binary doesn't exist or isn't executable
if [ ! -f /opt/factorio/bin/x64/factorio ] || [ ! -x /opt/factorio/bin/x64/factorio ]; then
    echo "Factorio not found or not executable, downloading..."
    if ! install_game; then
        echo "Failed to install Factorio, exiting"
        exit 1
    fi
else
    echo "Factorio already installed, skipping download"
fi

cd /opt/fsm && ./factorio-server-manager --conf /opt/fsm-data/conf.json --dir /opt/factorio --port 80

