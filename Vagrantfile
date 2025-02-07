# -*- mode: ruby -*-
# vi: set ft=ruby :

Vagrant.configure("2") do |config|
	config.vm.box = "debian/bookworm64"

	config.vm.synced_folder "cmd", "/home/vagrant/cmd"
	config.vm.provision "file", source: "go.mod", destination: "/home/vagrant/go.mod"
	config.vm.provision "file", source: "go.sum", destination: "/home/vagrant/go.sum"
	config.vm.provision "file", source: "test.sample.yaml", destination: "/home/vagrant/test.sample.yaml"

	# config.vm.network "private_network",
		# nic_type: "virtio"

	config.vm.provision "shell", inline: <<-SHELL
		apt-get update && apt-get upgrade -y
		apt-get install -y curl git

		cd /usr/local

		wget https://go.dev/dl/go1.23.5.linux-amd64.tar.gz
		tar -C /usr/local -xzf go1.23.5.linux-amd64.tar.gz

		export PATH=$PATH:/usr/local/go/bin
		echo "export PATH=$PATH:/usr/local/go/bin" >> /etc/profile
		go version
		SHELL
end
