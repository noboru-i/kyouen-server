# -*- mode: ruby -*-
# vi: set ft=ruby :

Vagrant.configure(2) do |config|
  config.vm.box = "ubuntu/trusty64"

  config.vm.network :private_network, ip: "192.168.50.18"
  config.vm.synced_folder ".", "/app", type: "rsync"

  config.vm.network "forwarded_port", guest: 80, host: 8080

  config.vm.provider :virtualbox do |vb|
    vb.customize ["modifyvm", :id, "--memory", 2048]
    vb.customize ["modifyvm", :id, "--cpus", 2]
  end

  config.vm.provision "shell", privileged: false, inline: <<-SHELL
    sudo apt-add-repository ppa:brightbox/ruby-ng
    sudo apt-get update
    sudo apt-get install zlib1g-dev -y
    sudo apt-get install ruby2.3 ruby2.3-dev -y
    sudo gem install bundle
    cd /app && bundle install
  SHELL
end
