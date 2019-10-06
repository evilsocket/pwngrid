all:
    go build -o pwngrid cmd/pwngrid/*.go

install:
    cp pwngrid /usr/local/bin/
    mkdir -p /etc/systemd/system/
    cp pwngrid.service /etc/systemd/system/
    chmod 644 /etc/systemd/system/pwngrid.service
    systemctl daemon-reload
    systemctl enable pwngrid.service

restart:
    service pwngrid restart