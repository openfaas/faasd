sudo apt update
sudo apt -y dist-upgrade
sudo apt-get install --reinstall linux-firmware
git config --global user.name "AM"
git config --global user.email "aditya.account@outlook.com"
sudo apt install -y openjdk-8-jdk
sudo apt install -y nodejs npm zip
sudo apt install -y python3-pip
sudo sysctl -w fs.file-max=262144
sudo sysctl net.core.somaxconn=1024
sudo sysctl net.core.netdev_max_backlog=2000
sudo sysctl net.ipv4.tcp_max_syn_backlog=2048
sudo apt install golang
pip3 install psutil
sudo cat /var/lib/faasd/secrets/basic-auth-password > password.txt
faas-cli login --password-stdin < password.txt
faas-cli store deploy cows
curl -d "Hello from faasd" http://127.0.0.1:8080/function/cows