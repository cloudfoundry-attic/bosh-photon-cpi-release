require "erb"

Vagrant.configure("2") do |config|
  config.vm.box = "ubuntu/trusty64"

  config.vm.network "forwarded_port", guest: 25555, host: 25555 # BOSH Director API

  config.vm.provider "virtualbox" do |v|
    v.memory = 4096
    v.cpus = 2
  end

  config.vm.provision "shell", inline: "apt-get -y install linux-image-extra-$(uname -r)" # aufs

  config.vm.provision "bosh" do |c|
    c.manifest = ERB.new(File.read("manifests/photon-bosh.yml")).result
  end
end
