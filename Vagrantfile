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
    sudo add-apt-repository ppa:webupd8team/java
    sudo apt-get update

    sudo apt-get install zlib1g-dev -y
    sudo apt-get install ruby2.3 ruby2.3-dev -y

    echo debconf shared/accepted-oracle-license-v1-1 select true | sudo debconf-set-selections
    sudo apt-get install oracle-java8-installer -y

    curl -L http://dynamodb-local.s3-website-us-west-2.amazonaws.com/dynamodb_local_latest.tar.gz -o dynamodb_local_latest.tar.gz
    tar zxvf dynamodb_local_latest.tar.gz
    java -Djava.library.path=./DynamoDBLocal_lib -jar DynamoDBLocal.jar -sharedDb &

    sudo gem install bundle
    cd /app && bundle install
  SHELL
end
