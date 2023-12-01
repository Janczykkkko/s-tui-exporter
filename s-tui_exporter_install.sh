#!/bin/bash

# Check if running with sudo
if [[ $EUID -ne 0 ]]; then
    echo "This script must be run with sudo."
    exit 1
fi

# Function to check and install packages
check_install() {
    package="$1"
    if ! command -v "$package" &> /dev/null; then
        echo "$package is not installed. Installing..."
        if [[ $(command -v apt-get) ]]; then
            sudo apt-get install -y "$package"
        elif [[ $(command -v yum) ]]; then
            sudo yum install -y "$package"
        else
            echo "Package manager not found. Please install $package manually."
            exit 1
        fi
    else
        echo "$package is already installed."
    fi
}

# Check the Linux distribution
distro=$(awk -F= '/^NAME/{print $2}' /etc/os-release)
echo "Detected Linux distribution: $distro"

# Check and install required packages
check_install "git"
check_install "s-tui"

# Install Go
if ! command -v go &> /dev/null; then
    echo "Go is not installed. Installing..."
    wget https://go.dev/dl/go1.21.4.linux-amd64.tar.gz
    sudo tar -C /usr/local -xzf go1.21.4.linux-amd64.tar.gz
    rm go1.21.4.linux-amd64.tar.gz
    export PATH=$PATH:/usr/local/go/bin
    echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
    echo 'export PATH=$PATH:/usr/local/go/bin' | sudo tee -a /etc/profile.d/go.sh > /dev/null
    source ~/.bashrc
fi

# Clone the repository and build the binary
git clone https://github.com/Janczykkkko/s-tui-exporter.git
cd s-tui-exporter || exit
go build
chmod +x s-tui-exporter
sudo mv s-tui-exporter /usr/local/bin/

# Create systemd service
sudo tee /etc/systemd/system/s-tui-exporter.service > /dev/null <<EOF
[Unit]
Description=s-tui prometheus exporter service
After=network.target

[Service]
User=root
Group=root
ExecStart=/usr/local/bin/s-tui-exporter
Restart=always

[Install]
WantedBy=multi-user.target
EOF

# Reload systemd, enable, and start the service
sudo systemctl daemon-reload
sudo systemctl enable s-tui-exporter.service
sudo systemctl start s-tui-exporter.service

echo "Setup complete. Service 's-tui-exporter' is enabled and running."
